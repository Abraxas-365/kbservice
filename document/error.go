package document

import "fmt"

// SplitterError represents errors that can occur during text splitting
type SplitterError struct {
	Op      string
	Message string
	Err     error
}

func (e *SplitterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("splitter.%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("splitter.%s: %s", e.Op, e.Message)
}

var (
	ErrMetadataTextMismatch = &SplitterError{
		Op:      "split_documents",
		Message: "number of texts and metadata entries must match",
	}
)
