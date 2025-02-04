package openai

import (
	"context"
	"errors"

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
		Temperature: 0.7,
		MaxTokens:   0, // 0 means no limit
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

	// Add functions if provided
	if len(options.Functions) > 0 {
		req.Functions = make([]openai.FunctionDefinition, len(options.Functions))
		for i, f := range options.Functions {
			req.Functions[i] = openai.FunctionDefinition{
				Name:        f.Name,
				Description: f.Description,
				Parameters:  f.Parameters,
			}
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

	return &llm.Message{
		Role:    resp.Choices[0].Message.Role,
		Content: resp.Choices[0].Message.Content,
		Name:    resp.Choices[0].Message.Name,
	}, nil
}

func (o *OpenAILLM) ChatStream(ctx context.Context, messages []llm.Message, opts ...llm.Option) (<-chan llm.StreamResponse, error) {
	options := &llm.ChatOptions{
		Temperature: 0.7,
		Stream:      true,
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
			if err != nil {
				responseChan <- llm.StreamResponse{
					Error: handleOpenAIError("ChatStream", err),
					Done:  true,
				}
				return
			}

			if len(response.Choices) > 0 {
				responseChan <- llm.StreamResponse{
					Message: llm.Message{
						Role:    response.Choices[0].Delta.Role,
						Content: response.Choices[0].Delta.Content,
					},
					Done: false,
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
