package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// Test command for executor tests
type ExecutorTestConfig struct {
	Name     string `posix:"n,name,Name parameter,required"`
	Count    int    `posix:"c,count,Count parameter,default=1"`
	Optional string `posix:"o,optional,Optional parameter"`
}

type ExecutorTestCommand struct {
	executed bool
	config   ExecutorTestConfig
	err      error
}

func (c *ExecutorTestCommand) Name() string {
	return "test"
}

func (c *ExecutorTestCommand) Description() string {
	return "Test command"
}

func (c *ExecutorTestCommand) Run(ctx context.Context, config ExecutorTestConfig) error {
	c.executed = true
	c.config = config
	return c.err
}

type ExecutorErrorCommand struct{}

func (c *ExecutorErrorCommand) Name() string {
	return "test"
}

func (c *ExecutorErrorCommand) Description() string {
	return "Test command"
}

func (c *ExecutorErrorCommand) Run(ctx context.Context, config ExecutorTestConfig) error {
	return errors.New("test error")
}

func TestNewExecutor(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)
	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}
}

func TestExecutor_ExecuteCommand_Success(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)
	cmd := &ExecutorTestCommand{}

	// Register the command
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	args := []string{"--name", "testvalue", "--count", "5"}
	err = executor.Execute(context.Background(), "test", args)

	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}

	if !cmd.executed {
		t.Error("Command was not executed")
	}

	if cmd.config.Name != "testvalue" {
		t.Errorf("Expected Name 'testvalue', got '%s'", cmd.config.Name)
	}

	if cmd.config.Count != 5 {
		t.Errorf("Expected Count 5, got %d", cmd.config.Count)
	}
}

func TestExecutor_ExecuteCommand_Error(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)
	cmd := &ExecutorErrorCommand{}

	// Register the command
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	args := []string{"--name", "testvalue"}
	err = executor.Execute(context.Background(), "test", args)

	if err == nil {
		t.Error("Expected error from command execution")
	}

	if err.Error() != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", err.Error())
	}
}

func TestExecutor_ExecuteCommand_MissingRequired(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)
	cmd := &ExecutorTestCommand{}

	// Register the command
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Missing required --name flag
	args := []string{"--count", "5"}
	err = executor.Execute(context.Background(), "test", args)

	if err == nil {
		t.Error("Expected error for missing required flag")
	}
}

func TestExtractConfigType_ValidCommand(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config ExecutorTestConfig) error {
		return nil
	})

	err := registry.Register(cmd)
	if err != nil {
		t.Errorf("Failed to register command: %v", err)
		return
	}

	descriptor, exists := registry.GetCommand("test")
	if !exists {
		t.Error("Command not found after registration")
		return
	}

	configType := descriptor.GetConfigType()
	expectedType := reflect.TypeOf(ExecutorTestConfig{})
	if configType != expectedType {
		t.Errorf("Expected config type %v, got %v", expectedType, configType)
	}
}

func TestExtractConfigType_NoRunMethod(t *testing.T) {
	type InvalidCommand struct{}

	registry := NewRegistry()
	err := registry.Register(&InvalidCommand{})
	if err == nil {
		t.Error("Expected error for command without Run method")
	}
}

type InvalidRunCommand struct{}

func (c *InvalidRunCommand) Name() string {
	return "invalid"
}

func (c *InvalidRunCommand) Description() string {
	return "Invalid command"
}

func (c *InvalidRunCommand) Run() error {
	return nil
}

func TestExtractConfigType_InvalidRunSignature(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(&InvalidRunCommand{})
	if err == nil {
		t.Error("Expected error for invalid Run method signature")
	}
}

func TestExecutor_ExecuteWithConfig(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)
	cmd := &ExecutorTestCommand{}

	// Register the command
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	config := ExecutorTestConfig{
		Name:  "testvalue",
		Count: 3,
	}

	err = executor.ExecuteWithConfig(context.Background(), "test", []string{}, config)

	if err != nil {
		t.Errorf("ExecuteWithConfig failed: %v", err)
	}

	if !cmd.executed {
		t.Error("Command was not executed")
	}

	if cmd.config.Name != "testvalue" {
		t.Errorf("Expected Name 'testvalue', got '%s'", cmd.config.Name)
	}

	if cmd.config.Count != 3 {
		t.Errorf("Expected Count 3, got %d", cmd.config.Count)
	}
}
