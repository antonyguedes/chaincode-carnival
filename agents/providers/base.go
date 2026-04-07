package providers

// LLMProvider maps the interface required for any agent dynamically fetching completions.
type LLMProvider interface {
	// Query takes a simple prompt string and returns the direct text response from the model
	Query(prompt string) (string, error)
	// Name returns the identifier tag for UI Arena rendering
	Name() string
}
