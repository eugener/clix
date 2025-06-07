package core

import "context"

// Command represents a type-safe CLI command with configuration of type T
type Command[T any] interface {
	// Run executes the command with the given context and parsed configuration
	Run(ctx context.Context, config T) error

	// Name returns the command name
	Name() string

	// Description returns a brief description of the command
	Description() string
}

// CLI represents the main CLI application
type CLI interface {
	// Register adds a command to the CLI
	Register(cmd any) error

	// Execute parses arguments and runs the appropriate command
	Execute(ctx context.Context, args []string) error

	// Help returns help text for the CLI or a specific command
	Help(command string) string
}

// Parser handles POSIX-compliant argument parsing
type Parser interface {
	// Parse parses command line arguments into the target struct
	Parse(args []string, target any) error

	// ValidateStruct validates struct field constraints
	ValidateStruct(target any) error
}

// Result represents the outcome of argument parsing
type Result struct {
	Command    string
	Flags      map[string]any
	Positional []string
	Remaining  []string
}
