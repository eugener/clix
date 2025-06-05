package core

import (
	"context"
	"fmt"
	"reflect"
)

// cli implements the CLI interface
type cli struct {
	name     string
	version  string
	commands map[string]commandEntry
	parser   Parser
}

// commandEntry holds command metadata and execution info
type commandEntry struct {
	cmd        any
	configType reflect.Type
	name       string
	desc       string
}

// NewCLI creates a new CLI instance
func NewCLI(name, version string, parser Parser) CLI {
	return &cli{
		name:     name,
		version:  version,
		commands: make(map[string]commandEntry),
		parser:   parser,
	}
}

// Register adds a command to the CLI
func (c *cli) Register(cmd any) error {
	// Use reflection to get the command interface and config type
	cmdValue := reflect.ValueOf(cmd)
	cmdType := cmdValue.Type()
	
	// Verify it implements Command[T] interface
	if !cmdType.Implements(reflect.TypeOf((*interface{})(nil)).Elem()) {
		return fmt.Errorf("command must implement Command[T] interface")
	}
	
	// Extract the generic type parameter T
	configType, err := extractConfigType(cmdType)
	if err != nil {
		return fmt.Errorf("failed to extract config type: %w", err)
	}
	
	// Get command name and description
	nameMethod := cmdValue.MethodByName("Name")
	descMethod := cmdValue.MethodByName("Description")
	
	if !nameMethod.IsValid() || !descMethod.IsValid() {
		return fmt.Errorf("command must implement Name() and Description() methods")
	}
	
	nameResult := nameMethod.Call(nil)
	descResult := descMethod.Call(nil)
	
	name := nameResult[0].String()
	desc := descResult[0].String()
	
	c.commands[name] = commandEntry{
		cmd:        cmd,
		configType: configType,
		name:       name,
		desc:       desc,
	}
	
	return nil
}

// Execute parses arguments and runs the appropriate command
func (c *cli) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	
	commandName := args[0]
	entry, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command: %s", commandName)
	}
	
	// Create a new instance of the config type
	configPtr := reflect.New(entry.configType)
	config := configPtr.Interface()
	
	// Parse remaining arguments into the config struct
	if err := c.parser.Parse(args[1:], config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	
	// Validate the config
	if err := c.parser.ValidateStruct(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Execute the command
	cmdValue := reflect.ValueOf(entry.cmd)
	runMethod := cmdValue.MethodByName("Run")
	
	results := runMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		configPtr.Elem(),
	})
	
	if len(results) > 0 && !results[0].IsNil() {
		return results[0].Interface().(error)
	}
	
	return nil
}

// Help returns help text for the CLI or a specific command
func (c *cli) Help(command string) string {
	if command == "" {
		return c.generateMainHelp()
	}
	
	entry, exists := c.commands[command]
	if !exists {
		return fmt.Sprintf("Unknown command: %s", command)
	}
	
	return c.generateCommandHelp(entry)
}

func (c *cli) generateMainHelp() string {
	help := fmt.Sprintf("%s v%s\n\nCommands:\n", c.name, c.version)
	for _, entry := range c.commands {
		help += fmt.Sprintf("  %-12s %s\n", entry.name, entry.desc)
	}
	return help
}

func (c *cli) generateCommandHelp(entry commandEntry) string {
	return fmt.Sprintf("%s - %s\n\nUsage: %s [options]\n", 
		entry.name, entry.desc, entry.name)
}

// extractConfigType extracts the type parameter T from Command[T]
func extractConfigType(cmdType reflect.Type) (reflect.Type, error) {
	// This is a simplified implementation
	// In practice, you'd need more sophisticated reflection to extract
	// the generic type parameter from the Command[T] interface
	
	// For now, assume the command has a method that returns the config type
	// This would be improved with proper generic type extraction
	
	// Return a placeholder - this needs proper implementation
	return reflect.TypeOf(struct{}{}), nil
}