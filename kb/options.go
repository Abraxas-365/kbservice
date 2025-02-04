package kb

import (
	"github.com/Abraxas-365/kbservice/llm"
	"github.com/Abraxas-365/kbservice/vectorstore"
)

// Options contains configuration for the knowledge base
type Options struct {
	Namespace      string
	ScoreThreshold float32
	Filters        vectorstore.Filter
	IndexName      string
	Dimensions     int
	Distance       vectorstore.DistanceMetric
	TopK           int
	LLM            *llm.LLM // Optional LLM
}

// Option is a function type to modify Options
type Option func(*Options)

// Default options
func defaultOptions() *Options {
	return &Options{
		ScoreThreshold: 0.0,
		Distance:       vectorstore.Cosine,
		TopK:           4,
		LLM:            nil, // Default to no LLM
	}
}

// WithNamespace sets the namespace for the knowledge base
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithScoreThreshold sets the minimum similarity score threshold
func WithScoreThreshold(threshold float32) Option {
	return func(o *Options) {
		o.ScoreThreshold = threshold
	}
}

// WithFilters sets default filters for queries
func WithFilters(filters vectorstore.Filter) Option {
	return func(o *Options) {
		o.Filters = filters
	}
}

// WithIndexName sets the index name
func WithIndexName(indexName string) Option {
	return func(o *Options) {
		o.IndexName = indexName
	}
}

// WithDimensions sets the vector dimensions
func WithDimensions(dimensions int) Option {
	return func(o *Options) {
		o.Dimensions = dimensions
	}
}

// WithDistanceMetric sets the distance calculation method
func WithDistanceMetric(metric vectorstore.DistanceMetric) Option {
	return func(o *Options) {
		o.Distance = metric
	}
}

// WithTopK sets the number of similar documents to retrieve
func WithTopK(k int) Option {
	return func(o *Options) {
		o.TopK = k
	}
}

// WithLLM sets the LLM for the knowledge base
func WithLLM(llm *llm.LLM) Option {
	return func(o *Options) {
		o.LLM = llm
	}
}
