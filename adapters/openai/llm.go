package openai

import (
	"context"
	"errors"
	"io"

	"github.com/Abraxas-365/kbservice/llm"
	"github.com/sashabaranov/go-openai"
)

type OpenAILLM struct {
	client *openai.Client
	model  string
}

func NewOpenAILLM(apiKey string, model string) *OpenAILLM {
	if model == "" {
		model = openai.GPT4TurboPreview
	}
	return &OpenAILLM{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (o *OpenAILLM) Chat(ctx context.Context, messages []llm.Message, opts ...llm.Option) (*llm.Message, error) {
	options := &llm.ChatOptions{
		Temperature: 0.1,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Convert messages to OpenAI format
	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	// Create request
	req := openai.ChatCompletionRequest{
		Model:            o.model,
		Messages:         openAIMessages,
		Temperature:      float32(options.Temperature),
		TopP:             float32(options.TopP),
		MaxTokens:        options.MaxTokens,
		Stop:             options.Stop,
		PresencePenalty:  float32(options.PresencePenalty),
		FrequencyPenalty: float32(options.FrequencyPenalty),
	}

	// Add tools if functions are provided
	if len(options.Functions) > 0 {
		tools := make([]openai.Tool, len(options.Functions))
		for i, f := range options.Functions {
			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        f.Name,
					Description: f.Description,
					Parameters:  f.Parameters,
				},
			}
		}
		req.Tools = tools

		// Set tool choice
		if options.FunctionCall != "" {
			req.ToolChoice = &openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: options.FunctionCall,
				},
			}
		} else {
			req.ToolChoice = "auto"
		}
	}

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, handleOpenAIError("Chat", err)
	}

	if len(resp.Choices) == 0 {
		return nil, &llm.LLMError{
			Op:      "Chat",
			Message: "no response choices returned",
		}
	}

	// Convert response to Message
	message := &llm.Message{
		Role:    resp.Choices[0].Message.Role,
		Content: resp.Choices[0].Message.Content,
		Name:    resp.Choices[0].Message.Name,
	}

	// Handle tool calls in response
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCall := resp.Choices[0].Message.ToolCalls[0]
		message.FuncCall = &llm.FunctionCall{
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		}
	}

	return message, nil
}

func (o *OpenAILLM) ChatStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (<-chan llm.StreamResponse, error) {
	options := &llm.ChatOptions{
		Temperature: 0.7,
	}
	for _, opt := range opts {
		opt(options)
	}

	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:            o.model,
		Messages:         openAIMessages,
		Temperature:      float32(options.Temperature),
		TopP:             float32(options.TopP),
		MaxTokens:        options.MaxTokens,
		Stop:             options.Stop,
		Stream:           true,
		PresencePenalty:  float32(options.PresencePenalty),
		FrequencyPenalty: float32(options.FrequencyPenalty),
	}

	// Add tools if functions are provided
	if len(options.Functions) > 0 {
		tools := make([]openai.Tool, len(options.Functions))
		for i, f := range options.Functions {
			tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        f.Name,
					Description: f.Description,
					Parameters:  f.Parameters,
				},
			}
		}
		req.Tools = tools

		if options.FunctionCall != "" {
			req.ToolChoice = &openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: options.FunctionCall,
				},
			}
		} else {
			req.ToolChoice = "auto"
		}
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, handleOpenAIError("ChatStream", err)
	}

	responseChan := make(chan llm.StreamResponse)

	go func() {
		defer close(responseChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				responseChan <- llm.StreamResponse{
					Done: true,
				}
				return
			}
			if err != nil {
				responseChan <- llm.StreamResponse{
					Error: handleOpenAIError("ChatStream", err),
					Done:  true,
				}
				return
			}

			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				if choice.Delta.Content != "" || choice.Delta.Role != "" {
					responseChan <- llm.StreamResponse{
						Message: llm.Message{
							Role:    choice.Delta.Role,
							Content: choice.Delta.Content,
						},
						Done: false,
					}
				}

				// Handle tool calls in streaming response
				if len(choice.Delta.ToolCalls) > 0 {
					toolCall := choice.Delta.ToolCalls[0]
					responseChan <- llm.StreamResponse{
						Message: llm.Message{
							Role: choice.Delta.Role,
							FuncCall: &llm.FunctionCall{
								Name:      toolCall.Function.Name,
								Arguments: toolCall.Function.Arguments,
							},
						},
						Done: false,
					}
				}

				if choice.FinishReason == "stop" {
					responseChan <- llm.StreamResponse{
						Done: true,
					}
					return
				}
			}
		}
	}()

	return responseChan, nil
}

func (o *OpenAILLM) Complete(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	messages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: prompt,
		},
	}

	resp, err := o.Chat(ctx, messages, opts...)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

func handleOpenAIError(op string, err error) error {
	if err == nil {
		return nil
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 400:
			return &llm.LLMError{
				Op:      op,
				Message: "invalid request",
				Err:     err,
			}
		case 401:
			return &llm.LLMError{
				Op:      op,
				Message: "invalid API key",
				Err:     err,
			}
		case 429:
			return &llm.LLMError{
				Op:      op,
				Message: "rate limit exceeded",
				Err:     err,
			}
		case 500:
			return &llm.LLMError{
				Op:      op,
				Message: "OpenAI server error",
				Err:     err,
			}
		}
	}

	return &llm.LLMError{
		Op:      op,
		Message: "unexpected error",
		Err:     err,
	}
}
