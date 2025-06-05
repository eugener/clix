package app

import (
	"context"
	"fmt"
	"os"

	"claude-code-test/config"
	"claude-code-test/core"
	"claude-code-test/help"
)

// Application represents a complete CLI application
type Application struct {
	config   *config.CLIConfig
	registry *core.Registry
	executor *core.Executor
	helpGen  *help.Generator
}

// NewApplication creates a new CLI application with the given configuration
func NewApplication(cfg *config.CLIConfig) *Application {
	registry := core.NewRegistry()
	executor := core.NewExecutor(registry)
	
	// Apply middleware from configuration
	if len(cfg.Middleware) > 0 {
		executor.Use(cfg.Middleware...)
	}
	
	// Set logger if provided
	if cfg.Logger != nil {
		executor.SetLogger(cfg.Logger)
	}
	
	// Create help generator
	helpGen := help.NewGenerator(cfg.HelpConfig)
	
	return &Application{
		config:   cfg,
		registry: registry,
		executor: executor,
		helpGen:  helpGen,
	}
}

// NewApplicationWithOptions creates a new CLI application with functional options
func NewApplicationWithOptions(opts ...config.Option) *Application {
	cfg := config.DefaultConfig()
	cfg.Apply(opts...)
	return NewApplication(cfg)
}

// Register adds a command to the application
func (app *Application) Register(cmd any) error {
	return app.registry.Register(cmd)
}

// RegisterCommands adds multiple commands to the application
func (app *Application) RegisterCommands(commands ...any) error {
	for _, cmd := range commands {
		if err := app.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

// Run executes the CLI application with the given arguments
func (app *Application) Run(ctx context.Context, args []string) int {
	// Apply before all hook
	if app.config.BeforeAll != nil {
		execCtx := core.NewExecutionContext(ctx, "", args)
		if err := app.config.BeforeAll(execCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Before all hook failed: %v\n", err)
			return app.config.ErrorHandler(err)
		}
	}
	
	defer func() {
		// Apply after all hook
		if app.config.AfterAll != nil {
			execCtx := core.NewExecutionContext(ctx, "", args)
			if err := app.config.AfterAll(execCtx); err != nil {
				fmt.Fprintf(os.Stderr, "After all hook failed: %v\n", err)
			}
		}
	}()
	
	// Handle no arguments - show main help
	if len(args) == 0 {
		app.showMainHelp()
		return 0
	}
	
	// Handle global help requests
	if app.isHelpRequest(args[0]) {
		return app.handleHelp(args)
	}
	
	// Handle version request
	if app.isVersionRequest(args[0]) {
		fmt.Printf("%s version %s\n", app.config.Name, app.config.Version)
		if app.config.Author != "" {
			fmt.Printf("Author: %s\n", app.config.Author)
		}
		return 0
	}
	
	// Execute command
	commandName := args[0]
	commandArgs := args[1:]
	
	// Apply before each hook
	if app.config.BeforeEach != nil {
		execCtx := core.NewExecutionContext(ctx, commandName, commandArgs)
		if err := app.config.BeforeEach(execCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Before each hook failed: %v\n", err)
			return app.config.ErrorHandler(err)
		}
	}
	
	defer func() {
		// Apply after each hook
		if app.config.AfterEach != nil {
			execCtx := core.NewExecutionContext(ctx, commandName, commandArgs)
			if err := app.config.AfterEach(execCtx); err != nil {
				fmt.Fprintf(os.Stderr, "After each hook failed: %v\n", err)
			}
		}
	}()
	
	// Execute the command
	if err := app.executor.Execute(ctx, commandName, commandArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return app.config.ErrorHandler(err)
	}
	
	return 0
}

// RunWithArgs executes the CLI application with os.Args
func (app *Application) RunWithArgs(ctx context.Context) int {
	return app.Run(ctx, os.Args[1:])
}

// showMainHelp displays the main help text
func (app *Application) showMainHelp() {
	commands := make(map[string]help.CommandInfo)
	for name, desc := range app.registry.ListCommands() {
		commands[name] = help.CommandInfo{
			Name:        desc.GetName(),
			Description: desc.GetDescription(),
			ConfigType:  desc.GetConfigType(),
		}
	}
	fmt.Print(app.helpGen.GenerateMainHelp(commands))
}

// handleHelp handles help requests
func (app *Application) handleHelp(args []string) int {
	if len(args) > 1 {
		// Command-specific help
		cmdName := args[1]
		if desc, exists := app.registry.GetCommand(cmdName); exists {
			info := help.CommandInfo{
				Name:        desc.GetName(),
				Description: desc.GetDescription(),
				ConfigType:  desc.GetConfigType(),
				Examples: []string{
					fmt.Sprintf("%s %s [options]", app.config.Name, cmdName),
				},
			}
			helpText, err := app.helpGen.GenerateCommandHelp(cmdName, info)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating help: %v\n", err)
				return 1
			}
			fmt.Print(helpText)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
			return 1
		}
	} else {
		// Main help
		app.showMainHelp()
	}
	return 0
}

// isHelpRequest checks if the argument is a help request
func (app *Application) isHelpRequest(arg string) bool {
	return arg == "help" || arg == "--help" || arg == "-h"
}

// isVersionRequest checks if the argument is a version request
func (app *Application) isVersionRequest(arg string) bool {
	return arg == "version" || arg == "--version" || arg == "-v"
}

// GetConfig returns the application configuration
func (app *Application) GetConfig() *config.CLIConfig {
	return app.config
}

// GetRegistry returns the command registry
func (app *Application) GetRegistry() *core.Registry {
	return app.registry
}

// GetExecutor returns the command executor
func (app *Application) GetExecutor() *core.Executor {
	return app.executor
}

// GetHelpGenerator returns the help generator
func (app *Application) GetHelpGenerator() *help.Generator {
	return app.helpGen
}

// Quick application creation functions

// QuickApp creates a simple application with minimal configuration
func QuickApp(name, description string, commands ...any) *Application {
	app := NewApplicationWithOptions(
		config.WithName(name),
		config.WithDescription(description),
		config.WithRecovery(),
		config.WithLogging(),
	)
	
	for _, cmd := range commands {
		if err := app.Register(cmd); err != nil {
			panic(fmt.Sprintf("Failed to register command: %v", err))
		}
	}
	
	return app
}

// DevApp creates an application with development-friendly settings
func DevApp(name, description string, commands ...any) *Application {
	cfg := config.DevelopmentConfig()
	cfg.Apply(
		config.WithName(name),
		config.WithDescription(description),
	)
	
	app := NewApplication(cfg)
	
	for _, cmd := range commands {
		if err := app.Register(cmd); err != nil {
			panic(fmt.Sprintf("Failed to register command: %v", err))
		}
	}
	
	return app
}

// ProdApp creates an application with production-friendly settings
func ProdApp(name, description string, commands ...any) *Application {
	cfg := config.ProductionConfig()
	cfg.Apply(
		config.WithName(name),
		config.WithDescription(description),
	)
	
	app := NewApplication(cfg)
	
	for _, cmd := range commands {
		if err := app.Register(cmd); err != nil {
			panic(fmt.Sprintf("Failed to register command: %v", err))
		}
	}
	
	return app
}