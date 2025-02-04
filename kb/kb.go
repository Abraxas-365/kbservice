package kb

import (
	"context"

	"github.com/Abraxas-365/kbservice/datasource"
	"github.com/Abraxas-365/kbservice/document"
	"github.com/Abraxas-365/kbservice/embedding"
	"github.com/Abraxas-365/kbservice/llm"
	"github.com/Abraxas-365/kbservice/vectorstore"
)

// KnowledgeBase represents the main knowledge base system
type KnowledgeBase struct {
	embedder   embedding.Embedder
	vStore     *vectorstore.VectorStore
	store      vectorstore.Store
	datasource datasource.DataSource
	splitter   document.Splitter
	opts       *Options
}

// New creates a new KnowledgeBase instance with the provided options
func New(
	embedder embedding.Embedder,
	store vectorstore.Store,
	datasource datasource.DataSource,
	splitter document.Splitter,
	opts ...Option,
) (*KnowledgeBase, error) {
	// Initialize with default options
	options := defaultOptions()

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	// Create vector store with options
	vStore := vectorstore.New(
		store,
		embedder,
		vectorstore.WithNamespace(options.Namespace),
		vectorstore.WithScoreThreshold(options.ScoreThreshold),
		vectorstore.WithFilters(options.Filters),
		vectorstore.WithIndexName(options.IndexName),
		vectorstore.WithDimensions(options.Dimensions),
		vectorstore.WithDistanceMetric(options.Distance),
	)

	kb := &KnowledgeBase{
		embedder:   embedder,
		vStore:     vStore,
		store:      store,
		datasource: datasource,
		splitter:   splitter,
		opts:       options,
	}

	return kb, nil
}

// GetOptions returns a copy of the current options
func (kb *KnowledgeBase) GetOptions() Options {
	return *kb.opts
}

// UpdateOptions updates the knowledge base options
func (kb *KnowledgeBase) UpdateOptions(opts ...Option) {
	for _, opt := range opts {
		opt(kb.opts)
	}

	// Update vector store options
	kb.vStore = vectorstore.New(
		kb.store,
		kb.embedder,
		vectorstore.WithNamespace(kb.opts.Namespace),
		vectorstore.WithScoreThreshold(kb.opts.ScoreThreshold),
		vectorstore.WithFilters(kb.opts.Filters),
		vectorstore.WithIndexName(kb.opts.IndexName),
		vectorstore.WithDimensions(kb.opts.Dimensions),
		vectorstore.WithDistanceMetric(kb.opts.Distance),
	)
}

// HasLLM returns whether the knowledge base has an LLM configured
func (kb *KnowledgeBase) HasLLM() bool {
	return kb.opts.LLM != nil
}

// GetLLM returns the configured LLM, may be nil
func (kb *KnowledgeBase) GetLLM() *llm.LLM {
	return kb.opts.LLM
}

// Close releases any resources held by the knowledge base
func (kb *KnowledgeBase) Close() error {
	// Add any cleanup logic here if needed
	return nil
}

func (kb *KnowledgeBase) InitStore(ctx context.Context, forceRecreate bool) error {
	return kb.store.InitDB(ctx, forceRecreate)
}
