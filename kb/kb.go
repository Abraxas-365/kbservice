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
	embedder embedding.Embedder
	vStore   *vectorstore.VectorStore
	store    vectorstore.Store
	splitter document.Splitter
	opts     *Options
}

// New creates a new KnowledgeBase instance with the provided options
func New(
	embedder embedding.Embedder,
	store vectorstore.Store,
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
		vectorstore.WithScoreThreshold(options.ScoreThreshold),
		vectorstore.WithFilters(options.Filters),
	)

	kb := &KnowledgeBase{
		embedder: embedder,
		vStore:   vStore,
		store:    store,
		splitter: splitter,
		opts:     options,
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
		vectorstore.WithScoreThreshold(kb.opts.ScoreThreshold),
		vectorstore.WithFilters(kb.opts.Filters),
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

func (kb *KnowledgeBase) Sync(ctx context.Context, ds datasource.DataSource) error {
	docChan, errChan := ds.Stream(ctx)
	for {
		select {
		case doc, ok := <-docChan:
			if !ok {
				return nil
			}
			if err := kb.processData(ctx, doc); err != nil {
				return err
			}
		case err := <-errChan:
			return err
		}
	}
}

func (kb *KnowledgeBase) processData(ctx context.Context, doc datasource.Document) error {
	doc.Metadata["source"] = doc.Source
	docu := document.Document{
		PageContent: doc.Content,
		Metadata:    doc.Metadata,
	}

	chucks, err := document.SplitDocuments(kb.splitter, []document.Document{docu})
	if err != nil {
		return err
	}

	exist, err := kb.vStore.DocumentExists(ctx, chucks)
	if err != nil {
		return err
	}

	documentToAdd := []document.Document{}
	for i, doc := range chucks {
		if exist[i] {
			continue
		}
		documentToAdd = append(documentToAdd, doc)
	}

	kb.vStore.AddDocuments(ctx, documentToAdd)

	return nil
}

func (kb *KnowledgeBase) SimilaritySearch(
	ctx context.Context,
	query string,
	limit int,
	filter vectorstore.Filter,
) ([]vectorstore.Document, error) {
	return kb.vStore.SimilaritySearch(ctx, query, limit, filter)
}
