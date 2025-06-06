package core

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

// Test configurations for error path tests
type ErrorTestConfig struct {
	Value string `posix:"v,value,Test value"`
	Count int    `posix:"c,count,Count value,default=1"`
}

// Test error paths and edge cases in executor functionality

func TestExecutor_ConfigMergingErrors(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config ErrorTestConfig) error {
		return nil
	})
	registry.Register(cmd)

	executor := NewExecutor(registry)

	// Test merging with incompatible types
	incompatibleConfig := struct {
		DifferentField string
	}{
		DifferentField: "value",
	}

	err := executor.ExecuteWithConfig(context.Background(), "test", []string{}, incompatibleConfig)
	if err == nil {
		t.Error("Expected error when merging incompatible config types")
	}

	// Test merging with nil base config
	err = executor.ExecuteWithConfig(context.Background(), "test", []string{}, nil)
	if err != nil {
		t.Errorf("Should handle nil base config gracefully: %v", err)
	}
}

func TestExecutor_EnvironmentVariableBinding(t *testing.T) {
	registry := NewRegistry()
	
	// Command with environment variable support
	envCmd := NewCommand("env", "Environment test", func(ctx context.Context, config TestConfigWithEnv) error {
		return nil
	})
	registry.Register(envCmd)

	executor := NewExecutor(registry)

	// Test environment variable parsing
	err := executor.Execute(context.Background(), "env", []string{})
	if err != nil {
		t.Errorf("Environment variable parsing failed: %v", err)
	}
}

func TestExecutor_ChoicesValidation(t *testing.T) {
	registry := NewRegistry()
	
	choicesCmd := NewCommand("choices", "Choices validation test", func(ctx context.Context, config TestConfigWithChoices) error {
		return nil
	})
	registry.Register(choicesCmd)

	executor := NewExecutor(registry)

	// Test valid choice
	err := executor.Execute(context.Background(), "choices", []string{"--choice", "opt1"})
	if err != nil {
		t.Errorf("Valid choice should not produce error: %v", err)
	}

	// Test invalid choice
	err = executor.Execute(context.Background(), "choices", []string{"--choice", "invalid"})
	if err == nil {
		t.Error("Expected validation error for invalid choice")
	}

	if !isValidationError(err.Error()) {
		t.Errorf("Expected choices validation error, got: %v", err)
	}
}

func TestExecutor_ReflectionErrors(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	// Test with command that has invalid config type (should be caught during registration)
	defer func() {
		if r := recover(); r != nil {
			// This is expected if the registry catches invalid types
		}
	}()

	// Create command descriptor manually to test error paths
	descriptor := &commandDescriptor{
		name:       "invalid",
		desc:       "Invalid command",
		configType: reflect.TypeOf(42), // Invalid type (not a struct)
		instance: &struct{
			name string
			desc string
		}{
			name: "invalid",
			desc: "Invalid command",
		},
	}

	// Directly test executeCommandWithConfig with invalid descriptor
	execCtx := NewExecutionContext(context.Background(), "invalid", []string{})
	err := executor.executeCommandWithConfig(execCtx, descriptor, []string{}, nil)
	if err == nil {
		t.Error("Expected error with invalid config type")
	}
}

func TestExecutor_BindingErrors(t *testing.T) {
	registry := NewRegistry()
	
	// Command with complex binding requirements
	complexCmd := NewCommand("complex", "Complex binding test", func(ctx context.Context, config ComplexTestConfig) error {
		return nil
	})
	registry.Register(complexCmd)

	executor := NewExecutor(registry)

	// Test type conversion errors
	err := executor.Execute(context.Background(), "complex", []string{"--int-field", "not-a-number"})
	if err == nil {
		t.Error("Expected error for invalid integer conversion")
	}

	// Test slice binding
	err = executor.Execute(context.Background(), "complex", []string{"--int-field", "42", "arg1", "arg2"})
	if err != nil {
		t.Errorf("Slice binding should work: %v", err)
	}
}

func TestExecutor_DefaultValueApplication(t *testing.T) {
	registry := NewRegistry()
	
	defaultsCmd := NewCommand("defaults", "Defaults test", func(ctx context.Context, config TestConfigWithDefaults) error {
		// Verify defaults were applied
		if config.DefaultField != "default-value" {
			return fmt.Errorf("expected default value, got: %s", config.DefaultField)
		}
		return nil
	})
	registry.Register(defaultsCmd)

	executor := NewExecutor(registry)

	// Test that defaults are applied when no value provided
	err := executor.Execute(context.Background(), "defaults", []string{})
	if err != nil {
		t.Errorf("Default value application failed: %v", err)
	}

	// Test that explicit values override defaults
	explicitCmd := NewCommand("explicit", "Explicit test", func(ctx context.Context, config TestConfigWithDefaults) error {
		if config.DefaultField != "explicit-value" {
			return fmt.Errorf("expected explicit value, got: %s", config.DefaultField)
		}
		return nil
	})
	registry.Register(explicitCmd)

	err = executor.Execute(context.Background(), "explicit", []string{"--default-field", "explicit-value"})
	if err != nil {
		t.Errorf("Explicit value override failed: %v", err)
	}
}

func TestExecutor_MiddlewareErrorHandling(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config ErrorTestConfig) error {
		return nil
	})
	registry.Register(cmd)

	executor := NewExecutor(registry)

	// Middleware that returns error
	errorMiddleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			return fmt.Errorf("middleware error")
		}
	}

	executor.Use(errorMiddleware)

	err := executor.Execute(context.Background(), "test", []string{})
	if err == nil {
		t.Error("Expected error from middleware")
	}

	if !containsString(err.Error(), "middleware error") {
		t.Errorf("Expected middleware error, got: %v", err)
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	registry := NewRegistry()
	
	// Command that checks for context cancellation
	cancelCmd := NewCommand("cancel", "Cancellation test", func(ctx context.Context, config ErrorTestConfig) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})
	registry.Register(cancelCmd)

	executor := NewExecutor(registry)

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := executor.Execute(cancelledCtx, "cancel", []string{})
	if err == nil {
		t.Error("Expected error from cancelled context")
	}

	if !containsString(err.Error(), "canceled") {
		t.Errorf("Expected cancellation error, got: %v", err)
	}
}

func TestExecutor_EdgeCaseArguments(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("edge", "Edge case test", func(ctx context.Context, config ErrorTestConfig) error {
		return nil
	})
	registry.Register(cmd)

	executor := NewExecutor(registry)

	// Test empty args
	err := executor.Execute(context.Background(), "edge", []string{})
	if err != nil {
		t.Errorf("Empty args should not cause error: %v", err)
	}

	// Test args with only flags
	err = executor.Execute(context.Background(), "edge", []string{"--value", "test"})
	if err != nil {
		t.Errorf("Flag-only args should not cause error: %v", err)
	}

	// Test args with special characters
	err = executor.Execute(context.Background(), "edge", []string{"--value", "test with spaces and symbols !@#$%"})
	if err != nil {
		t.Errorf("Special character args should not cause error: %v", err)
	}
}

func TestExecutor_BuildMiddlewareChainEdgeCases(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config ErrorTestConfig) error {
		return nil
	})
	registry.Register(cmd)

	executor := NewExecutor(registry)

	// Test with no middleware
	err := executor.Execute(context.Background(), "test", []string{})
	if err != nil {
		t.Errorf("Should work with no middleware: %v", err)
	}

	// Test with single middleware
	counter := 0
	singleMiddleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			counter++
			return next(ctx)
		}
	}

	executor.Use(singleMiddleware)
	err = executor.Execute(context.Background(), "test", []string{})
	if err != nil {
		t.Errorf("Should work with single middleware: %v", err)
	}

	if counter != 1 {
		t.Errorf("Expected middleware to be called once, got %d", counter)
	}
}

func TestExecutor_ValidationWithComplexTypes(t *testing.T) {
	registry := NewRegistry()
	
	complexValidationCmd := NewCommand("complex-validation", "Complex validation test", func(ctx context.Context, config ComplexValidationConfig) error {
		return nil
	})
	registry.Register(complexValidationCmd)

	executor := NewExecutor(registry)

	// Test with nested validation requirements
	err := executor.Execute(context.Background(), "complex-validation", []string{})
	if err == nil {
		t.Error("Expected validation error for missing required fields")
	}

	// Test with valid complex config
	err = executor.Execute(context.Background(), "complex-validation", []string{
		"--required-field", "value",
		"--int-field", "42",
		"positional-arg",
	})
	if err != nil {
		t.Errorf("Valid complex config should not produce error: %v", err)
	}
}

// Test configuration structs

type TestConfigWithEnv struct {
	Value string `posix:"v,value,Test value,env=TEST_VALUE"`
}

type TestConfigWithDefaults struct {
	DefaultField string `posix:"d,default-field,Field with default,default=default-value"`
}

type ComplexTestConfig struct {
	IntField int      `posix:"i,int-field,Integer field"`
	Slice    []string `posix:",,Slice field,positional"`
}

type ComplexValidationConfig struct {
	RequiredField string   `posix:"r,required-field,Required field,required"`
	IntField      int      `posix:"i,int-field,Integer field"`
	Choice        string   `posix:"c,choice,Choice field,choices=opt1;opt2;opt3"`
	Positional    []string `posix:",,Positional args,positional"`
}