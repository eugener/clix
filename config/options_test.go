package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/eugener/clix/core"
	"github.com/eugener/clix/internal/help"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	
	if config.Name != "cli" {
		t.Errorf("Expected default name 'cli', got '%s'", config.Name)
	}
	
	if config.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", config.Version)
	}
	
	if config.ErrorHandler == nil {
		t.Error("ErrorHandler should be initialized")
	}
}

func TestWithName(t *testing.T) {
	config := &CLIConfig{}
	option := WithName("test-app")
	option(config)
	
	if config.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", config.Name)
	}
}

func TestWithVersion(t *testing.T) {
	config := &CLIConfig{}
	option := WithVersion("2.0.0")
	option(config)
	
	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}
}

func TestWithDescription(t *testing.T) {
	config := &CLIConfig{}
	option := WithDescription("Test application")
	option(config)
	
	if config.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got '%s'", config.Description)
	}
}

func TestWithAuthor(t *testing.T) {
	config := &CLIConfig{}
	option := WithAuthor("Test Author")
	option(config)
	
	if config.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", config.Author)
	}
}

func TestWithLogger(t *testing.T) {
	config := &CLIConfig{}
	logger := slog.Default()
	option := WithLogger(logger)
	option(config)
	
	if config.Logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestWithDefaultTimeout(t *testing.T) {
	config := &CLIConfig{}
	timeout := 30 * time.Second
	option := WithDefaultTimeout(timeout)
	option(config)
	
	if config.DefaultTimeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, config.DefaultTimeout)
	}
}

func TestWithMiddleware(t *testing.T) {
	config := &CLIConfig{}
	middleware1 := core.LoggingMiddleware
	middleware2 := core.RecoveryMiddleware
	
	option := WithMiddleware(middleware1, middleware2)
	option(config)
	
	if len(config.Middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(config.Middleware))
	}
}

func TestWithRecovery(t *testing.T) {
	config := &CLIConfig{}
	option := WithRecovery()
	option(config)
	
	if len(config.Middleware) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(config.Middleware))
	}
}

func TestWithLogging(t *testing.T) {
	config := &CLIConfig{}
	option := WithLogging()
	option(config)
	
	if len(config.Middleware) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(config.Middleware))
	}
}

func TestWithTimeout(t *testing.T) {
	config := &CLIConfig{}
	timeout := 15 * time.Second
	option := WithTimeout(timeout)
	option(config)
	
	if len(config.Middleware) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(config.Middleware))
	}
}

func TestWithHelpConfig(t *testing.T) {
	config := &CLIConfig{}
	helpConfig := &help.HelpConfig{
		ColorEnabled: true,
		MaxWidth:     120,
	}
	option := WithHelpConfig(helpConfig)
	option(config)
	
	if config.HelpConfig != helpConfig {
		t.Error("HelpConfig not set correctly")
	}
}

func TestWithColoredOutput(t *testing.T) {
	config := &CLIConfig{
		HelpConfig: &help.HelpConfig{},
	}
	option := WithColoredOutput(true)
	option(config)
	
	if !config.HelpConfig.ColorEnabled {
		t.Error("ColorEnabled should be true")
	}
	
	option = WithColoredOutput(false)
	option(config)
	
	if config.HelpConfig.ColorEnabled {
		t.Error("ColorEnabled should be false")
	}
}

func TestWithMaxHelpWidth(t *testing.T) {
	config := &CLIConfig{
		HelpConfig: &help.HelpConfig{},
	}
	option := WithMaxHelpWidth(100)
	option(config)
	
	if config.HelpConfig.MaxWidth != 100 {
		t.Errorf("Expected MaxWidth 100, got %d", config.HelpConfig.MaxWidth)
	}
}

func TestWithErrorHandler(t *testing.T) {
	config := &CLIConfig{}
	handler := func(error) int { return 42 }
	option := WithErrorHandler(handler)
	option(config)
	
	if config.ErrorHandler == nil {
		t.Error("ErrorHandler not set")
	}
	
	// Test the handler works
	result := config.ErrorHandler(nil)
	if result != 42 {
		t.Errorf("Expected error handler to return 42, got %d", result)
	}
}

func TestWithBeforeAll(t *testing.T) {
	config := &CLIConfig{}
	hook := func(*core.ExecutionContext) error { return nil }
	option := WithBeforeAll(hook)
	option(config)
	
	if config.BeforeAll == nil {
		t.Error("BeforeAll hook not set")
	}
}

func TestWithAfterAll(t *testing.T) {
	config := &CLIConfig{}
	hook := func(*core.ExecutionContext) error { return nil }
	option := WithAfterAll(hook)
	option(config)
	
	if config.AfterAll == nil {
		t.Error("AfterAll hook not set")
	}
}

func TestWithBeforeEach(t *testing.T) {
	config := &CLIConfig{}
	hook := func(*core.ExecutionContext) error { return nil }
	option := WithBeforeEach(hook)
	option(config)
	
	if config.BeforeEach == nil {
		t.Error("BeforeEach hook not set")
	}
}

func TestWithAfterEach(t *testing.T) {
	config := &CLIConfig{}
	hook := func(*core.ExecutionContext) error { return nil }
	option := WithAfterEach(hook)
	option(config)
	
	if config.AfterEach == nil {
		t.Error("AfterEach hook not set")
	}
}

func TestWithGlobalFlag(t *testing.T) {
	config := &CLIConfig{
		GlobalFlags: make(map[string]interface{}),
	}
	option := WithGlobalFlag("verbose", true)
	option(config)
	
	if len(config.GlobalFlags) != 1 {
		t.Errorf("Expected 1 global flag, got %d", len(config.GlobalFlags))
	}
	
	if config.GlobalFlags["verbose"] != true {
		t.Error("Global flag not set correctly")
	}
}

func TestWithConfigFile(t *testing.T) {
	config := &CLIConfig{}
	option := WithConfigFile("config.yaml")
	option(config)
	
	if config.ConfigFile != "config.yaml" {
		t.Errorf("Expected config file 'config.yaml', got '%s'", config.ConfigFile)
	}
}

func TestWithConfigPaths(t *testing.T) {
	config := &CLIConfig{}
	paths := []string{"/etc/myapp", "~/.myapp"}
	option := WithConfigPaths(paths)
	option(config)
	
	if len(config.ConfigPaths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(config.ConfigPaths))
	}
	
	if config.ConfigPaths[0] != "/etc/myapp" {
		t.Errorf("Expected first path '/etc/myapp', got '%s'", config.ConfigPaths[0])
	}
}

func TestWithAutoLoadConfig(t *testing.T) {
	config := &CLIConfig{}
	option := WithAutoLoadConfig(true)
	option(config)
	
	if !config.AutoLoadConfig {
		t.Error("AutoLoadConfig should be true")
	}
	
	option = WithAutoLoadConfig(false)
	option(config)
	
	if config.AutoLoadConfig {
		t.Error("AutoLoadConfig should be false")
	}
}

func TestWithInteractiveMode(t *testing.T) {
	config := &CLIConfig{}
	option := WithInteractiveMode(true)
	option(config)
	
	if !config.InteractiveMode {
		t.Error("InteractiveMode should be true")
	}
	
	option = WithInteractiveMode(false)
	option(config)
	
	if config.InteractiveMode {
		t.Error("InteractiveMode should be false")
	}
}

func TestCLIConfig_Apply(t *testing.T) {
	config := DefaultConfig()
	config.Apply(
		WithName("test-app"),
		WithVersion("2.0.0"),
		WithDescription("Test application"),
		WithAuthor("Test Author"),
		WithRecovery(),
		WithLogging(),
	)
	
	if config.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", config.Name)
	}
	
	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}
	
	if config.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got '%s'", config.Description)
	}
	
	if config.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", config.Author)
	}
	
	if len(config.Middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(config.Middleware))
	}
	
	if config.HelpConfig == nil {
		t.Error("HelpConfig should be initialized after Apply")
	}
}