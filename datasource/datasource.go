package datasource

import "context"

// Document represents a document from a data source
type Document struct {
	Content  string
	Metadata map[string]interface{}
	Source   string
}

// DataSource represents a source of documents
type DataSource interface {
	// Load loads documents from the source
	Load(ctx context.Context, opts ...Option) ([]Document, error)
}
