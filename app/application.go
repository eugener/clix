package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/eugener/clix/config"
	"github.com/eugener/clix/core"
	"github.com/eugener/clix/internal/bind"
	"github.com/eugener/clix/internal/configfile"
	"github.com/eugener/clix/internal/help"
	"github.com/eugener/clix/internal/interactive"
)

// Application represents a complete CLI application
type Application struct {
	config      *config.CLIConfig
	registry    *core.Registry
	executor    *core.Executor
	helpGen     *help.Generator
	errorFormat *help.ErrorFormatter
	suggestions *help.SuggestionEngine
	prompter    *interactive.SmartPrompter
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

	// Handle version request first (simple case)
	if app.isVersionRequest(args[0]) {
		fmt.Printf("%s version %s\n", app.config.Name, app.config.Version)
		if app.config.Author != "" {
			fmt.Printf("Author: %s\n", app.config.Author)
		}
		return 0
	}

	// Resolve the command path by separating commands from arguments
	resolvedCmd, resolvedPath, commandArgs := app.resolveCommandPath(args)
	if resolvedCmd == nil {
		// Command resolution failed - unknown command
		commandName := args[0]
		allCommands := app.getAllCommandNames()
		suggestions := app.suggestions.SuggestCommands(commandName, allCommands)

		errorCtx := help.NewErrorContext().
			Type(help.ErrorTypeUnknownCommand).
			Command(commandName).
			Suggestions(suggestions).
			AllCommands(allCommands).
			Build()

		formattedError := app.errorFormat.FormatError(fmt.Errorf("command not found: %s", commandName), errorCtx)
		fmt.Fprint(os.Stderr, formattedError)
		return 1
	}

	// Check if this is a help request for the resolved command
	if len(commandArgs) > 0 && app.isHelpRequest(commandArgs[0]) {
		if resolvedCmd.HasSubcommands() {
			// Show parent command help
			parentInfo := app.buildParentCommandInfo(resolvedCmd)
			fmt.Print(app.helpGen.GenerateParentCommandHelp(parentInfo))
			return 0
		} else {
			// Show command help
			return app.handleCommandHelp(resolvedCmd, resolvedPath)
		}
	}

	// Check if user is trying to execute a parent command directly (without subcommands)
	// If so, show error but also provide helpful guidance
	if resolvedCmd.HasSubcommands() && len(commandArgs) == 0 {
		commandPath := strings.Join(resolvedPath, " ")
		
		// Create error context for proper formatting
		errorCtx := help.NewErrorContext().
			Type(help.ErrorTypeGeneric).
			Command(commandPath).
			Build()
		
		// Format the error with colors
		errorMsg := fmt.Sprintf("command %s has subcommands and cannot be executed directly", commandPath)
		formattedError := app.errorFormat.FormatError(fmt.Errorf(errorMsg), errorCtx)
		
		// Show formatted error
		fmt.Fprint(os.Stderr, formattedError)
		
		// Add helpful guidance
		parentInfo := app.buildParentCommandInfo(resolvedCmd)
		helpText := app.helpGen.GenerateParentCommandHelp(parentInfo)
		fmt.Fprint(os.Stderr, helpText)
		return 1
	}

	// Construct command name from resolved path for execution
	commandName := strings.Join(resolvedPath, " ")

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
	var baseConfig any
	if app.config.AutoLoadConfig {
		config, err := app.loadConfigurationFile(commandName, commandArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load configuration file: %v\n", err)
		} else {
			baseConfig = config
		}
	}

	// Execute the resolved command directly
	if err := app.executeResolvedCommand(ctx, resolvedCmd, commandName, commandArgs, baseConfig); err != nil {
		// Check if this is a missing required field error and interactive mode is enabled
		if app.config.InteractiveMode && app.isMissingRequiredFieldError(err) {
			if interactiveErr := app.handleInteractivePrompt(ctx, commandName, commandArgs, err); interactiveErr == nil {
				// Successfully prompted and got values, try again
				if retryErr := app.executeResolvedCommand(ctx, resolvedCmd, commandName, commandArgs, baseConfig); retryErr == nil {
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
			Aliases:     desc.GetAliases(),
		}
	}

	parentCommands := make(map[string]help.ParentCommandInfo)
	for name, cmd := range app.registry.ListCommands() {
		if cmd.HasSubcommands() {
			parentCommands[name] = app.buildParentCommandInfo(cmd)
		}
	}

	fmt.Print(app.helpGen.GenerateMainHelpWithParentCommands(commands, parentCommands))
}

// buildParentCommandInfo recursively builds help information for a parent command
func (app *Application) buildParentCommandInfo(cmd core.Command) help.ParentCommandInfo {
	commands := make(map[string]help.CommandInfo)
	parentCommands := make(map[string]help.ParentCommandInfo)
	
	// Get subcommands from the command
	for name, subCmd := range cmd.ListSubcommands() {
		if subCmd.HasSubcommands() {
			// This is a nested command
			parentCommands[name] = app.buildParentCommandInfo(subCmd)
		} else {
			// This is a regular command
			commands[name] = help.CommandInfo{
				Name:        subCmd.GetName(),
				Description: subCmd.GetDescription(),
				Aliases:     subCmd.GetAliases(),
				ConfigType:  subCmd.GetConfigType(),
			}
		}
	}

	return help.ParentCommandInfo{
		Name:        cmd.GetName(),
		Description: cmd.GetDescription(),
		Aliases:     cmd.GetAliases(),
		Commands:    commands,
		ParentCommands: parentCommands,
		Path:        cmd.GetPath(),
	}
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
				Aliases:     desc.GetAliases(),
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

// handleParentCommandHelp handles help requests for parent commands
func (app *Application) handleParentCommandHelp(parentName string) int {
	if cmd, exists := app.registry.GetCommand(parentName); exists {
		parentInfo := app.buildParentCommandInfo(cmd)
		fmt.Print(app.helpGen.GenerateParentCommandHelp(parentInfo))
		return 0
	}
	fmt.Fprintf(os.Stderr, "Unknown parent command: %s\n", parentName)
	return 1
}

// handleCommandHelp handles help requests for specific commands
func (app *Application) handleCommandHelp(cmd core.Command, path []string) int {
	commandName := strings.Join(path, " ")
	info := help.CommandInfo{
		Name:        cmd.GetName(),
		Description: cmd.GetDescription(),
		ConfigType:  cmd.GetConfigType(),
		Aliases:     cmd.GetAliases(),
		Examples: []string{
			fmt.Sprintf("%s %s [options]", app.config.Name, commandName),
		},
	}
	helpText, err := app.helpGen.GenerateCommandHelp(commandName, info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help: %v\n", err)
		return 1
	}
	fmt.Print(helpText)
	return 0
}

// resolveCommandPath resolves the command path by separating commands from arguments
// Returns the resolved command, the path that was resolved, and remaining arguments
func (app *Application) resolveCommandPath(args []string) (core.Command, []string, []string) {
	if len(args) == 0 {
		return nil, nil, nil
	}

	// Start with the first argument as the top-level command
	topLevelCmd, exists := app.registry.GetCommand(args[0])
	if !exists {
		return nil, nil, args
	}

	// If only one argument or it has no subcommands, return it
	if len(args) == 1 || !topLevelCmd.HasSubcommands() {
		return topLevelCmd, []string{args[0]}, args[1:]
	}

	// Try to resolve nested commands
	currentCmd := topLevelCmd
	resolvedPath := []string{args[0]}
	
	for i := 1; i < len(args); i++ {
		arg := args[i]
		
		// If this looks like a flag or option, stop here
		if strings.HasPrefix(arg, "-") {
			return currentCmd, resolvedPath, args[i:]
		}
		
		// Try to find this as a subcommand
		if subCmd, exists := currentCmd.GetSubcommand(arg); exists {
			currentCmd = subCmd
			resolvedPath = append(resolvedPath, arg)
		} else {
			// Not a subcommand, so this and everything after are arguments
			return currentCmd, resolvedPath, args[i:]
		}
	}

	// All arguments were part of the command path
	return currentCmd, resolvedPath, []string{}
}

// executeResolvedCommand executes a resolved command with configuration and middleware
func (app *Application) executeResolvedCommand(ctx context.Context, cmd core.Command, commandName string, args []string, baseConfig any) error {
	// Create execution context
	execCtx := core.NewExecutionContext(ctx, commandName, args)

	// Create config instance
	configType := cmd.GetConfigType()
	configPtr := reflect.New(configType)
	config := configPtr.Interface()

	// Apply base configuration if provided (from config file)
	if baseConfig != nil {
		if err := app.mergeConfigs(config, baseConfig); err != nil {
			return fmt.Errorf("failed to apply base configuration: %w", err)
		}
	}

	// Parse arguments using enhanced parser
	binder := bind.NewBinder("posix")
	parser := core.NewEnhancedParser(binder)
	if err := parser.Parse(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate configuration
	if err := app.validateConfig(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Log execution start
	execCtx.Logger.Info("executing command",
		"command", execCtx.CommandName,
		"args", execCtx.Args,
		"duration_so_far", execCtx.Duration(),
	)

	// Execute the command directly
	return cmd.Execute(execCtx.Context, config)
}

// Helper methods to access executor internals
func (app *Application) mergeConfigs(target, base any) error {
	// This logic is copied from the executor - ideally it should be refactored to a shared utility
	targetValue := reflect.ValueOf(target)
	baseValue := reflect.ValueOf(base)

	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	if baseValue.Kind() == reflect.Ptr {
		baseValue = baseValue.Elem()
	}

	if baseValue.Kind() != reflect.Struct {
		return fmt.Errorf("base must be a struct or pointer to struct")
	}

	targetStruct := targetValue.Elem()
	baseStruct := baseValue

	// Check that both structs have the same type
	if targetStruct.Type() != baseStruct.Type() {
		return fmt.Errorf("target and base configurations must have the same type")
	}

	// Copy non-zero values from base to target where target field is zero
	for i := 0; i < targetStruct.NumField(); i++ {
		targetField := targetStruct.Field(i)
		baseField := baseStruct.Field(i)

		// Skip unexported fields
		if !targetField.CanSet() {
			continue
		}

		// If target field is zero and base field is not zero, copy from base
		if targetField.IsZero() && !baseField.IsZero() {
			if targetField.Type() == baseField.Type() {
				targetField.Set(baseField)
			}
		}
	}

	return nil
}

func (app *Application) validateConfig(config any) error {
	// This logic is copied from the executor - ideally it should be refactored to a shared utility
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	analyzer := bind.NewAnalyzer("posix")
	metadata, err := analyzer.Analyze(configValue.Type())
	if err != nil {
		return err
	}

	// Check required fields
	for _, fieldInfo := range metadata.Fields {
		if !fieldInfo.Required {
			continue
		}

		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			return fmt.Errorf("required field %s is missing", fieldInfo.Name)
		}
	}

	// Check choices validation
	for _, fieldInfo := range metadata.Fields {
		if len(fieldInfo.Choices) == 0 {
			continue
		}

		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			continue
		}

		value := fmt.Sprintf("%v", field.Interface())
		valid := false
		for _, choice := range fieldInfo.Choices {
			if value == choice {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("field %s must be one of: %v", fieldInfo.Name, fieldInfo.Choices)
		}
	}

	return nil
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

	// Error type detection is ordered by specificity and frequency:
	// 1. Unknown commands (most common user error - typos in command names)
	// 2. Command conflicts (development/registration errors - precise matching needed)
	// 3. Configuration errors (broad category - checked before more general patterns)
	// 4. Unknown flags (command-specific errors - require command context)
	// 5. Missing required fields (validation errors - require detailed field analysis)
	// 6. Invalid values (value-specific errors - require field and value context)
	//
	// Ordering rationale:
	// - Most specific patterns first to avoid false positives
	// - User-facing errors before development errors
	// - Command-level errors before field-level errors
	// - Common errors before rare edge cases

	// Unknown commands - most common user error (typos, wrong command names)
	if strings.Contains(errorMsg, "unknown command") || strings.Contains(errorMsg, "command not found") {
		allCommands := app.getAllCommandNames()
		suggestions := app.suggestions.SuggestCommands(commandName, allCommands)

		return help.NewErrorContext().
			Type(help.ErrorTypeUnknownCommand).
			Command(commandName).
			Suggestions(suggestions).
			AllCommands(allCommands).
			Build()
	}

	// Command conflicts - development errors during command registration
	if strings.Contains(errorMsg, "already registered") || strings.Contains(errorMsg, "conflict") {
		allCommands := app.getAllCommandNames()

		return help.NewErrorContext().
			Type(help.ErrorTypeCommandConflict).
			Command(commandName).
			AllCommands(allCommands).
			Build()
	}

	// Configuration errors - file loading, parsing, or validation issues
	if strings.Contains(errorMsg, "config") || strings.Contains(errorMsg, "configuration") {
		examples := []string{
			"# Example YAML configuration",
			"database:",
			"  host: localhost",
			"  port: 5432",
		}

		return help.NewErrorContext().
			Type(help.ErrorTypeConfigurationError).
			Command(commandName).
			Examples(examples).
			Build()
	}

	// Unknown flags - command-specific errors requiring flag suggestions
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

	// Missing required fields - validation errors needing field-specific guidance
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

	// Validation failures - value format or constraint violations
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
func (app *Application) loadConfigurationFile(commandName string, args []string) (any, error) {
	// Get command descriptor to know the config type
	descriptor, exists := app.registry.GetCommand(commandName)
	if !exists {
		return nil, nil // Command doesn't exist, skip config loading
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
		return nil, err
	}

	// Return the loaded config for merging with CLI arguments
	return config, nil
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
			// Non-fatal, continue with prompting - log or handle as needed
			_ = err // explicitly ignore error
		}
	}

	// Parse existing command line arguments to get partial config
	if err := app.parseArgsIntoStruct(commandName, args, config); err != nil {
		// Non-fatal, continue with prompting - log or handle as needed
		_ = err // explicitly ignore error
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
