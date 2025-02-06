package chathistory

import (
	"context"
	"time"

	"github.com/Abraxas-365/kbservice/llm"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        string         `json:"id"`
	Messages  []llm.Message  `json:"messages"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Filter represents query filters for chat history
type Filter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Roles     []string
	Search    string
	Metadata  map[string]any
}

func (f Filter) IsEmpty() bool {
	return f.StartTime == nil &&
		f.EndTime == nil &&
		len(f.Roles) == 0 &&
		f.Search == "" &&
		len(f.Metadata) == 0
}

// ChatHistoryRepository interface defines methods for chat history operations
type ChatHistoryRepository interface {
	// AddMessage adds a new message to a specific conversation
	AddMessage(ctx context.Context, conversationID string, message llm.Message) error

	// GetMessages retrieves messages from a specific conversation
	GetMessages(ctx context.Context, conversationID string, limit int) ([]llm.Message, error)

	// GetMessagesByFilter retrieves messages using provided filters
	GetMessagesByFilter(ctx context.Context, conversationID string, filter Filter, limit int) ([]llm.Message, error)

	// DeleteMessages deletes messages that match the filter from a conversation
	DeleteMessages(ctx context.Context, conversationID string, filter Filter) error

	// ClearHistory deletes all messages from a conversation
	ClearHistory(ctx context.Context, conversationID string) error

	// DeleteConversation deletes an entire conversation
	DeleteConversation(ctx context.Context, conversationID string) error

	// CreateConversation creates a new conversation
	CreateConversation(ctx context.Context, conv Conversation) error

	// GetConversation retrieves a conversation by ID
	GetConversation(ctx context.Context, conversationID string) (*Conversation, error)

	// ListConversations retrieves all conversations with optional filters
	ListConversations(ctx context.Context, filter Filter, limit, offset int) ([]Conversation, error)

	// UpdateConversationMetadata updates conversation metadata
	UpdateConversationMetadata(ctx context.Context, conversationID string, metadata map[string]any) error

	// GetMessageCount returns the total number of messages in a conversation
	GetMessageCount(ctx context.Context, conversationID string, filter Filter) (int, error)
}
