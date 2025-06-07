package core

import (
	"context"
	"reflect"
)


// Command provides the interface for command management
// This allows the registry and help system to work with commands of different types
type Command interface {
	// Basic command information
	GetName() string
	GetDescription() string
	GetAliases() []string
	GetConfigType() reflect.Type
	
	// Execution
	Execute(ctx context.Context, config any) error
	
	// Hierarchy support
	HasSubcommands() bool
	AddSubcommand(cmd Command) error
	GetSubcommand(name string) (Command, bool)
	ListSubcommands() map[string]Command
	GetPath() []string
	
	// Parent relationship
	SetParent(parent Command)
	GetParent() Command
	
	// Internal access to the typed command instance
	GetInstance() any
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
