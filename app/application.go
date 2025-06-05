package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"claude-code-test/config"
	"claude-code-test/configfile"
	"claude-code-test/core"
	"claude-code-test/help"
	"claude-code-test/interactive"
)

// Application represents a complete CLI application
type Application struct {
	config       *config.CLIConfig
	registry     *core.Registry
	executor     *core.Executor
	helpGen      *help.Generator
	errorFormat  *help.ErrorFormatter
	suggestions  *help.SuggestionEngine
	prompter     *interactive.SmartPrompter
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
	
	// Create error formatter and suggestion engine
	errorFormat := help.NewErrorFormatter(cfg.Name, cfg.HelpConfig.ColorEnabled)
	suggestions := help.NewSuggestionEngine()
	
	// Create interactive prompter
	prompter := interactive.NewSmartPrompter()
	
	return &Application{
		config:      cfg,
		registry:    registry,
		executor:    executor,
		helpGen:     helpGen,
		errorFormat: errorFormat,
		suggestions: suggestions,
		prompter:    prompter,
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
	
	// Check if command exists before proceeding
	commandName := args[0]
	if _, exists := app.registry.GetCommand(commandName); !exists {
		// Unknown command error with suggestions
		allCommands := app.getAllCommandNames()
		suggestions := app.suggestions.SuggestCommands(commandName, allCommands)
		
		errorCtx := help.NewErrorContext().
			Type(help.ErrorTypeUnknownCommand).
			Command(commandName).
			Suggestions(suggestions).
			AllCommands(allCommands).
			Build()
		
		formattedError := app.errorFormat.FormatError(fmt.Errorf("unknown command"), errorCtx)
		fmt.Fprint(os.Stderr, formattedError)
		return 1
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
	
	// Load configuration file if enabled
	if app.config.AutoLoadConfig {
		if err := app.loadConfigurationFile(commandName, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load configuration file: %v\n", err)
		}
	}
	
	// Execute the command
	if err := app.executor.Execute(ctx, commandName, commandArgs); err != nil {
		// Check if this is a missing required field error and interactive mode is enabled
		if app.config.InteractiveMode && app.isMissingRequiredFieldError(err) {
			if interactiveErr := app.handleInteractivePrompt(ctx, commandName, commandArgs, err); interactiveErr == nil {
				// Successfully prompted and got values, try again
				if retryErr := app.executor.Execute(ctx, commandName, commandArgs); retryErr == nil {
					return 0 // Success after interactive prompting
				}
			}
		}
		
		// Enhanced error formatting
		errorCtx := app.buildErrorContext(err, commandName, commandArgs)
		formattedError := app.errorFormat.FormatError(err, errorCtx)
		fmt.Fprint(os.Stderr, formattedError)
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

// buildErrorContext builds error context for better error messages
func (app *Application) buildErrorContext(err error, commandName string, args []string) *help.ErrorContext {
	errorMsg := err.Error()
	
	// Detect error type from message
	if strings.Contains(errorMsg, "unknown command") {
		allCommands := app.getAllCommandNames()
		suggestions := app.suggestions.SuggestCommands(commandName, allCommands)
		
		return help.NewErrorContext().
			Type(help.ErrorTypeUnknownCommand).
			Command(commandName).
			Suggestions(suggestions).
			AllCommands(allCommands).
			Build()
	}
	
	if strings.Contains(errorMsg, "unknown flag") {
		// Extract flag from error message
		flag := app.extractFlagFromError(errorMsg)
		allFlags := app.getAllFlagsForCommand(commandName)
		suggestions := app.suggestions.SuggestFlags(flag, allFlags)
		
		return help.NewErrorContext().
			Type(help.ErrorTypeUnknownFlag).
			Command(commandName).
			Flag(flag).
			Suggestions(suggestions).
			AllFlags(allFlags).
			Build()
	}
	
	if strings.Contains(errorMsg, "required field") || strings.Contains(errorMsg, "missing") {
		// Extract field from error message
		field := app.extractFieldFromError(errorMsg)
		requiredFlags := app.getRequiredFlagsForCommand(commandName)
		examples := app.getExamplesForCommand(commandName)
		
		return help.NewErrorContext().
			Type(help.ErrorTypeMissingRequired).
			Command(commandName).
			Flag(field).
			RequiredFlags(requiredFlags).
			Examples(examples).
			Build()
	}
	
	if strings.Contains(errorMsg, "validation failed") {
		examples := app.getExamplesForCommand(commandName)
		
		return help.NewErrorContext().
			Type(help.ErrorTypeValidation).
			Command(commandName).
			Examples(examples).
			Build()
	}
	
	// Default to generic error
	return help.NewErrorContext().
		Type(help.ErrorTypeGeneric).
		Command(commandName).
		Build()
}

// getAllCommandNames returns all available command names
func (app *Application) getAllCommandNames() []string {
	var commands []string
	for name := range app.registry.ListCommands() {
		commands = append(commands, name)
	}
	return commands
}

// getAllFlagsForCommand returns all flags for a command
func (app *Application) getAllFlagsForCommand(commandName string) []string {
	// This would need to be implemented based on command metadata
	// For now, return common flags
	return []string{"--help", "--verbose", "--debug", "--config"}
}

// getRequiredFlagsForCommand returns required flags for a command
func (app *Application) getRequiredFlagsForCommand(commandName string) []string {
	// This would analyze the command's config struct
	// For now, return placeholder
	return []string{}
}

// getExamplesForCommand returns usage examples for a command
func (app *Application) getExamplesForCommand(commandName string) []string {
	return []string{
		fmt.Sprintf("%s %s [options]", app.config.Name, commandName),
		fmt.Sprintf("%s help %s", app.config.Name, commandName),
	}
}

// extractFlagFromError extracts flag name from error message
func (app *Application) extractFlagFromError(errorMsg string) string {
	// Simple extraction - could be improved with regex
	if strings.Contains(errorMsg, "--") {
		parts := strings.Split(errorMsg, "--")
		if len(parts) > 1 {
			flag := strings.Fields(parts[1])[0]
			return "--" + flag
		}
	}
	if strings.Contains(errorMsg, "-") {
		parts := strings.Split(errorMsg, "-")
		if len(parts) > 1 {
			flag := strings.Fields(parts[1])[0]
			return "-" + flag
		}
	}
	return "unknown"
}

// extractFieldFromError extracts field name from error message
func (app *Application) extractFieldFromError(errorMsg string) string {
	// Extract field name from validation error messages
	if strings.Contains(errorMsg, "field ") {
		parts := strings.Split(errorMsg, "field ")
		if len(parts) > 1 {
			field := strings.Fields(parts[1])[0]
			return "--" + strings.ToLower(field)
		}
	}
	return "unknown"
}

// loadConfigurationFile loads configuration from file for the command
func (app *Application) loadConfigurationFile(commandName string, args []string) error {
	// Get command descriptor to know the config type
	descriptor, exists := app.registry.GetCommand(commandName)
	if !exists {
		return nil // Command doesn't exist, skip config loading
	}
	
	// Create config loader
	configFileName := app.config.ConfigFile
	if configFileName == "" {
		configFileName = app.config.Name // Use app name as default
	}
	
	searchPaths := app.config.ConfigPaths
	if len(searchPaths) == 0 {
		// Default search paths
		searchPaths = []string{
			".",
			"$HOME/.config/" + app.config.Name,
			"/etc/" + app.config.Name,
		}
	}
	
	loader := configfile.NewLoader(configFileName, searchPaths...)
	
	// Create instance of config struct for this command
	configType := descriptor.GetConfigType()
	configPtr := reflect.New(configType)
	config := configPtr.Interface()
	
	// Load configuration from file
	if err := loader.Load(config); err != nil {
		return err
	}
	
	// TODO: Merge with command line arguments
	// For now, we just loaded the config - in a full implementation,
	// we'd need to merge this with CLI args where CLI takes precedence
	
	return nil
}

// GenerateConfigFile generates an example configuration file
func (app *Application) GenerateConfigFile(commandName string, format string) ([]byte, error) {
	descriptor, exists := app.registry.GetCommand(commandName)
	if !exists {
		return nil, fmt.Errorf("command %s not found", commandName)
	}
	
	generator := configfile.NewConfigGenerator()
	configType := descriptor.GetConfigType()
	
	switch strings.ToLower(format) {
	case "yaml", "yml":
		return generator.GenerateYAML(configType)
	case "json":
		return generator.GenerateJSON(configType)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// isMissingRequiredFieldError checks if error is about missing required fields
func (app *Application) isMissingRequiredFieldError(err error) bool {
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "required field") || 
		   strings.Contains(errorMsg, "missing") ||
		   strings.Contains(errorMsg, "validation failed")
}

// handleInteractivePrompt handles interactive prompting for missing fields
func (app *Application) handleInteractivePrompt(ctx context.Context, commandName string, args []string, originalErr error) error {
	// Get command descriptor
	descriptor, exists := app.registry.GetCommand(commandName)
	if !exists {
		return fmt.Errorf("command not found: %s", commandName)
	}
	
	// Create instance of config struct for this command
	configType := descriptor.GetConfigType()
	configPtr := reflect.New(configType)
	config := configPtr.Interface()
	
	// Load any existing configuration from files
	if app.config.AutoLoadConfig {
		if err := app.loadConfigIntoStruct(commandName, config); err != nil {
			// Non-fatal, continue with prompting
		}
	}
	
	// Parse existing command line arguments to get partial config
	if err := app.parseArgsIntoStruct(commandName, args, config); err != nil {
		// Non-fatal, continue with prompting
	}
	
	// Show friendly message about interactive mode
	fmt.Println()
	fmt.Printf("🤖 Interactive mode enabled. I'll help you provide the missing information.\n")
	fmt.Printf("   Press Ctrl+C to cancel at any time.\n")
	fmt.Println()
	
	// Prompt for missing required fields
	if err := app.prompter.PromptMissing(config); err != nil {
		return fmt.Errorf("interactive prompting failed: %w", err)
	}
	
	fmt.Println()
	fmt.Printf("✅ All required information collected!\n")
	fmt.Println()
	
	// Store the filled config somewhere the executor can access it
	// This is a simplified approach - in practice, you'd need to integrate this
	// more deeply with the argument parsing system
	
	return nil
}

// loadConfigIntoStruct loads configuration file into struct
func (app *Application) loadConfigIntoStruct(commandName string, config any) error {
	configFileName := app.config.ConfigFile
	if configFileName == "" {
		configFileName = app.config.Name
	}
	
	searchPaths := app.config.ConfigPaths
	if len(searchPaths) == 0 {
		searchPaths = []string{
			".",
			"$HOME/.config/" + app.config.Name,
			"/etc/" + app.config.Name,
		}
	}
	
	loader := configfile.NewLoader(configFileName, searchPaths...)
	return loader.Load(config)
}

// parseArgsIntoStruct parses command line arguments into struct
func (app *Application) parseArgsIntoStruct(commandName string, args []string, config any) error {
	// This would use the enhanced parser to populate the struct with CLI args
	// For now, this is a placeholder
	return nil
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