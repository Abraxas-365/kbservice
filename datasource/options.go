package datasource

// LoadOptions represents options for loading documents
type LoadOptions struct {
	// Recursive indicates whether to recursively load from directories/prefixes
	Recursive bool
	// Filter is a function that determines whether to load a document
	Filter func(metadata map[string]interface{}) bool
	// MaxItems is the maximum number of items to load (0 for no limit)
	MaxItems int
}

// Option is a function type to modify LoadOptions
type Option func(*LoadOptions)

// WithRecursive sets whether to load recursively
func WithRecursive(recursive bool) Option {
	return func(o *LoadOptions) {
		o.Recursive = recursive
	}
}

// WithFilter sets a filter function for documents
func WithFilter(filter func(metadata map[string]interface{}) bool) Option {
	return func(o *LoadOptions) {
		o.Filter = filter
	}
}

// WithMaxItems sets the maximum number of items to load
func WithMaxItems(max int) Option {
	return func(o *LoadOptions) {
		o.MaxItems = max
	}
}
