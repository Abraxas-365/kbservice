package document

// Document represents a text document with metadata
type Document struct {
	PageContent string                 `json:"page_content"`
	Metadata    map[string]interface{} `json:"metadata"`
}
