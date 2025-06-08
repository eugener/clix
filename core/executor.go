package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"time"

	"github.com/eugener/clix/internal/bind"
	configutils "github.com/eugener/clix/internal/config"
	"github.com/eugener/clix/internal/posix"
)

// ExecutionContext provides enhanced context for command execution
type ExecutionContext struct {
	context.Context
	Logger      *slog.Logger
	StartTime   time.Time
	CommandName string
	Args        []string
	Metadata    map[string]any
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, commandName string, args []string) *ExecutionContext {
	return &ExecutionContext{
		Context:     ctx,
		Logger:      slog.Default(),
		StartTime:   time.Now(),
		CommandName: commandName,
		Args:        args,
		Metadata:    make(map[string]any),
	}
}

// WithLogger sets a custom logger
func (ec *ExecutionContext) WithLogger(logger *slog.Logger) *ExecutionContext {
	ec.Logger = logger
	return ec
}

// WithMetadata adds metadata to the context
func (ec *ExecutionContext) WithMetadata(key string, value any) *ExecutionContext {
	ec.Metadata[key] = value
	return ec
}

// Duration returns the elapsed time since context creation
func (ec *ExecutionContext) Duration() time.Duration {
	return time.Since(ec.StartTime)
}

// Middleware represents command execution middleware
type Middleware func(next ExecuteFunc) ExecuteFunc

// ExecuteFunc represents a command execution function
type ExecuteFunc func(ctx *ExecutionContext) error

// Executor manages command execution with middleware support
type Executor struct {
	registry   *Registry
	binder     *bind.Binder
	middleware []Middleware
	logger     *slog.Logger
}

// NewExecutor creates a new command executor
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry:   registry,
		binder:     bind.NewBinder("posix"),
		middleware: make([]Middleware, 0),
		logger:     slog.Default(),
	}
}

// Use adds middleware to the execution chain
func (e *Executor) Use(middleware ...Middleware) {
	e.middleware = append(e.middleware, middleware...)
}

// SetLogger sets the executor logger
func (e *Executor) SetLogger(logger *slog.Logger) {
	e.logger = logger
}

// Execute runs a command with the given context and arguments
func (e *Executor) Execute(ctx context.Context, commandName string, args []string) error {
	return e.ExecuteWithConfig(ctx, commandName, args, nil)
}

// ExecuteWithConfig runs a command with the given context, arguments, and base configuration
func (e *Executor) ExecuteWithConfig(ctx context.Context, commandName string, args []string, baseConfig any) error {
	// Create execution context
	execCtx := NewExecutionContext(ctx, commandName, args).WithLogger(e.logger)

	// Get command descriptor
	descriptor, exists := e.registry.GetCommand(commandName)
	if !exists {
		return fmt.Errorf("command not found: %s", commandName)
	}

	// Create the base execution function
	baseFunc := func(execCtx *ExecutionContext) error {
		return e.executeCommandWithConfig(execCtx, descriptor, args, baseConfig)
	}

	// Build middleware chain
	executeFunc := e.buildMiddlewareChain(baseFunc)

	// Execute with middleware
	return executeFunc(execCtx)
}

// executeCommand method removed - use executeCommandWithConfig directly

// executeCommandWithConfig executes the actual command with base configuration
func (e *Executor) executeCommandWithConfig(execCtx *ExecutionContext, descriptor Command, args []string, baseConfig any) error {
	// Create config instance
	configType := descriptor.GetConfigType()
	configPtr := reflect.New(configType)
	config := configPtr.Interface()

	// Apply base configuration if provided (from config file)
	if baseConfig != nil {
		if err := configutils.MergeConfigs(config, baseConfig); err != nil {
			return fmt.Errorf("failed to apply base configuration: %w", err)
		}
	}

	// Parse arguments using enhanced parser (CLI args override config file)
	parser := NewEnhancedParser(e.binder)
	if err := parser.Parse(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate configuration
	if err := configutils.ValidateConfig(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Log execution start
	execCtx.Logger.Info("executing command",
		"command", execCtx.CommandName,
		"args", execCtx.Args,
		"duration_so_far", execCtx.Duration(),
	)

	// Execute the command
	return e.registry.Execute(execCtx.Context, execCtx.CommandName, config)
}

// buildMiddlewareChain builds the middleware execution chain
func (e *Executor) buildMiddlewareChain(base ExecuteFunc) ExecuteFunc {
	// Start with the base function
	executeFunc := base

	// Apply middleware in reverse order (last added runs first)
	for i := len(e.middleware) - 1; i >= 0; i-- {
		executeFunc = e.middleware[i](executeFunc)
	}

	return executeFunc
}


// EnhancedParser wraps the POSIX parser with additional functionality
type EnhancedParser struct {
	binder *bind.Binder
}

// NewEnhancedParser creates a new enhanced parser
func NewEnhancedParser(binder *bind.Binder) *EnhancedParser {
	return &EnhancedParser{binder: binder}
}

// Parse parses arguments and applies environment variables and defaults
func (ep *EnhancedParser) Parse(args []string, target any) error {
	// Apply environment variables first
	if err := ep.applyEnvironmentVariables(target); err != nil {
		return err
	}

	// Parse command line arguments using POSIX parser
	parser := posix.NewConfigurableParser(nil)
	result, err := parser.Parse(args)
	if err != nil {
		return err
	}

	// Bind values to struct
	return ep.binder.BindValues(target, result.Flags, result.Positional)
}

// applyEnvironmentVariables applies environment variable values
func (ep *EnhancedParser) applyEnvironmentVariables(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	analyzer := bind.NewAnalyzer("posix")
	metadata, err := analyzer.Analyze(targetStruct.Type())
	if err != nil {
		return err
	}

	// Apply environment variables
	envValues := make(map[string]any)
	for envVar, fieldInfo := range metadata.Environment {
		if value := os.Getenv(envVar); value != "" {
			envValues[fieldInfo.Long] = value
		}
	}

	if len(envValues) > 0 {
		return ep.binder.BindValues(target, envValues, nil)
	}

	return nil
}


// Built-in middleware

// LoggingMiddleware logs command execution
func LoggingMiddleware(next ExecuteFunc) ExecuteFunc {
	return func(ctx *ExecutionContext) error {
		start := time.Now()

		ctx.Logger.Info("command started",
			"command", ctx.CommandName,
			"args", ctx.Args,
		)

		err := next(ctx)

		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelError
		}

		ctx.Logger.Log(ctx.Context, level, "command completed",
			"command", ctx.CommandName,
			"duration", time.Since(start),
			"error", err,
		)

		return err
	}
}

// TimeoutMiddleware adds timeout support
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx *ExecutionContext) error {
			timeoutCtx, cancel := context.WithTimeout(ctx.Context, timeout)
			defer cancel()

			newCtx := &ExecutionContext{
				Context:     timeoutCtx,
				Logger:      ctx.Logger,
				StartTime:   ctx.StartTime,
				CommandName: ctx.CommandName,
				Args:        ctx.Args,
				Metadata:    ctx.Metadata,
			}

			done := make(chan error, 1)
			go func() {
				done <- next(newCtx)
			}()

			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				return fmt.Errorf("command timed out after %v", timeout)
			}
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next ExecuteFunc) ExecuteFunc {
	return func(ctx *ExecutionContext) (err error) {
		defer func() {
			if r := recover(); r != nil {
				ctx.Logger.Error("command panicked",
					"command", ctx.CommandName,
					"panic", r,
				)
				err = fmt.Errorf("command panicked: %v", r)
			}
		}()

		return next(ctx)
	}
}
