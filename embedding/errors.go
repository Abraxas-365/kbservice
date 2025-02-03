package embedding

import "fmt"

// EmbeddingError represents errors that can occur during embedding operations
type EmbeddingError struct {
	Op      string
	Err     error
	Code    string
	Message string
}

// Error implements the error interface
func (e *EmbeddingError) Error() string {
	return fmt.Sprintf("embedding.%s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying error
func (e *EmbeddingError) Unwrap() error {
	return e.Err
}

// Common error codes for embedding operations
const (
	ErrCodeInvalidInput       = "InvalidInput"
	ErrCodeTokenLimitExceeded = "TokenLimitExceeded"
	ErrCodeModelNotAvailable  = "ModelNotAvailable"
	ErrCodeRateLimitExceeded  = "RateLimitExceeded"
	ErrCodeContextCanceled    = "ContextCanceled"
	ErrCodeInvalidDimensions  = "InvalidDimensions"
	ErrCodeEmptyInput         = "EmptyInput"
	ErrCodeAPIError           = "APIError"
	ErrCodeInternal           = "Internal"
)

// NewEmbeddingError creates a new EmbeddingError
func NewEmbeddingError(op string, err error, code, message string) *EmbeddingError {
	return &EmbeddingError{
		Op:      op,
		Err:     err,
		Code:    code,
		Message: message,
	}
}

// Common error constructors for frequent error cases
func ErrInvalidInput(op string, err error, details string) error {
	return NewEmbeddingError(op, err, ErrCodeInvalidInput,
		fmt.Sprintf("invalid input: %s", details))
}

func ErrTokenLimitExceeded(op string, err error) error {
	return NewEmbeddingError(op, err, ErrCodeTokenLimitExceeded,
		"token limit exceeded for input text")
}

func ErrModelNotAvailable(op string, err error) error {
	return NewEmbeddingError(op, err, ErrCodeModelNotAvailable,
		"embedding model is not available")
}

func ErrRateLimitExceeded(op string, err error) error {
	return NewEmbeddingError(op, err, ErrCodeRateLimitExceeded,
		"rate limit exceeded for embedding requests")
}

func ErrEmptyInput(op string) error {
	return NewEmbeddingError(op, nil, ErrCodeEmptyInput,
		"input text or documents cannot be empty")
}
