package chathistory

import (
	"context"
	"time"

	"github.com/Abraxas-365/kbservice/llm"
)

type Memory struct {
	repo ChatHistoryRepository
	opts *Options
}

func New(repo ChatHistoryRepository, opts ...Option) *Memory {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &Memory{
		repo: repo,
		opts: options,
	}
}

// CreateConversation creates a new conversation
func (m *Memory) CreateConversation(ctx context.Context, metadata map[string]any) (*Conversation, error) {
	conv := Conversation{
		ID:        m.opts.GenerateID(),
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := m.repo.CreateConversation(ctx, conv)
	if err != nil {
		return nil, err
	}

	return &conv, nil
}

func (m *Memory) CreateConversationWithID(ctx context.Context, metadata map[string]any, id string) (*Conversation, error) {
	conv := Conversation{
		ID:        id,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := m.repo.CreateConversation(ctx, conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// AddMessage adds a message to a specific conversation
func (m *Memory) AddMessage(ctx context.Context, conversationID string, msg llm.Message) error {
	return m.repo.AddMessage(ctx, conversationID, msg)
}

// GetMessages retrieves messages from a specific conversation
func (m *Memory) GetMessages(ctx context.Context, conversationID string, limit int) ([]llm.Message, error) {
	if limit <= 0 {
		limit = m.opts.ReturnLimit
	}
	return m.repo.GetMessages(ctx, conversationID, limit)
}

// GetConversation retrieves a conversation by ID
func (m *Memory) GetConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	return m.repo.GetConversation(ctx, conversationID)
}

// ListConversations retrieves all conversations with optional filters
func (m *Memory) ListConversations(ctx context.Context, filter Filter, limit, offset int) ([]Conversation, error) {
	return m.repo.ListConversations(ctx, filter, limit, offset)
}

// DeleteConversation deletes an entire conversation
func (m *Memory) DeleteConversation(ctx context.Context, conversationID string) error {
	return m.repo.DeleteConversation(ctx, conversationID)
}

// UpdateConversationMetadata updates conversation metadata
func (m *Memory) UpdateConversationMetadata(ctx context.Context, conversationID string, metadata map[string]any) error {
	return m.repo.UpdateConversationMetadata(ctx, conversationID, metadata)
}

// GetMessagesByFilter retrieves messages using filter from a specific conversation
func (m *Memory) GetMessagesByFilter(ctx context.Context, conversationID string, filter Filter) ([]llm.Message, error) {
	return m.repo.GetMessagesByFilter(ctx, conversationID, filter, m.opts.ReturnLimit)
}

// ClearHistory clears all messages from a specific conversation
func (m *Memory) ClearHistory(ctx context.Context, conversationID string) error {
	return m.repo.ClearHistory(ctx, conversationID)
}

func (m *Memory) GetMessageCount(ctx context.Context, conversationID string, filter Filter) (int, error) {
	return m.repo.GetMessageCount(ctx, conversationID, filter)
}

func (m *Memory) GetID() string {
	return m.opts.GenerateID()
}
