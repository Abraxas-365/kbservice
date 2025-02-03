package storage

// StorageError represents errors that can occur during storage operations
type StorageError struct {
	Op      string
	Key     string
	Err     error
	Code    string
	Message string
}

// Error implements the error interface
func (e *StorageError) Error() string {
	if e.Key == "" {
		return "storage." + e.Op + ": " + e.Message
	}
	return "storage." + e.Op + " " + e.Key + ": " + e.Message
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrCodeNotFound         = "NotFound"
	ErrCodeAlreadyExists    = "AlreadyExists"
	ErrCodeInvalidArgument  = "InvalidArgument"
	ErrCodePermissionDenied = "PermissionDenied"
	ErrCodeUnauthenticated  = "Unauthenticated"
	ErrCodeInternal         = "Internal"
)

// NewStorageError creates a new StorageError
func NewStorageError(op, key string, err error, code, message string) *StorageError {
	return &StorageError{
		Op:      op,
		Key:     key,
		Err:     err,
		Code:    code,
		Message: message,
	}
}
