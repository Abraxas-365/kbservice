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

	// Stream processes documents one at a time through the channel
	Stream(ctx context.Context, opts ...Option) (<-chan Document, <-chan error)
}
