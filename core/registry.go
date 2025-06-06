package core

import (
	"context"
	"fmt"
	"reflect"
)

// CommandBase provides a base implementation for commands
type CommandBase[T any] struct {
	name        string
	description string
	runner      func(ctx context.Context, config T) error
}

// NewCommand creates a new generic command
func NewCommand[T any](name, description string, runner func(ctx context.Context, config T) error) *CommandBase[T] {
	return &CommandBase[T]{
		name:        name,
		description: description,
		runner:      runner,
	}
}

// Name returns the command name
func (c *CommandBase[T]) Name() string {
	return c.name
}

// Description returns the command description
func (c *CommandBase[T]) Description() string {
	return c.description
}

// Run executes the command
func (c *CommandBase[T]) Run(ctx context.Context, config T) error {
	return c.runner(ctx, config)
}

// GetConfigType returns the reflect.Type for the config struct
func (c *CommandBase[T]) GetConfigType() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

// Registry manages command registration with type safety
type Registry struct {
	commands map[string]*commandDescriptor
}

type commandDescriptor struct {
	instance   any
	configType reflect.Type
	name       string
	desc       string
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*commandDescriptor),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd any) error {
	// Check if it's a CommandBase
	if baseCmd, ok := cmd.(interface{ GetConfigType() reflect.Type }); ok {
		return r.registerBaseCommand(baseCmd)
	}
	
	// Try to extract from generic interface
	return r.registerGenericCommand(cmd)
}

func (r *Registry) registerBaseCommand(cmd interface{ GetConfigType() reflect.Type }) error {
	configType := cmd.GetConfigType()
	
	// Get name and description through interface
	nameGetter, hasName := cmd.(interface{ Name() string })
	descGetter, hasDesc := cmd.(interface{ Description() string })
	
	if !hasName || !hasDesc {
		return fmt.Errorf("command must implement Name() and Description() methods")
	}
	
	name := nameGetter.Name()
	desc := descGetter.Description()
	
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}
	
	r.commands[name] = &commandDescriptor{
		instance:   cmd,
		configType: configType,
		name:       name,
		desc:       desc,
	}
	
	return nil
}

func (r *Registry) registerGenericCommand(cmd any) error {
	cmdValue := reflect.ValueOf(cmd)
	
	// Verify it has the required methods
	nameMethod := cmdValue.MethodByName("Name")
	descMethod := cmdValue.MethodByName("Description")
	runMethod := cmdValue.MethodByName("Run")
	
	if !nameMethod.IsValid() || !descMethod.IsValid() || !runMethod.IsValid() {
		return fmt.Errorf("command must implement Name(), Description(), and Run() methods")
	}
	
	// Extract config type from Run method signature
	runType := runMethod.Type()
	if runType.NumIn() != 2 { // context, config (receiver is already bound)
		return fmt.Errorf("Run method must have signature: Run(context.Context, T) error")
	}
	
	configType := runType.In(1) // Second parameter is the config
	
	// Get name and description
	nameResult := nameMethod.Call(nil)
	descResult := descMethod.Call(nil)
	
	name := nameResult[0].String()
	desc := descResult[0].String()
	
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}
	
	r.commands[name] = &commandDescriptor{
		instance:   cmd,
		configType: configType,
		name:       name,
		desc:       desc,
	}
	
	return nil
}

// GetCommand returns a command descriptor by name
func (r *Registry) GetCommand(name string) (*commandDescriptor, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// ListCommands returns all registered commands
func (r *Registry) ListCommands() map[string]*commandDescriptor {
	result := make(map[string]*commandDescriptor)
	for k, v := range r.commands {
		result[k] = v
	}
	return result
}

// Execute runs a command with the given arguments
func (r *Registry) Execute(ctx context.Context, name string, config any) error {
	descriptor, exists := r.commands[name]
	if !exists {
		return fmt.Errorf("command not found: %s", name)
	}
	
	// Call the Run method using reflection
	cmdValue := reflect.ValueOf(descriptor.instance)
	runMethod := cmdValue.MethodByName("Run")
	
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}
	
	results := runMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		configValue,
	})
	
	if len(results) > 0 && !results[0].IsNil() {
		return results[0].Interface().(error)
	}
	
	return nil
}

// GetConfigType returns the config type for a command
func (d *commandDescriptor) GetConfigType() reflect.Type {
	return d.configType
}

// GetName returns the command name
func (d *commandDescriptor) GetName() string {
	return d.name
}

// GetDescription returns the command description
func (d *commandDescriptor) GetDescription() string {
	return d.desc
}

// GetInstance returns the command instance
func (d *commandDescriptor) GetInstance() any {
	return d.instance
}