package llm

import (
	"context"
)

// LLM represents a large language model interface
type LLM interface {
	// Chat generates a response based on the conversation history
	Chat(ctx context.Context, messages []Message, opts ...Option) (*Message, error)

	// ChatStream streams the response tokens
	ChatStream(ctx context.Context, messages []Message, opts ...Option) (<-chan StreamResponse, error)

	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string, opts ...Option) (string, error)
}

// StreamResponse represents a streaming response
type StreamResponse struct {
	Message Message
	Error   error
	Done    bool
}

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleFunction  = "function"
)
