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
	LLM            *llm.LLM // Optional LLM
}

// Option is a function type to modify Options
type Option func(*Options)

// Default options
func defaultOptions() *Options {
	return &Options{
		ScoreThreshold: 0.0,
		LLM:            nil, // Default to no LLM
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

// WithLLM sets the LLM for the knowledge base
func WithLLM(llm *llm.LLM) Option {
	return func(o *Options) {
		o.LLM = llm
	}
}
