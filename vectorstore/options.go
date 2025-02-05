package vectorstore

// Options contains configuration for the vector store
type Options struct {
	ScoreThreshold float32
	Filters        Filter
}

// DistanceMetric represents the distance calculation method
type DistanceMetric string

const (
	Cosine     DistanceMetric = "cosine"
	Euclidean  DistanceMetric = "euclidean"
	DotProduct DistanceMetric = "dot_product"
)

// Option is a function type to modify Options
type Option func(*Options)

// WithScoreThreshold sets the minimum similarity score threshold
func WithScoreThreshold(threshold float32) Option {
	return func(o *Options) {
		o.ScoreThreshold = threshold
	}
}

// WithFilters sets default filters for queries
func WithFilters(filters Filter) Option {
	return func(o *Options) {
		o.Filters = filters
	}
}
