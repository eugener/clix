package core

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

// Test command for registry tests
type RegistryTestConfig struct {
	Value string `posix:"v,value,Test value"`
}

type RegistryTestCommand struct {
	executed bool
	config   RegistryTestConfig
}

func (c *RegistryTestCommand) Name() string {
	return "test"
}

func (c *RegistryTestCommand) Description() string {
	return "Test command"
}

func (c *RegistryTestCommand) Run(ctx context.Context, config RegistryTestConfig) error {
	c.executed = true
	c.config = config
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(registry.commands) != 0 {
		t.Error("Registry should start with no commands")
	}
}

func TestRegistry_RegisterCommand(t *testing.T) {
	registry := NewRegistry()
	cmd := &RegistryTestCommand{}
	
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	
	commands := registry.ListCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
	
	registeredCmd, exists := registry.GetCommand("test")
	if !exists {
		t.Error("Command not found in registry")
	}
	if registeredCmd.GetName() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", registeredCmd.GetName())
	}
	if registeredCmd.GetDescription() != "Test command" {
		t.Errorf("Expected description 'Test command', got '%s'", registeredCmd.GetDescription())
	}
}

func TestRegistry_RegisterCommand_Duplicate(t *testing.T) {
	registry := NewRegistry()
	cmd1 := &RegistryTestCommand{}
	cmd2 := &RegistryTestCommand{}
	
	err := registry.Register(cmd1)
	if err != nil {
		t.Fatalf("First Register failed: %v", err)
	}
	
	err = registry.Register(cmd2)
	if err == nil {
		t.Error("Expected error for duplicate command registration")
	}
}

func TestRegistry_GetCommand(t *testing.T) {
	registry := NewRegistry()
	cmd := &RegistryTestCommand{}
	
	registry.Register(cmd)
	
	retrievedCmd, exists := registry.GetCommand("test")
	if !exists {
		t.Error("GetCommand returned false for existing command")
	}
	if retrievedCmd.GetName() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", retrievedCmd.GetName())
	}
	
	_, exists = registry.GetCommand("nonexistent")
	if exists {
		t.Error("GetCommand returned true for non-existent command")
	}
}

func TestRegistry_ListCommands(t *testing.T) {
	registry := NewRegistry()
	cmd1 := NewCommand("test1", "Test command 1", func(ctx context.Context, config RegistryTestConfig) error {
		return nil
	})
	cmd2 := NewCommand("test2", "Test command 2", func(ctx context.Context, config RegistryTestConfig) error {
		return nil
	})
	
	registry.Register(cmd1)
	registry.Register(cmd2)
	
	commands := registry.ListCommands()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		names = append(names, cmd.GetName())
	}
	
	// Check both commands exist
	found1, found2 := false, false
	for _, name := range names {
		if name == "test1" {
			found1 = true
		} else if name == "test2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("Expected to find both test1 and test2 commands, got names: %v", names)
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := NewRegistry()
	executed := false
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config RegistryTestConfig) error {
		executed = true
		return nil
	})
	
	registry.Register(cmd)
	
	err := registry.Execute(context.Background(), "test", RegistryTestConfig{Value: "testvalue"})
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	
	if !executed {
		t.Error("Command was not executed")
	}
}

func TestRegistry_Execute_NotFound(t *testing.T) {
	registry := NewRegistry()
	
	err := registry.Execute(context.Background(), "nonexistent", RegistryTestConfig{})
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
	
	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("Expected 'command not found' error, got: %v", err)
	}
}

func TestCommandDescriptor_Methods(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config RegistryTestConfig) error {
		return nil
	})
	
	registry.Register(cmd)
	
	descriptor, exists := registry.GetCommand("test")
	if !exists {
		t.Fatal("Command not found")
	}
	
	if descriptor.GetName() != "test" {
		t.Errorf("Expected name 'test', got '%s'", descriptor.GetName())
	}
	
	if descriptor.GetDescription() != "Test command" {
		t.Errorf("Expected description 'Test command', got '%s'", descriptor.GetDescription())
	}
	
	configType := descriptor.GetConfigType()
	expectedType := reflect.TypeOf(RegistryTestConfig{})
	if configType != expectedType {
		t.Errorf("Expected config type %v, got %v", expectedType, configType)
	}
}