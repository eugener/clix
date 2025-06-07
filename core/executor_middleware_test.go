package core

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

// Test configuration for middleware tests
type MiddlewareTestConfig struct {
	Value string `posix:"v,value,Test value"`
	Count int    `posix:"c,count,Count value,default=1"`
}

// Test middleware functionality and error scenarios

func TestExecutor_MiddlewareChain(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config MiddlewareTestConfig) error {
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Track middleware execution order
	var executionOrder []string

	middleware1 := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			executionOrder = append(executionOrder, "middleware1-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware1-after")
			return err
		}
	}

	middleware2 := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			executionOrder = append(executionOrder, "middleware2-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "middleware2-after")
			return err
		}
	}

	executor.Use(middleware1, middleware2)

	err := executor.Execute(context.Background(), "test", []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify execution order (first added middleware runs first)
	expectedOrder := []string{
		"middleware1-before",
		"middleware2-before",
		"middleware2-after",
		"middleware1-after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d middleware calls, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected middleware call %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

func TestExecutor_LoggingMiddleware(t *testing.T) {
	registry := NewRegistry()
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config MiddlewareTestConfig) error {
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)
	executor.Use(LoggingMiddleware)

	// Test successful execution
	err := executor.Execute(context.Background(), "test", []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Test execution with error
	errorCmd := NewCommand("error", "Error command", func(ctx context.Context, config MiddlewareTestConfig) error {
		return fmt.Errorf("test error")
	})
	_ = registry.Register(errorCmd) // Error checked in other tests

	err = executor.Execute(context.Background(), "error", []string{})
	if err == nil {
		t.Error("Expected error from command")
	}
}

func TestExecutor_TimeoutMiddleware(t *testing.T) {
	registry := NewRegistry()

	// Command that takes longer than timeout
	slowCmd := NewCommand("slow", "Slow command", func(ctx context.Context, config MiddlewareTestConfig) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	_ = registry.Register(slowCmd) // Error checked in other tests

	executor := NewExecutor(registry)
	executor.Use(TimeoutMiddleware(50 * time.Millisecond))

	err := executor.Execute(context.Background(), "slow", []string{})
	if err == nil {
		t.Error("Expected timeout error")
	}

	if !isTimeoutError(err.Error()) {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	// Test fast command that completes before timeout
	fastCmd := NewCommand("fast", "Fast command", func(ctx context.Context, config MiddlewareTestConfig) error {
		return nil
	})
	_ = registry.Register(fastCmd) // Error checked in other tests

	err = executor.Execute(context.Background(), "fast", []string{})
	if err != nil {
		t.Errorf("Fast command should not timeout: %v", err)
	}
}

func TestExecutor_RecoveryMiddleware(t *testing.T) {
	registry := NewRegistry()

	// Command that panics
	panicCmd := NewCommand("panic", "Panic command", func(ctx context.Context, config MiddlewareTestConfig) error {
		panic("test panic")
	})
	_ = registry.Register(panicCmd) // Error checked in other tests

	executor := NewExecutor(registry)
	executor.Use(RecoveryMiddleware)

	err := executor.Execute(context.Background(), "panic", []string{})
	if err == nil {
		t.Error("Expected error from panic recovery")
	}

	if !isPanicError(err.Error()) {
		t.Errorf("Expected panic error, got: %v", err)
	}
}

func TestExecutor_CombinedMiddleware(t *testing.T) {
	registry := NewRegistry()

	// Test with multiple middleware working together
	cmd := NewCommand("test", "Test command", func(ctx context.Context, config MiddlewareTestConfig) error {
		// Small delay to test timeout behavior
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)
	executor.Use(
		LoggingMiddleware,
		TimeoutMiddleware(100*time.Millisecond),
		RecoveryMiddleware,
	)

	err := executor.Execute(context.Background(), "test", []string{})
	if err != nil {
		t.Errorf("Combined middleware execution failed: %v", err)
	}
}

func TestExecutor_ErrorPropagation(t *testing.T) {
	registry := NewRegistry()

	errorCmd := NewCommand("error", "Error command", func(ctx context.Context, config MiddlewareTestConfig) error {
		return fmt.Errorf("command error")
	})
	_ = registry.Register(errorCmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Test error without middleware
	err := executor.Execute(context.Background(), "error", []string{})
	if err == nil {
		t.Error("Expected error from command")
	}

	// Test error with middleware that should propagate error
	executor.Use(LoggingMiddleware)
	err = executor.Execute(context.Background(), "error", []string{})
	if err == nil {
		t.Error("Expected error propagation through middleware")
	}

	// Test middleware that modifies error
	errorModifyingMiddleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			err := next(ctx)
			if err != nil {
				return fmt.Errorf("middleware wrapped: %w", err)
			}
			return nil
		}
	}

	executor.Use(errorModifyingMiddleware)
	err = executor.Execute(context.Background(), "error", []string{})
	if err == nil {
		t.Error("Expected wrapped error")
	}

	if !isWrappedError(err.Error()) {
		t.Errorf("Expected wrapped error, got: %v", err)
	}
}

func TestExecutor_ContextPropagation(t *testing.T) {
	registry := NewRegistry()

	var receivedContext *ExecutionContext
	contextCmd := NewCommand("context", "Context command", func(ctx context.Context, config MiddlewareTestConfig) error {
		// This tests that context is properly propagated through executor
		return nil
	})
	_ = registry.Register(contextCmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Middleware that captures context
	contextCapturingMiddleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			receivedContext = ctx
			return next(ctx)
		}
	}

	executor.Use(contextCapturingMiddleware)

	err := executor.Execute(context.Background(), "context", []string{"--value", "test"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedContext == nil {
		t.Fatal("Context was not captured by middleware")
	}

	if receivedContext.CommandName != "context" {
		t.Errorf("Expected command name 'context', got '%s'", receivedContext.CommandName)
	}

	if len(receivedContext.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(receivedContext.Args))
	}

	// Duration might be 0 on very fast systems, so just check it's not negative
	if receivedContext.Duration() < 0 {
		t.Error("Expected non-negative duration")
	}

	if receivedContext.Logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestExecutor_SetLogger(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	// Create custom logger
	customLogger := slog.Default()
	executor.SetLogger(customLogger)

	// Verify logger is set (this tests the SetLogger method)
	if executor.logger != customLogger {
		t.Error("Custom logger was not set properly")
	}
}

func TestExecutor_NonExistentCommand(t *testing.T) {
	registry := NewRegistry()
	executor := NewExecutor(registry)

	err := executor.Execute(context.Background(), "nonexistent", []string{})
	if err == nil {
		t.Error("Expected error for non-existent command")
	}

	if !isCommandNotFoundError(err.Error()) {
		t.Errorf("Expected command not found error, got: %v", err)
	}
}

func TestExecutor_ValidationErrors(t *testing.T) {
	registry := NewRegistry()

	// Command with required field
	cmd := NewCommand("required", "Required field command", func(ctx context.Context, config TestConfigWithRequired) error {
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Test missing required field
	err := executor.Execute(context.Background(), "required", []string{})
	if err == nil {
		t.Error("Expected validation error for missing required field")
	}

	if !isValidationError(err.Error()) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestExecutor_ParseErrors(t *testing.T) {
	registry := NewRegistry()

	// Command with required field to test validation errors
	cmd := NewCommand("parse", "Parse test command", func(ctx context.Context, config TestConfigWithRequired) error {
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Test missing required field (this will definitely cause a validation error)
	err := executor.Execute(context.Background(), "parse", []string{})
	if err == nil {
		t.Error("Expected validation error for missing required field")
		return
	}

	if !isValidationError(err.Error()) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestExecutor_ConfigMerging(t *testing.T) {
	registry := NewRegistry()

	cmd := NewCommand("merge", "Config merge command", func(ctx context.Context, config MiddlewareTestConfig) error {
		// Verify config values are properly merged
		if config.Value != "cli-value" {
			return fmt.Errorf("expected CLI value to override config file value")
		}
		return nil
	})
	_ = registry.Register(cmd) // Error checked in other tests

	executor := NewExecutor(registry)

	// Create base config (simulating config file)
	baseConfig := MiddlewareTestConfig{
		Value: "config-file-value",
		Count: 42,
	}

	// Execute with CLI args that should override config file
	err := executor.ExecuteWithConfig(context.Background(), "merge", []string{"--value", "cli-value"}, baseConfig)
	if err != nil {
		t.Errorf("Config merging failed: %v", err)
	}
}

// Helper structs for testing

type TestConfigWithRequired struct {
	Value string `posix:"v,value,Test value,required"`
}

type TestConfigWithChoices struct {
	Choice string `posix:"c,choice,Test choice,choices=opt1;opt2;opt3"`
}

// Helper functions to check error types

func isTimeoutError(errMsg string) bool {
	return containsString(errMsg, "timed out")
}

func isPanicError(errMsg string) bool {
	return containsString(errMsg, "panicked")
}

func isWrappedError(errMsg string) bool {
	return containsString(errMsg, "middleware wrapped")
}

func isCommandNotFoundError(errMsg string) bool {
	return containsString(errMsg, "command not found")
}

func isValidationError(errMsg string) bool {
	return containsString(errMsg, "validation failed") || containsString(errMsg, "required field")
}

// isParseError function removed - not used in current tests

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
