package app

import (
	"context"
	"testing"

	"github.com/eugener/clix/config"
	"github.com/eugener/clix/core"
)

// Test command for app tests
type AppTestConfig struct {
	Value string `posix:"v,value,Test value"`
}

type AppTestCommand struct {
	executed bool
	config   AppTestConfig
}

func (c *AppTestCommand) Name() string {
	return "test"
}

func (c *AppTestCommand) Description() string {
	return "Test command"
}

func (c *AppTestCommand) Run(ctx context.Context, config AppTestConfig) error {
	c.executed = true
	c.config = config
	return nil
}

func TestNewApplication(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Name = "test-app"
	cfg.Version = "1.0.0"
	cfg.Apply() // This ensures HelpConfig is properly initialized
	
	app := NewApplication(cfg)
	if app == nil {
		t.Fatal("NewApplication returned nil")
	}
	
	if app.config != cfg {
		t.Error("Application config not set correctly")
	}
	
	if app.registry == nil {
		t.Error("Registry not initialized")
	}
	
	if app.executor == nil {
		t.Error("Executor not initialized")
	}
}

func TestNewApplicationWithOptions(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
		config.WithVersion("2.0.0"),
		config.WithDescription("Test application"),
	)
	
	if app == nil {
		t.Fatal("NewApplicationWithOptions returned nil")
	}
	
	if app.config.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", app.config.Name)
	}
	
	if app.config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", app.config.Version)
	}
	
	if app.config.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got '%s'", app.config.Description)
	}
}

func TestApplication_Register(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	cmd := &AppTestCommand{}
	err := app.Register(cmd)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	
	// Check that command was registered
	descriptor, exists := app.registry.GetCommand("test")
	if !exists {
		t.Error("Command not found after registration")
	}
	
	if descriptor.GetName() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", descriptor.GetName())
	}
}

func TestApplication_Register_Duplicate(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	cmd1 := &AppTestCommand{}
	cmd2 := &AppTestCommand{}
	
	err := app.Register(cmd1)
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}
	
	err = app.Register(cmd2)
	if err == nil {
		t.Error("Expected error for duplicate command registration")
	}
}

func TestApplication_Run(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	cmd := &AppTestCommand{}
	err := app.Register(cmd)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	
	// Test successful command execution
	exitCode := app.Run(context.Background(), []string{"test", "--value", "testvalue"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	if !cmd.executed {
		t.Error("Command was not executed")
	}
	
	if cmd.config.Value != "testvalue" {
		t.Errorf("Expected value 'testvalue', got '%s'", cmd.config.Value)
	}
}

func TestApplication_Run_CommandNotFound(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	// Test with non-existent command
	exitCode := app.Run(context.Background(), []string{"nonexistent"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent command")
	}
}

func TestApplication_Run_EmptyArgs(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	// Test with empty args (should show help)
	exitCode := app.Run(context.Background(), []string{})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}
}

func TestApplication_GetRegistry(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	cmd := &AppTestCommand{}
	_ = app.Register(cmd) // Error checked in other tests
	
	registry := app.GetRegistry()
	if registry == nil {
		t.Error("GetRegistry returned nil")
	}
	
	// Check that command was registered through registry
	_, exists := registry.GetCommand("test")
	if !exists {
		t.Error("Command not found in registry")
	}
}

func TestApplication_GetExecutor(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	executor := app.GetExecutor()
	if executor == nil {
		t.Error("GetExecutor returned nil")
	}
}

func TestApplication_GetConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Name = "test-app"
	cfg.Version = "1.0.0"
	cfg.Apply()
	
	app := NewApplication(cfg)
	
	if app.GetConfig() != cfg {
		t.Error("GetConfig did not return the correct config")
	}
}

func TestApplication_RegisterCommands(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	// Create commands with different names
	cmd1 := core.NewCommand("cmd1", "First command", func(ctx context.Context, config AppTestConfig) error {
		return nil
	})
	cmd2 := core.NewCommand("cmd2", "Second command", func(ctx context.Context, config AppTestConfig) error {
		return nil
	})
	
	err := app.RegisterCommands(cmd1, cmd2)
	if err != nil {
		t.Fatalf("RegisterCommands failed: %v", err)
	}
	
	// Verify both commands were registered
	registry := app.GetRegistry()
	commands := registry.ListCommands()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}

func TestApplication_GetHelpGenerator(t *testing.T) {
	app := NewApplicationWithOptions(
		config.WithName("test-app"),
	)
	
	helpGen := app.GetHelpGenerator()
	if helpGen == nil {
		t.Error("GetHelpGenerator returned nil")
	}
}