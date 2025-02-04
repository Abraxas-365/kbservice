package llm

// FunctionCall represents a function call in the chat
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Function represents a function that can be called by the LLM
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"` // JSON Schema object
}
