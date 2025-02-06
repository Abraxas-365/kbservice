package inmemory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Abraxas-365/kbservice/chathistory"
	"github.com/Abraxas-365/kbservice/llm"
)

// InMemoryRepository implements ChatHistoryRepository using in-memory storage
type InMemoryRepository struct {
	conversations map[string]chathistory.Conversation
	mu            sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		conversations: make(map[string]chathistory.Conversation),
	}
}

func (r *InMemoryRepository) AddMessage(ctx context.Context, conversationID string, message llm.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	conv.Messages = append(conv.Messages, message)
	conv.UpdatedAt = time.Now()
	r.conversations[conversationID] = conv

	return nil
}

func (r *InMemoryRepository) GetMessages(ctx context.Context, conversationID string, limit int) ([]llm.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	if limit <= 0 || limit > len(conv.Messages) {
		limit = len(conv.Messages)
	}

	start := len(conv.Messages) - limit
	if start < 0 {
		start = 0
	}

	return conv.Messages[start:], nil
}

func (r *InMemoryRepository) GetMessagesByFilter(ctx context.Context, conversationID string, filter chathistory.Filter, limit int) ([]llm.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	var filtered []llm.Message
	for _, msg := range conv.Messages {
		if r.messageMatchesFilter(msg, filter) {
			filtered = append(filtered, msg)
		}
	}

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}

	start := len(filtered) - limit
	if start < 0 {
		start = 0
	}

	return filtered[start:], nil
}

func (r *InMemoryRepository) DeleteMessages(ctx context.Context, conversationID string, filter chathistory.Filter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	var remaining []llm.Message
	for _, msg := range conv.Messages {
		if !r.messageMatchesFilter(msg, filter) {
			remaining = append(remaining, msg)
		}
	}

	conv.Messages = remaining
	conv.UpdatedAt = time.Now()
	r.conversations[conversationID] = conv

	return nil
}

func (r *InMemoryRepository) ClearHistory(ctx context.Context, conversationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	conv.Messages = []llm.Message{}
	conv.UpdatedAt = time.Now()
	r.conversations[conversationID] = conv

	return nil
}

func (r *InMemoryRepository) DeleteConversation(ctx context.Context, conversationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.conversations[conversationID]; !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	delete(r.conversations, conversationID)
	return nil
}

func (r *InMemoryRepository) CreateConversation(ctx context.Context, conv chathistory.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.conversations[conv.ID]; exists {
		return fmt.Errorf("conversation already exists: %s", conv.ID)
	}

	r.conversations[conv.ID] = conv
	return nil
}

func (r *InMemoryRepository) GetConversation(ctx context.Context, conversationID string) (*chathistory.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	return &conv, nil
}

func (r *InMemoryRepository) ListConversations(ctx context.Context, filter chathistory.Filter, limit, offset int) ([]chathistory.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var conversations []chathistory.Conversation
	for _, conv := range r.conversations {
		if r.conversationMatchesFilter(conv, filter) {
			conversations = append(conversations, conv)
		}
	}

	// Sort by UpdatedAt descending
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].UpdatedAt.After(conversations[j].UpdatedAt)
	})

	if offset >= len(conversations) {
		return []chathistory.Conversation{}, nil
	}

	end := offset + limit
	if end > len(conversations) {
		end = len(conversations)
	}

	return conversations[offset:end], nil
}

func (r *InMemoryRepository) UpdateConversationMetadata(ctx context.Context, conversationID string, metadata map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	conv.Metadata = metadata
	conv.UpdatedAt = time.Now()
	r.conversations[conversationID] = conv

	return nil
}

func (r *InMemoryRepository) GetMessageCount(ctx context.Context, conversationID string, filter chathistory.Filter) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conv, exists := r.conversations[conversationID]
	if !exists {
		return 0, fmt.Errorf("conversation not found: %s", conversationID)
	}

	if filter.IsEmpty() {
		return len(conv.Messages), nil
	}

	count := 0
	for _, msg := range conv.Messages {
		if r.messageMatchesFilter(msg, filter) {
			count++
		}
	}

	return count, nil
}

func (r *InMemoryRepository) messageMatchesFilter(msg llm.Message, filter chathistory.Filter) bool {
	if filter.StartTime != nil && msg.Metadata != nil {
		if timestamp, ok := msg.Metadata["timestamp"].(time.Time); ok {
			if timestamp.Before(*filter.StartTime) {
				return false
			}
		}
	}

	if filter.EndTime != nil && msg.Metadata != nil {
		if timestamp, ok := msg.Metadata["timestamp"].(time.Time); ok {
			if timestamp.After(*filter.EndTime) {
				return false
			}
		}
	}

	if len(filter.Roles) > 0 {
		roleMatch := false
		for _, role := range filter.Roles {
			if msg.Role == role {
				roleMatch = true
				break
			}
		}
		if !roleMatch {
			return false
		}
	}

	if filter.Search != "" {
		if !strings.Contains(strings.ToLower(msg.Content), strings.ToLower(filter.Search)) {
			return false
		}
	}

	return true
}

func (r *InMemoryRepository) conversationMatchesFilter(conv chathistory.Conversation, filter chathistory.Filter) bool {
	if filter.StartTime != nil && conv.CreatedAt.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && conv.CreatedAt.After(*filter.EndTime) {
		return false
	}

	if filter.Metadata != nil {
		for k, v := range filter.Metadata {
			if convValue, exists := conv.Metadata[k]; !exists || convValue != v {
				return false
			}
		}
	}

	return true
}
