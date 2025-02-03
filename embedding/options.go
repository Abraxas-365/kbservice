package embedding

// EmbeddingOptions represents configuration options for embedding operations
type EmbeddingOptions struct {
	// Model specifies which embedding model to use
	Model string

	// BatchSize specifies the maximum number of documents to embed in a single request
	BatchSize int

	// Normalize indicates whether to normalize the resulting vectors
	Normalize bool

	// Truncate indicates whether to truncate text that exceeds token limits
	Truncate bool
}

// Option is a function type to modify EmbeddingOptions
type Option func(*EmbeddingOptions)

// WithModel sets the embedding model
func WithModel(model string) Option {
	return func(o *EmbeddingOptions) {
		o.Model = model
	}
}

// WithBatchSize sets the batch size for document embedding
func WithBatchSize(size int) Option {
	return func(o *EmbeddingOptions) {
		o.BatchSize = size
	}
}

// WithNormalization sets whether to normalize vectors
func WithNormalization(normalize bool) Option {
	return func(o *EmbeddingOptions) {
		o.Normalize = normalize
	}
}

// WithTruncation sets whether to truncate long texts
func WithTruncation(truncate bool) Option {
	return func(o *EmbeddingOptions) {
		o.Truncate = truncate
	}
}
