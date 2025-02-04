package llm

// Message represents a chat message
type Message struct {
	Role     string                 `json:"role"`    // e.g., "system", "user", "assistant"
	Content  string                 `json:"content"` // The message content
	Name     string                 `json:"name,omitempty"`
	FuncCall *FunctionCall          `json:"function_call,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
