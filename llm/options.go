package llm

import "encoding/json"

// ResponseFormatType represents the type of response format
type ResponseFormatType string

const (
	JSONObject ResponseFormatType = "json_object"
	JSONSchema ResponseFormatType = "json_schema"
)

// ResponseFormat represents the format specification for model output
type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JSONSchema interface{}        `json:"schema,omitempty"` // Optional JSON schema
}

// Add to ChatOptions struct:
type ChatOptions struct {
	Temperature      float32         // Controls randomness (0.0 to 2.0)
	TopP             float32         // Controls diversity (0.0 to 1.0)
	MaxTokens        int             // Maximum number of tokens to generate
	Stop             []string        // Stop sequences
	Functions        []Function      // Available functions
	FunctionCall     string          // Force specific function call
	PresencePenalty  float32         // Penalty for new tokens based on presence in text
	FrequencyPenalty float32         // Penalty for new tokens based on frequency in text
	Stream           bool            // Whether to stream the response
	ResponseFormat   *ResponseFormat // Response format specification
}

// Option is a function type to modify ChatOptions
type Option func(*ChatOptions)

// JSONSchemaMarshaler is a helper type that implements json.Marshaler
type JSONSchemaMarshaler struct {
	Schema interface{}
}

func (j JSONSchemaMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Schema)
}

// WithJSONObjectFormat sets the response format to JSON object mode
func WithJSONObjectFormat() Option {
	return func(o *ChatOptions) {
		o.ResponseFormat = &ResponseFormat{
			Type: JSONObject,
		}
	}
}

// WithJSONSchemaFormat sets the response format to JSON schema mode with a specific schema
func WithJSONSchemaFormat(schema interface{}) Option {
	return func(o *ChatOptions) {
		o.ResponseFormat = &ResponseFormat{
			Type:       JSONSchema,
			JSONSchema: schema,
		}
	}
}

// WithResponseFormat sets a custom response format
func WithResponseFormat(format *ResponseFormat) Option {
	return func(o *ChatOptions) {
		o.ResponseFormat = format
	}
}

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
