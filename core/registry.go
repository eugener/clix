package core

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// commandBase provides a unified implementation for commands with optional nesting
type commandBase[T any] struct {
	name        string
	description string
	aliases     []string
	runner      func(ctx context.Context, config T) error
	
	// Optional nesting support
	subcommands map[string]Command
	parent      Command
	mu          sync.RWMutex
}

// NewCommand creates a new generic command
// If runner is nil, this creates a parent command (command with subcommands but no direct execution)
func NewCommand[T any](name, description string, runner func(ctx context.Context, config T) error) Command {
	return &commandBase[T]{
		name:        name,
		description: description,
		runner:      runner,
		subcommands: make(map[string]Command),
	}
}

// NewCommandWithAliases creates a new generic command with aliases
func NewCommandWithAliases[T any](name, description string, aliases []string, runner func(ctx context.Context, config T) error) Command {
	return &commandBase[T]{
		name:        name,
		description: description,
		aliases:     aliases,
		runner:      runner,
		subcommands: make(map[string]Command),
	}
}


// Run executes the command with typed config (for direct usage)
func (c *commandBase[T]) Run(ctx context.Context, config T) error {
	if c.runner == nil {
		return fmt.Errorf("command %s has subcommands and cannot be executed directly", c.name)
	}
	return c.runner(ctx, config)
}


// HasSubcommands returns true if this command has subcommands
func (c *commandBase[T]) HasSubcommands() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subcommands) > 0 || c.runner == nil
}

// AddSubcommand adds a subcommand to this command
func (c *commandBase[T]) AddSubcommand(cmd Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	name := cmd.GetName()
	
	// Check if name already exists
	if _, exists := c.subcommands[name]; exists {
		return fmt.Errorf("subcommand %s already exists in %s", name, c.name)
	}

	// Check aliases for conflicts
	for _, alias := range cmd.GetAliases() {
		if _, exists := c.subcommands[alias]; exists {
			return fmt.Errorf("subcommand alias %s conflicts with existing subcommand in %s", alias, c.name)
		}
	}

	// Add main command
	c.subcommands[name] = cmd
	
	// Add all aliases
	for _, alias := range cmd.GetAliases() {
		c.subcommands[alias] = cmd
	}

	// Set parent relationship
	cmd.SetParent(c)

	return nil
}

// GetSubcommand retrieves a subcommand by name (including aliases)
func (c *commandBase[T]) GetSubcommand(name string) (Command, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cmd, exists := c.subcommands[name]
	return cmd, exists
}

// ListSubcommands returns all subcommands
func (c *commandBase[T]) ListSubcommands() map[string]Command {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]Command)
	for name, cmd := range c.subcommands {
		// Only include main command names, not aliases
		if cmd.GetName() == name {
			result[name] = cmd
		}
	}
	return result
}

// GetPath returns the full command path
func (c *commandBase[T]) GetPath() []string {
	if c.parent == nil {
		return []string{c.name}
	}
	parentPath := c.parent.GetPath()
	return append(parentPath, c.name)
}

// SetParent sets the parent command
func (c *commandBase[T]) SetParent(parent Command) {
	c.parent = parent
}

// GetParent returns the parent command
func (c *commandBase[T]) GetParent() Command {
	return c.parent
}

// Command interface implementation

// GetName returns the command name
func (c *commandBase[T]) GetName() string {
	return c.name
}

// GetDescription returns the command description
func (c *commandBase[T]) GetDescription() string {
	return c.description
}

// GetAliases returns the command aliases
func (c *commandBase[T]) GetAliases() []string {
	return c.aliases
}

// GetConfigType returns the reflect.Type for the config struct
func (c *commandBase[T]) GetConfigType() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

// Execute runs the command with type-erased config
func (c *commandBase[T]) Execute(ctx context.Context, config any) error {
	if c.runner == nil {
		return fmt.Errorf("command %s has subcommands and cannot be executed directly", c.name)
	}

	// Type assertion for config with better error handling
	typedConfig, ok := config.(T)
	if !ok {
		// Try to handle pointer/value conversions
		var zero T
		zeroType := reflect.TypeOf(zero)
		configValue := reflect.ValueOf(config)
		configType := configValue.Type()
		
		// Handle pointer to value conversion
		if configType.Kind() == reflect.Ptr && !configValue.IsNil() {
			configValue = configValue.Elem()
			configType = configValue.Type()
		}
		
		// If both are empty structs (struct{} types), allow it
		if zeroType.Kind() == reflect.Struct && zeroType.NumField() == 0 && 
		   configType.Kind() == reflect.Struct && configType.NumField() == 0 {
			// Use the zero value for empty struct
			typedConfig = zero
		} else if configType == zeroType {
			// Same type, try to convert
			typedConfig = configValue.Interface().(T)
		} else {
			return fmt.Errorf("invalid config type for command %s: expected %v, got %v", 
				c.name, zeroType, reflect.TypeOf(config))
		}
	}

	return c.runner(ctx, typedConfig)
}

// GetInstance returns the command instance
func (c *commandBase[T]) GetInstance() any {
	return c
}

// LegacyCommandAdapter adapts old-style commands to the new Command interface
type LegacyCommandAdapter struct {
	name        string
	description string
	aliases     []string
	configType  reflect.Type
	instance    any
	runMethod   reflect.Value
}

// NewLegacyCommandAdapter creates an adapter for legacy commands
func NewLegacyCommandAdapter(cmd any) (*LegacyCommandAdapter, error) {
	cmdValue := reflect.ValueOf(cmd)
	
	// Verify it has the required methods
	nameMethod := cmdValue.MethodByName("Name")
	descMethod := cmdValue.MethodByName("Description")
	runMethod := cmdValue.MethodByName("Run")
	
	if !nameMethod.IsValid() || !descMethod.IsValid() || !runMethod.IsValid() {
		return nil, fmt.Errorf("command must implement Name(), Description(), and Run() methods")
	}
	
	// Extract config type from Run method signature
	runType := runMethod.Type()
	if runType.NumIn() != 2 { // context, config (receiver is already bound)
		return nil, fmt.Errorf("Run method must have signature: Run(context.Context, T) error")
	}
	
	configType := runType.In(1) // Second parameter is the config
	
	// Get name and description
	nameResult := nameMethod.Call(nil)
	descResult := descMethod.Call(nil)
	
	name := nameResult[0].String()
	desc := descResult[0].String()
	
	// Check for aliases (optional)
	var aliases []string
	if aliasMethod := cmdValue.MethodByName("Aliases"); aliasMethod.IsValid() {
		aliasResult := aliasMethod.Call(nil)
		if aliasResult[0].Kind() == reflect.Slice {
			for i := 0; i < aliasResult[0].Len(); i++ {
				aliases = append(aliases, aliasResult[0].Index(i).String())
			}
		}
	}
	
	return &LegacyCommandAdapter{
		name:        name,
		description: desc,
		aliases:     aliases,
		configType:  configType,
		instance:    cmd,
		runMethod:   runMethod,
	}, nil
}

// Command interface implementation for LegacyCommandAdapter

func (l *LegacyCommandAdapter) GetName() string {
	return l.name
}

func (l *LegacyCommandAdapter) GetDescription() string {
	return l.description
}

func (l *LegacyCommandAdapter) GetAliases() []string {
	return l.aliases
}

func (l *LegacyCommandAdapter) GetConfigType() reflect.Type {
	return l.configType
}

func (l *LegacyCommandAdapter) Execute(ctx context.Context, config any) error {
	// Handle config value type conversion
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}
	
	// Ensure the config type matches what the method expects
	expectedType := l.configType
	if expectedType.Kind() == reflect.Ptr {
		expectedType = expectedType.Elem()
	}
	
	if configValue.Type() != expectedType {
		// Try to convert the config to the expected type
		if configValue.Type().ConvertibleTo(expectedType) {
			configValue = configValue.Convert(expectedType)
		} else {
			return fmt.Errorf("invalid config type for command %s: expected %v, got %v", 
				l.name, expectedType, configValue.Type())
		}
	}
	
	// Call the Run method using reflection
	results := l.runMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		configValue,
	})
	
	if len(results) > 0 && !results[0].IsNil() {
		return results[0].Interface().(error)
	}
	
	return nil
}

func (l *LegacyCommandAdapter) HasSubcommands() bool {
	return false // Legacy commands cannot have subcommands
}

func (l *LegacyCommandAdapter) AddSubcommand(cmd Command) error {
	return fmt.Errorf("legacy command %s does not support subcommands", l.name)
}

func (l *LegacyCommandAdapter) GetSubcommand(name string) (Command, bool) {
	return nil, false // Legacy commands have no subcommands
}

func (l *LegacyCommandAdapter) ListSubcommands() map[string]Command {
	return make(map[string]Command) // Legacy commands have no subcommands
}

func (l *LegacyCommandAdapter) GetPath() []string {
	return []string{l.name} // Legacy commands are always top-level
}

func (l *LegacyCommandAdapter) SetParent(parent Command) {
	// Legacy commands don't support parent relationships
}

func (l *LegacyCommandAdapter) GetParent() Command {
	return nil // Legacy commands have no parent
}

func (l *LegacyCommandAdapter) GetInstance() any {
	return l.instance
}

// Registry manages command registration with type safety and supports nested commands
type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd any) error {
	// Check if it implements Command directly
	if descriptor, ok := cmd.(Command); ok {
		return r.registerDescriptor(descriptor)
	}

	// Check if it's a commandBase (which implements Command)
	if baseCmd, ok := cmd.(interface{ GetConfigType() reflect.Type }); ok {
		// Convert to Command
		if descriptor, ok := baseCmd.(Command); ok {
			return r.registerDescriptor(descriptor)
		}
		return fmt.Errorf("command implements GetConfigType but not Command")
	}

	// Handle legacy commands that implement the old interface
	return r.registerLegacyCommand(cmd)
}

// registerDescriptor registers a command that implements Command
func (r *Registry) registerDescriptor(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cmd.GetName()

	// Check if command already exists
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}

	// Check aliases for conflicts
	for _, alias := range cmd.GetAliases() {
		if _, exists := r.commands[alias]; exists {
			return fmt.Errorf("command alias %s already registered", alias)
		}
	}

	// Register main command
	r.commands[name] = cmd

	// Register all aliases
	for _, alias := range cmd.GetAliases() {
		r.commands[alias] = cmd
	}

	return nil
}

// registerLegacyCommand wraps a legacy command in a Command adapter
func (r *Registry) registerLegacyCommand(cmd any) error {
	// Create a legacy command adapter
	adapter, err := NewLegacyCommandAdapter(cmd)
	if err != nil {
		return err
	}
	
	return r.registerDescriptor(adapter)
}

// GetCommand returns a command descriptor by name
func (r *Registry) GetCommand(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	return cmd, exists
}

// ListCommands returns all registered commands (top-level only)
func (r *Registry) ListCommands() map[string]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]Command)
	for name, cmd := range r.commands {
		// Only include main command names, not aliases
		if cmd.GetName() == name {
			result[name] = cmd
		}
	}
	return result
}

// Execute runs a command with the given arguments
func (r *Registry) Execute(ctx context.Context, name string, config any) error {
	r.mu.RLock()
	cmd, exists := r.commands[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("command not found: %s", name)
	}

	return cmd.Execute(ctx, config)
}

// ResolveCommand resolves a command from a path (e.g., ["docker", "container", "ls"])
// Returns the command descriptor, the resolved path, and remaining arguments
func (r *Registry) ResolveCommand(path []string) (Command, []string, []string, error) {
	if len(path) == 0 {
		return nil, nil, path, fmt.Errorf("empty command path")
	}

	// Get the top-level command
	topLevelCmd, exists := r.GetCommand(path[0])
	if !exists {
		return nil, nil, path, fmt.Errorf("command not found: %s", path[0])
	}

	// If there's only one element, return the top-level command
	if len(path) == 1 {
		return topLevelCmd, path, nil, nil
	}

	// Navigate through nested commands
	return r.resolveNestedCommand(topLevelCmd, path, 1)
}

// resolveNestedCommand recursively resolves nested commands
func (r *Registry) resolveNestedCommand(cmd Command, path []string, index int) (Command, []string, []string, error) {
	if index >= len(path) {
		return cmd, path[:index], path[index:], nil
	}

	// Check if the current command has the next subcommand
	nextName := path[index]
	subCmd, exists := cmd.GetSubcommand(nextName)
	if !exists {
		// If this command has no subcommands, it's the final command
		if !cmd.HasSubcommands() {
			return cmd, path[:index], path[index:], nil
		}
		return nil, path[:index], path[index:], fmt.Errorf("subcommand not found: %s", nextName)
	}

	// Continue resolving with the subcommand
	return r.resolveNestedCommand(subCmd, path, index+1)
}

// ExecuteNested runs a nested command with the given path and arguments
func (r *Registry) ExecuteNested(ctx context.Context, path []string, config any) error {
	cmd, resolvedPath, remainingArgs, err := r.ResolveCommand(path)
	if err != nil {
		return err
	}

	// Log resolved path for debugging (can be removed in production)
	_ = resolvedPath
	_ = remainingArgs

	return cmd.Execute(ctx, config)
}

// GetAllCommands returns a flat map of all commands including nested ones
func (r *Registry) GetAllCommands() map[string]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]Command)
	
	// Add top-level commands
	for name, cmd := range r.commands {
		if cmd.GetName() == name { // Only main names, not aliases
			result[name] = cmd
			r.addNestedCommands(cmd, name, result)
		}
	}
	
	return result
}

// addNestedCommands recursively adds nested commands to the result map
func (r *Registry) addNestedCommands(cmd Command, prefix string, result map[string]Command) {
	for name, subCmd := range cmd.ListSubcommands() {
		fullName := prefix + "." + name
		result[fullName] = subCmd
		r.addNestedCommands(subCmd, fullName, result)
	}
}

// GetCommandByPath returns a command by its full path (e.g., "docker.container.ls")
func (r *Registry) GetCommandByPath(path string) (Command, bool) {
	parts := strings.Split(path, ".")
	cmd, _, _, err := r.ResolveCommand(parts)
	return cmd, err == nil
}