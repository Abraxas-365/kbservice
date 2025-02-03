package datasource

import "fmt"

// DataSourceError represents errors that can occur during data source operations
type DataSourceError struct {
	Source  string
	Op      string
	Err     error
	Code    string
	Message string
}

func (e *DataSourceError) Error() string {
	return fmt.Sprintf("datasource.%s [%s]: %s", e.Op, e.Source, e.Message)
}

func (e *DataSourceError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrCodeNotFound          = "NotFound"
	ErrCodeInvalidSource     = "InvalidSource"
	ErrCodeAccessDenied      = "AccessDenied"
	ErrCodeInvalidFormat     = "InvalidFormat"
	ErrCodeRateLimitExceeded = "RateLimitExceeded"
	ErrCodeInternal          = "Internal"
)
