package llm

import "strings"

const (
	// SystemRole represents a system message
	SystemRole = "system"
	// UserRole represents a user message
	UserRole = "user"
	// AssistantRole represents an assistant message
	AssistantRole = "assistant"
	// FunctionRole represents a function message
	FunctionRole = "function"
)

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Message represents a chat message
type Message struct {
	Role     string                 `json:"role"`    // e.g., "system", "user", "assistant"
	Content  string                 `json:"content"` // The message content
	Name     string                 `json:"name,omitempty"`
	FuncCall *FunctionCall          `json:"function_call,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetUsage returns the usage statistics from the message metadata
func (m *Message) GetUsage() *Usage {
	if m.Metadata == nil {
		return nil
	}

	if usageMap, ok := m.Metadata["usage"].(map[string]interface{}); ok {
		usage := &Usage{}

		if promptTokens, ok := usageMap["prompt_tokens"].(int); ok {
			usage.PromptTokens = promptTokens
		}
		if completionTokens, ok := usageMap["completion_tokens"].(int); ok {
			usage.CompletionTokens = completionTokens
		}
		if totalTokens, ok := usageMap["total_tokens"].(int); ok {
			usage.TotalTokens = totalTokens
		}

		return usage
	}

	return nil
}

// SetUsage sets the usage statistics in the message metadata
func (m *Message) SetUsage(usage *Usage) {
	if usage == nil {
		return
	}

	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}

	m.Metadata["usage"] = map[string]interface{}{
		"prompt_tokens":     usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"total_tokens":      usage.TotalTokens,
	}
}

func MessagesToString(messages []Message) string {
	var sb strings.Builder
	for _, message := range messages {
		if message.FuncCall != nil || message.Role == FunctionRole || message.Role == SystemRole {
			continue
		}
		sb.WriteString(message.Role)
		sb.WriteString(": ")
		sb.WriteString(message.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}
