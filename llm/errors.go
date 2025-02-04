package llm

import "fmt"

// LLMError represents errors that can occur during LLM operations
type LLMError struct {
	Op      string
	Message string
	Err     error
}

func (e *LLMError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("llm.%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("llm.%s: %s", e.Op, e.Message)
}

func (e *LLMError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrInvalidInput       = "InvalidInput"
	ErrTokenLimitExceeded = "TokenLimitExceeded"
	ErrModelNotAvailable  = "ModelNotAvailable"
	ErrRateLimitExceeded  = "RateLimitExceeded"
	ErrContextCanceled    = "ContextCanceled"
	ErrAPIError           = "APIError"
	ErrInternal           = "Internal"
)
