// Package cli provides a fluent, developer-friendly API for creating CLI applications.
//
// This package offers both traditional struct-based configuration and modern
// fluent builders to create powerful command-line interfaces with minimal boilerplate.
//
// # Quick Start
//
// The simplest way to create a CLI application:
//
//	func main() {
//	    cli.Quick("my-app",
//	        cli.Cmd("hello", "Say hello", func() error {
//	            fmt.Println("Hello, World!")
//	            return nil
//	        }),
//	    )
//	}
//
// # Fluent API
//
// For more control, use the fluent builder:
//
//	app := cli.New("my-app").
//	    Version("1.0.0").
//	    Description("My awesome CLI").
//	    Interactive().
//	    AutoConfig().
//	    WithCommands(
//	        cli.Cmd("deploy", "Deploy application", deployHandler),
//	        cli.VersionCmd("1.0.0"),
//	    ).
//	    Build()
//
//	app.RunWithArgs(context.Background())
//
// # Configuration
//
// The framework supports automatic config file loading (YAML/JSON),
// environment variables, and interactive prompting:
//
//	app := cli.New("my-app").
//	    AutoConfig().     // Load my-app.yaml/json automatically
//	    Interactive().    // Prompt for missing required fields
//	    Recovery().       // Handle panics gracefully
//	    Logging()         // Log command execution
//
// # Command Types
//
// Simple commands with no arguments:
//
//	cli.Cmd("status", "Show status", func() error {
//	    fmt.Println("OK")
//	    return nil
//	})
//
// Complex commands with struct-based configuration:
//
//	type DeployConfig struct {
//	    Environment string `posix:"e,env,Environment,required"`
//	    Version     string `posix:"v,version,Version,required"`
//	}
//
//	core.NewCommand("deploy", "Deploy app", func(ctx context.Context, config DeployConfig) error {
//	    // Access config.Environment and config.Version
//	    return deployApp(config.Environment, config.Version)
//	})
//
// # Presets
//
// Use preset configurations for common scenarios:
//
//	cli.Dev("my-app", commands...)    // Development: interactive, colors, recovery
//	cli.Prod("my-app", commands...)   // Production: logging, recovery, no colors
//	cli.Quick("my-app", commands...)  // Minimal: just recovery
//
// # Features
//
//   - Type-safe command configuration with generics
//   - Automatic help generation with colored output
//   - POSIX-compliant argument parsing
//   - Interactive prompting for missing required fields
//   - Configuration file support (YAML, JSON)
//   - Environment variable integration
//   - Shell completion generation
//   - Middleware support (logging, recovery, timeout)
//   - Lifecycle hooks (before/after command execution)
//   - Intelligent error messages with suggestions
//
// # Examples
//
// See the examples directory for comprehensive demonstrations of different
// usage patterns and features.
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"claude-code-test/app"
	"claude-code-test/config"
	"claude-code-test/core"
)

// App represents a CLI application builder with fluent API
type App struct {
	name        string
	version     string
	description string
	author      string
	options     []config.Option
	commands    []any
}

// New creates a new CLI application with fluent API
func New(name string) *App {
	return &App{
		name:     name,
		version:  "1.0.0",
		options:  make([]config.Option, 0),
		commands: make([]any, 0),
	}
}

// Quick creates a CLI app with sensible defaults and runs it immediately
func Quick(name string, commands ...any) {
	New(name).
		Recovery().
		WithCommands(commands...).
		RunWithArgs(context.Background())
}

// Dev creates a development CLI with all developer-friendly features
func Dev(name string, commands ...any) {
	New(name).
		Interactive().
		ColoredOutput(true).
		Recovery().
		Logging().
		AutoConfig().
		WithCommands(commands...).
		RunWithArgs(context.Background())
}

// Prod creates a production CLI with production-ready defaults
func Prod(name string, commands ...any) {
	New(name).
		Recovery().
		Logging().
		ColoredOutput(false).
		Timeout(30 * time.Second).
		WithCommands(commands...).
		RunWithArgs(context.Background())
}

// Fluent methods for App

// Version sets the application version
func (a *App) Version(version string) *App {
	a.version = version
	return a
}

// Description sets the application description
func (a *App) Description(description string) *App {
	a.description = description
	return a
}

// Author sets the application author
func (a *App) Author(author string) *App {
	a.author = author
	return a
}

// Recovery enables panic recovery middleware
func (a *App) Recovery() *App {
	a.options = append(a.options, config.WithRecovery())
	return a
}

// Logging enables logging middleware
func (a *App) Logging() *App {
	a.options = append(a.options, config.WithLogging())
	return a
}

// Interactive enables interactive mode for missing required fields
func (a *App) Interactive() *App {
	a.options = append(a.options, config.WithInteractiveMode(true))
	return a
}

// ColoredOutput enables or disables colored output
func (a *App) ColoredOutput(enabled bool) *App {
	a.options = append(a.options, config.WithColoredOutput(enabled))
	return a
}

// AutoConfig enables automatic config file loading using the app name
func (a *App) AutoConfig() *App {
	a.options = append(a.options, config.WithAutoLoadConfig(true))
	return a
}

// ConfigFile enables automatic config file loading with the given filename
func (a *App) ConfigFile(filename string) *App {
	a.options = append(a.options, config.WithConfigFile(filename))
	return a
}

// Timeout sets the default command timeout
func (a *App) Timeout(timeout time.Duration) *App {
	a.options = append(a.options, config.WithDefaultTimeout(timeout))
	return a
}

// WithCommands adds multiple commands to the application
func (a *App) WithCommands(commands ...any) *App {
	a.commands = append(a.commands, commands...)
	return a
}

// AddCommand adds a single command to the application
func (a *App) AddCommand(command any) *App {
	a.commands = append(a.commands, command)
	return a
}

// BeforeAll sets a hook to run before all commands
func (a *App) BeforeAll(hook func(*core.ExecutionContext) error) *App {
	a.options = append(a.options, config.WithBeforeAll(hook))
	return a
}

// AfterAll sets a hook to run after all commands
func (a *App) AfterAll(hook func(*core.ExecutionContext) error) *App {
	a.options = append(a.options, config.WithAfterAll(hook))
	return a
}

// BeforeEach sets a hook to run before each command
func (a *App) BeforeEach(hook func(*core.ExecutionContext) error) *App {
	a.options = append(a.options, config.WithBeforeEach(hook))
	return a
}

// AfterEach sets a hook to run after each command
func (a *App) AfterEach(hook func(*core.ExecutionContext) error) *App {
	a.options = append(a.options, config.WithAfterEach(hook))
	return a
}

// Build creates the underlying CLI application
func (a *App) Build() *app.Application {
	// Prepare all options
	allOptions := []config.Option{
		config.WithName(a.name),
		config.WithVersion(a.version),
	}
	
	if a.description != "" {
		allOptions = append(allOptions, config.WithDescription(a.description))
	}
	
	if a.author != "" {
		allOptions = append(allOptions, config.WithAuthor(a.author))
	}
	
	// Add all builder options
	allOptions = append(allOptions, a.options...)
	
	// Create application
	application := app.NewApplicationWithOptions(allOptions...)
	
	// Register all commands
	for _, cmd := range a.commands {
		if err := application.Register(cmd); err != nil {
			panic("Failed to register command: " + err.Error())
		}
	}
	
	return application
}

// Run builds and runs the application with the given arguments
func (a *App) Run(ctx context.Context, args []string) int {
	return a.Build().Run(ctx, args)
}

// RunWithArgs builds and runs the application with os.Args
func (a *App) RunWithArgs(ctx context.Context) {
	exitCode := a.Build().RunWithArgs(ctx)
	os.Exit(exitCode)
}

// Simple command creation helpers

// Cmd creates a command that takes no arguments
func Cmd(name, description string, handler func() error) any {
	return core.NewCommand(name, description, func(ctx context.Context, config struct{}) error {
		return handler()
	})
}

// Common commands that most CLIs need

// VersionCmd creates a standard version command
func VersionCmd(version string) any {
	return Cmd("version", "Show application version", func() error {
		fmt.Println(version)
		return nil
	})
}

// HelpCmd creates a help command (though most apps handle this automatically)
func HelpCmd() any {
	return Cmd("help", "Show help information", func() error {
		fmt.Println("Use --help with any command for detailed usage")
		return nil
	})
}

// StandardCmds returns commonly used commands for any CLI
func StandardCmds(appName, version string) []any {
	return []any{
		VersionCmd(version),
		Cmd("completion", "Generate shell completion", func() error {
			fmt.Printf("Completion for %s would be generated here\n", appName)
			return nil
		}),
	}
}