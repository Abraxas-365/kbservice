package vectorstore

import (
	"fmt"
)

// ErrorCode represents specific error types in vector store operations
type ErrorCode string

const (
	ErrCodeDBExists          ErrorCode = "DB_EXISTS"
	ErrCodeDBNotFound        ErrorCode = "DB_NOT_FOUND"
	ErrCodeInitFailed        ErrorCode = "INIT_FAILED"
	ErrCodeAddFailed         ErrorCode = "ADD_FAILED"
	ErrCodeSearchFailed      ErrorCode = "SEARCH_FAILED"
	ErrCodeDeleteFailed      ErrorCode = "DELETE_FAILED"
	ErrCodeInvalidDimensions ErrorCode = "INVALID_DIMENSIONS"
	ErrCodeInvalidFilter     ErrorCode = "INVALID_FILTER"
	ErrCodeEmbeddingFailed   ErrorCode = "EMBEDDING_FAILED"
)

// VectorStoreError represents an error that occurred in vector store operations
type VectorStoreError struct {
	Code    ErrorCode
	Op      string
	Store   string
	Message string
	Err     error
}

func (e *VectorStoreError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (store: %s, operation: %s) - %v",
			e.Code, e.Message, e.Store, e.Op, e.Err)
	}
	return fmt.Sprintf("%s: %s (store: %s, operation: %s)",
		e.Code, e.Message, e.Store, e.Op)
}

func (e *VectorStoreError) Unwrap() error {
	return e.Err
}

// Helper functions to create errors
func NewDBExistsError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeDBExists,
		Op:      "InitDB",
		Store:   store,
		Message: "database already exists",
		Err:     err,
	}
}

func NewDBNotFoundError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeDBNotFound,
		Op:      "Access",
		Store:   store,
		Message: "database not found",
		Err:     err,
	}
}

func NewInitFailedError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeInitFailed,
		Op:      "InitDB",
		Store:   store,
		Message: "failed to initialize database",
		Err:     err,
	}
}

func NewAddFailedError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeAddFailed,
		Op:      "AddDocuments",
		Store:   store,
		Message: "failed to add documents",
		Err:     err,
	}
}

func NewSearchFailedError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeSearchFailed,
		Op:      "SimilaritySearch",
		Store:   store,
		Message: "failed to perform similarity search",
		Err:     err,
	}
}

func NewDeleteFailedError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeDeleteFailed,
		Op:      "Delete",
		Store:   store,
		Message: "failed to delete documents",
		Err:     err,
	}
}

func NewInvalidDimensionsError(store string, expected, got int) error {
	return &VectorStoreError{
		Code:    ErrCodeInvalidDimensions,
		Op:      "AddDocuments",
		Store:   store,
		Message: fmt.Sprintf("invalid vector dimensions: expected %d, got %d", expected, got),
	}
}

func NewInvalidFilterError(store string, details string) error {
	return &VectorStoreError{
		Code:    ErrCodeInvalidFilter,
		Op:      "Filter",
		Store:   store,
		Message: fmt.Sprintf("invalid filter: %s", details),
	}
}

func NewEmbeddingFailedError(store string, err error) error {
	return &VectorStoreError{
		Code:    ErrCodeEmbeddingFailed,
		Op:      "Embedding",
		Store:   store,
		Message: "failed to generate embeddings",
		Err:     err,
	}
}
