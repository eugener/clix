package core

// ParseResult represents the result of parsing command line arguments
type ParseResult struct {
	Flags      map[string]any
	Positional []string
	Remaining  []string
}