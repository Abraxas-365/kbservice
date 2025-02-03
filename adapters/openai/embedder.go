package openai

import (
	"context"
	"fmt"

	"github.com/Abraxas-365/kbservice/embedding"
	"github.com/sashabaranov/go-openai"
)

type OpenAIEmbedder struct {
	client  *openai.Client
	options *embedding.EmbeddingOptions
}

// DefaultOptions returns the default options for OpenAI embeddings
func DefaultOptions() *embedding.EmbeddingOptions {
	return &embedding.EmbeddingOptions{
		Model:     string(openai.AdaEmbeddingV2),
		BatchSize: 100,
		Normalize: true,
		Truncate:  true,
	}
}

// NewOpenAIEmbedder creates a new OpenAI embedder with the given API key and options
func NewOpenAIEmbedder(apiKey string, opts ...embedding.Option) *OpenAIEmbedder {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	client := openai.NewClient(apiKey)

	return &OpenAIEmbedder{
		client:  client,
		options: options,
	}
}

// EmbedDocuments implements the Embedder interface
func (e *OpenAIEmbedder) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	if len(documents) == 0 {
		return nil, embedding.ErrEmptyInput("EmbedDocuments")
	}

	// Process in batches if needed
	if len(documents) > e.options.BatchSize {
		return e.embedInBatches(ctx, documents)
	}

	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: documents,
		Model: openai.EmbeddingModel(e.options.Model),
	})

	if err != nil {
		return nil, e.handleError("EmbedDocuments", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, item := range resp.Data {
		embeddings[i] = item.Embedding
		if e.options.Normalize {
			normalizeVector(embeddings[i])
		}
	}

	return embeddings, nil
}

// EmbedQuery implements the Embedder interface
func (e *OpenAIEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, embedding.ErrEmptyInput("EmbedQuery")
	}

	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(e.options.Model),
	})

	if err != nil {
		return nil, e.handleError("EmbedQuery", err)
	}

	if len(resp.Data) == 0 {
		return nil, embedding.NewEmbeddingError("EmbedQuery", nil, embedding.ErrCodeAPIError,
			"no embedding returned from API")
	}

	embedding := resp.Data[0].Embedding
	if e.options.Normalize {
		normalizeVector(embedding)
	}

	return embedding, nil
}

// embedInBatches processes documents in batches
func (e *OpenAIEmbedder) embedInBatches(ctx context.Context, documents []string) ([][]float32, error) {
	var allEmbeddings [][]float32

	for i := 0; i < len(documents); i += e.options.BatchSize {
		end := i + e.options.BatchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		batchEmbeddings, err := e.EmbedDocuments(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("error processing batch %d: %w", i/e.options.BatchSize, err)
		}

		allEmbeddings = append(allEmbeddings, batchEmbeddings...)
	}

	return allEmbeddings, nil
}

// handleError converts OpenAI API errors to embedding errors
func (e *OpenAIEmbedder) handleError(op string, err error) error {
	if err == nil {
		return nil
	}

	switch apiErr := err.(type) {
	case *openai.APIError:
		switch apiErr.HTTPStatusCode {
		case 400:
			return embedding.ErrInvalidInput(op, err, apiErr.Message)
		case 401:
			return embedding.NewEmbeddingError(op, err, "Unauthorized", "invalid API key")
		case 429:
			return embedding.ErrRateLimitExceeded(op, err)
		case 500:
			return embedding.NewEmbeddingError(op, err, embedding.ErrCodeModelNotAvailable,
				"OpenAI API server error")
		default:
			return embedding.NewEmbeddingError(op, err, embedding.ErrCodeAPIError,
				fmt.Sprintf("OpenAI API error: %s", apiErr.Message))
		}
	default:
		return embedding.NewEmbeddingError(op, err, embedding.ErrCodeInternal,
			"unexpected error")
	}
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vector []float32) {
	var sum float32
	for _, v := range vector {
		sum += v * v
	}
	magnitude := float32(1)
	if sum > 0 {
		magnitude = float32(1 / float32(sum))
	}
	for i := range vector {
		vector[i] *= magnitude
	}
}
