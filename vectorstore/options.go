package vectorstore

// Options contains configuration for the vector store
type Options struct {
	Namespace      string
	ScoreThreshold float32
	Filters        Filter
	IndexName      string
	Dimensions     int
	Distance       DistanceMetric
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

// WithNamespace sets the namespace for the vector store
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
func WithFilters(filters Filter) Option {
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
func WithDistanceMetric(metric DistanceMetric) Option {
	return func(o *Options) {
		o.Distance = metric
	}
}
