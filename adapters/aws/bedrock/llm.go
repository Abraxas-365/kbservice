package bedrock

import (
	"context"
	"encoding/json"

	"github.com/Abraxas-365/kbservice/llm"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/smithy-go/ptr"
)

// LLMModelID represents available Bedrock models
type LLMModelID string

const (
	Claude2         LLMModelID = "anthropic.claude-v2"
	Claude2Instant  LLMModelID = "anthropic.claude-instant-v1"
	Claude3         LLMModelID = "anthropic.claude-3-sonnet-20240229-v1:0"
	Titan           LLMModelID = "amazon.titan-text-express-v1"
	LLama2_70B      LLMModelID = "meta.llama2-70b-v1"
	LLama2_13B      LLMModelID = "meta.llama2-13b-v1"
	LLama2_70B_Chat LLMModelID = "meta.llama2-70b-chat-v1"
	LLama2_13B_Chat LLMModelID = "meta.llama2-13b-chat-v1"
)

type BedrockLLM struct {
	client *bedrockruntime.Client
	model  LLMModelID
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Messages         []anthropicMessage `json:"messages"`
	MaxTokens        int                `json:"max_tokens"`
	Temperature      float32            `json:"temperature,omitempty"`
	TopP             float32            `json:"top_p,omitempty"`
	TopK             int                `json:"top_k,omitempty"`
	StopSequences    []string           `json:"stop_sequences,omitempty"`
	AnthropicVersion string             `json:"anthropic_version"`
	Stream           bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	Type       string `json:"type,omitempty"`
	Content    string `json:"content,omitempty"`
	Completion string `json:"completion,omitempty"` // for backwards compatibility
	StopReason string `json:"stop_reason,omitempty"`
	Model      string `json:"model,omitempty"`
}

func NewBedrockLLM(client *bedrockruntime.Client, model LLMModelID) *BedrockLLM {
	if model == "" {
		model = Claude2
	}
	return &BedrockLLM{
		client: client,
		model:  model,
	}
}

func convertToAnthropicMessages(messages []llm.Message) []anthropicMessage {
	anthropicMsgs := make([]anthropicMessage, len(messages))
	for i, msg := range messages {
		role := msg.Role
		if role == llm.RoleFunction {
			role = "assistant"
		}
		anthropicMsgs[i] = anthropicMessage{
			Role:    role,
			Content: msg.Content,
		}
	}
	return anthropicMsgs
}

func (b *BedrockLLM) Chat(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Message, error) {
	options := &llm.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
	}
	for _, opt := range opts {
		opt(options)
	}

	var requestBody []byte
	var err error

	switch b.model {
	case Claude2, Claude2Instant, Claude3:
		anthropicReq := anthropicRequest{
			Messages:         convertToAnthropicMessages(messages),
			MaxTokens:        options.MaxTokens,
			Temperature:      options.Temperature,
			TopP:             options.TopP,
			StopSequences:    options.Stop,
			AnthropicVersion: "bedrock-2023-05-31",
		}
		requestBody, err = json.Marshal(anthropicReq)
		if err != nil {
			return nil, &llm.LLMError{
				Op:      "Chat",
				Message: "failed to marshal request",
				Err:     err,
			}
		}
	default:
		return nil, &llm.LLMError{
			Op:      "Chat",
			Message: "unsupported model",
		}
	}

	output, err := b.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     ptr.String(string(b.model)),
		Body:        requestBody,
		ContentType: ptr.String("application/json"),
	})
	if err != nil {
		return nil, handleBedrockError("Chat", err)
	}

	var resp anthropicResponse
	if err := json.Unmarshal(output.Body, &resp); err != nil {
		return nil, &llm.LLMError{
			Op:      "Chat",
			Message: "failed to unmarshal response",
			Err:     err,
		}
	}

	content := resp.Content
	if content == "" {
		content = resp.Completion // fallback for older API versions
	}

	return &llm.Message{
		Role:    llm.RoleAssistant,
		Content: content,
	}, nil
}

func (b *BedrockLLM) ChatStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (<-chan llm.StreamResponse, error) {
	options := &llm.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      true,
	}
	for _, opt := range opts {
		opt(options)
	}

	responseChan := make(chan llm.StreamResponse)

	var requestBody []byte
	var err error

	switch b.model {
	case Claude2, Claude2Instant, Claude3:
		anthropicReq := anthropicRequest{
			Messages:         convertToAnthropicMessages(messages),
			MaxTokens:        options.MaxTokens,
			Temperature:      options.Temperature,
			TopP:             options.TopP,
			StopSequences:    options.Stop,
			AnthropicVersion: "bedrock-2023-05-31",
			Stream:           true,
		}
		requestBody, err = json.Marshal(anthropicReq)
		if err != nil {
			return nil, &llm.LLMError{
				Op:      "ChatStream",
				Message: "failed to marshal request",
				Err:     err,
			}
		}
	default:
		return nil, &llm.LLMError{
			Op:      "ChatStream",
			Message: "unsupported model",
		}
	}

	output, err := b.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     ptr.String(string(b.model)),
		Body:        requestBody,
		ContentType: ptr.String("application/json"),
	})
	if err != nil {
		return nil, handleBedrockError("ChatStream", err)
	}

	go func() {
		defer close(responseChan)

		stream := output.GetStream()
		defer stream.Close()

		for event := range stream.Events() {
			select {
			case <-ctx.Done():
				responseChan <- llm.StreamResponse{
					Error: &llm.LLMError{
						Op:      "ChatStream",
						Message: "context cancelled",
						Err:     ctx.Err(),
					},
					Done: true,
				}
				return
			default:
				if chunk, ok := event.(*types.ResponseStreamMemberChunk); ok {
					var resp anthropicResponse
					if err := json.Unmarshal(chunk.Value.Bytes, &resp); err != nil {
						responseChan <- llm.StreamResponse{
							Error: &llm.LLMError{
								Op:      "ChatStream",
								Message: "failed to unmarshal chunk",
								Err:     err,
							},
							Done: true,
						}
						return
					}

					content := resp.Content
					if content == "" {
						content = resp.Completion // fallback for older API versions
					}

					responseChan <- llm.StreamResponse{
						Message: llm.Message{
							Role:    llm.RoleAssistant,
							Content: content,
						},
						Done: false,
					}

					if resp.StopReason != "" {
						responseChan <- llm.StreamResponse{Done: true}
						return
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			responseChan <- llm.StreamResponse{
				Error: &llm.LLMError{
					Op:      "ChatStream",
					Message: "stream error",
					Err:     err,
				},
				Done: true,
			}
			return
		}
	}()

	return responseChan, nil
}

func (b *BedrockLLM) Complete(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	messages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: prompt,
		},
	}

	resp, err := b.Chat(ctx, messages, opts...)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

func handleBedrockError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &llm.LLMError{
		Op:      op,
		Message: "Bedrock API error",
		Err:     err,
	}
}
