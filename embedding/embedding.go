package embedding

import (
	"context"
)

// Embedder represents an interface for text embedding operations
type Embedder interface {
	// EmbedDocuments converts a slice of documents into vector embeddings
	EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error)

	// EmbedQuery converts a single query text into a vector embedding
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}
