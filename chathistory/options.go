package chathistory

import "github.com/google/uuid"

type IDGenerator func() string

// Options contains configuration for chat history memory
type Options struct {
	MaxMessages  int         // Maximum number of messages to keep in history
	ReturnLimit  int         // Default limit for GetMessages
	IncludeRoles []string    // Specific roles to include (empty means all)
	ExcludeRoles []string    // Specific roles to exclude
	SystemPrompt string      // System prompt to always include at the start
	GenerateID   IDGenerator // Function to generate conversation IDs
}

// Option is a function type to modify Options
type Option func(*Options)

// WithMaxMessages sets the maximum number of messages to keep
func WithMaxMessages(max int) Option {
	return func(o *Options) {
		o.MaxMessages = max
	}
}

// WithReturnLimit sets the default limit for GetMessages
func WithReturnLimit(limit int) Option {
	return func(o *Options) {
		o.ReturnLimit = limit
	}
}

// WithIncludeRoles sets specific roles to include
func WithIncludeRoles(roles []string) Option {
	return func(o *Options) {
		o.IncludeRoles = roles
	}
}

// WithExcludeRoles sets specific roles to exclude
func WithExcludeRoles(roles []string) Option {
	return func(o *Options) {
		o.ExcludeRoles = roles
	}
}

// WithSystemPrompt sets the system prompt to always include
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// DefaultIDGenerator generates a UUID string
func DefaultIDGenerator() string {
	return uuid.New().String()
}

// WithGenerateID sets the ID generation function
func WithGenerateID(generator IDGenerator) Option {
	return func(o *Options) {
		o.GenerateID = generator
	}
}

// DefaultOptions returns the default options
func DefaultOptions() *Options {
	return &Options{
		MaxMessages:  100,                // Default to 100 messages
		ReturnLimit:  20,                 // Default to last 20 messages
		IncludeRoles: []string{},         // Include all roles by default
		ExcludeRoles: []string{},         // Exclude none by default
		GenerateID:   DefaultIDGenerator, // Default ID generator
	}
}
