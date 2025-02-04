package llm

// ChatOptions represents options for chat completion
type ChatOptions struct {
	Temperature      float32    // Controls randomness (0.0 to 2.0)
	TopP             float32    // Controls diversity (0.0 to 1.0)
	MaxTokens        int        // Maximum number of tokens to generate
	Stop             []string   // Stop sequences
	Functions        []Function // Available functions
	FunctionCall     string     // Force specific function call
	PresencePenalty  float32    // Penalty for new tokens based on presence in text
	FrequencyPenalty float32    // Penalty for new tokens based on frequency in text
	Stream           bool       // Whether to stream the response
}

// Option is a function type to modify ChatOptions
type Option func(*ChatOptions)

// Common option functions
func WithTemperature(temp float32) Option {
	return func(o *ChatOptions) {
		o.Temperature = temp
	}
}

func WithTopP(topP float32) Option {
	return func(o *ChatOptions) {
		o.TopP = topP
	}
}

func WithMaxTokens(tokens int) Option {
	return func(o *ChatOptions) {
		o.MaxTokens = tokens
	}
}

func WithStop(stop []string) Option {
	return func(o *ChatOptions) {
		o.Stop = stop
	}
}

func WithFunctions(functions []Function) Option {
	return func(o *ChatOptions) {
		o.Functions = functions
	}
}

func WithFunctionCall(functionCall string) Option {
	return func(o *ChatOptions) {
		o.FunctionCall = functionCall
	}
}

func WithStream(stream bool) Option {
	return func(o *ChatOptions) {
		o.Stream = stream
	}
}
