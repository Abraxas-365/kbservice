package vectorstore

import (
	"context"

	"github.com/Abraxas-365/kbservice/document"
	"github.com/Abraxas-365/kbservice/embedding"
)

// Filter represents a query filter
type Filter map[string]interface{}

// Document extends document.Document with a score
type Document struct {
	PageContent string                 `json:"page_content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Score       float32                `json:"score"`
}

// ToDocument converts a vectorstore.Document to document.Document
func (d Document) ToDocument() document.Document {
	return document.Document{
		PageContent: d.PageContent,
		Metadata:    d.Metadata,
	}
}

// FromDocument creates a vectorstore.Document from document.Document
func FromDocument(doc document.Document) Document {
	return Document{
		PageContent: doc.PageContent,
		Metadata:    doc.Metadata,
	}
}

// Store interface defines the operations that any vector database adapter must implement
type Store interface {
	// AddDocuments adds documents to the vector store
	AddDocuments(ctx context.Context, docs []Document, vectors [][]float32) error

	// SimilaritySearch performs a similarity search using the provided vector
	SimilaritySearch(ctx context.Context, vector []float32, limit int, filter Filter) ([]Document, error)

	// Delete removes documents from the store
	Delete(ctx context.Context, filter Filter) error
}

// VectorStore is the main struct that combines the database adapter and embedder
type VectorStore struct {
	store    Store
	embedder embedding.Embedder
	opts     *Options
}

// New creates a new VectorStore instance
func New(store Store, embedder embedding.Embedder, opts ...Option) *VectorStore {
	options := &Options{
		ScoreThreshold: 0.0,
		Distance:       Cosine,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &VectorStore{
		store:    store,
		embedder: embedder,
		opts:     options,
	}
}

// AddDocuments adds documents to the vector store
func (vs *VectorStore) AddDocuments(ctx context.Context, docs []document.Document) error {
	texts := make([]string, len(docs))
	vsDocs := make([]Document, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
		vsDocs[i] = FromDocument(doc)
	}

	vectors, err := vs.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	return vs.store.AddDocuments(ctx, vsDocs, vectors)
}

// SimilaritySearch performs a similarity search using the query text
func (vs *VectorStore) SimilaritySearch(ctx context.Context, query string, limit int, filter Filter) ([]Document, error) {
	vector, err := vs.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	// Merge default filters with query filters
	mergedFilter := make(Filter)
	if vs.opts.Filters != nil {
		for k, v := range vs.opts.Filters {
			mergedFilter[k] = v
		}
	}
	if filter != nil {
		for k, v := range filter {
			mergedFilter[k] = v
		}
	}

	vsDocs, err := vs.store.SimilaritySearch(ctx, vector, limit, mergedFilter)
	if err != nil {
		return nil, err
	}

	// Apply score threshold and convert to document.Document
	docs := make([]Document, 0, len(vsDocs))
	for _, vsDoc := range vsDocs {
		if vs.opts.ScoreThreshold <= 0 || vsDoc.Score >= vs.opts.ScoreThreshold {
			docs = append(docs, vsDoc)
		}
	}

	return docs, nil
}

// Delete removes documents from the store
func (vs *VectorStore) Delete(ctx context.Context, filter Filter) error {
	return vs.store.Delete(ctx, filter)
}
