package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"time"

	"claude-code-test/bind"
	"claude-code-test/posix"
)

// ExecutionContext provides enhanced context for command execution
type ExecutionContext struct {
	context.Context
	Logger     *slog.Logger
	StartTime  time.Time
	CommandName string
	Args       []string
	Metadata   map[string]any
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

// executeCommand executes the actual command
func (e *Executor) executeCommand(execCtx *ExecutionContext, descriptor *commandDescriptor, args []string) error {
	return e.executeCommandWithConfig(execCtx, descriptor, args, nil)
}

// executeCommandWithConfig executes the actual command with base configuration
func (e *Executor) executeCommandWithConfig(execCtx *ExecutionContext, descriptor *commandDescriptor, args []string, baseConfig any) error {
	// Create config instance
	configType := descriptor.GetConfigType()
	configPtr := reflect.New(configType)
	config := configPtr.Interface()
	
	// Apply base configuration if provided (from config file)
	if baseConfig != nil {
		if err := e.mergeConfigs(config, baseConfig); err != nil {
			return fmt.Errorf("failed to apply base configuration: %w", err)
		}
	}
	
	// Parse arguments using enhanced parser (CLI args override config file)
	parser := NewEnhancedParser(e.binder)
	if err := parser.Parse(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	
	// Validate configuration
	if err := e.validateConfig(config); err != nil {
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

// validateConfig validates the parsed configuration
func (e *Executor) validateConfig(config any) error {
	// Use the binder's analyzer for validation
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}
	
	analyzer := bind.NewAnalyzer("posix")
	metadata, err := analyzer.Analyze(configValue.Type())
	if err != nil {
		return err
	}
	
	// Check required fields
	for _, fieldInfo := range metadata.Fields {
		if !fieldInfo.Required {
			continue
		}
		
		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			return fmt.Errorf("required field %s is missing", fieldInfo.Name)
		}
	}
	
	// Check choices validation
	for _, fieldInfo := range metadata.Fields {
		if len(fieldInfo.Choices) == 0 {
			continue
		}
		
		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			continue
		}
		
		value := fmt.Sprintf("%v", field.Interface())
		valid := false
		for _, choice := range fieldInfo.Choices {
			if value == choice {
				valid = true
				break
			}
		}
		
		if !valid {
			return fmt.Errorf("field %s must be one of: %v", fieldInfo.Name, fieldInfo.Choices)
		}
	}
	
	return nil
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

// mergeConfigs merges base configuration into target configuration
// Values in target (from CLI args) take precedence over base (from config file)
func (e *Executor) mergeConfigs(target, base any) error {
	targetValue := reflect.ValueOf(target)
	baseValue := reflect.ValueOf(base)
	
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}
	
	if baseValue.Kind() == reflect.Ptr {
		baseValue = baseValue.Elem()
	}
	
	if baseValue.Kind() != reflect.Struct {
		return fmt.Errorf("base must be a struct or pointer to struct")
	}
	
	targetStruct := targetValue.Elem()
	baseStruct := baseValue
	
	// Check that both structs have the same type
	if targetStruct.Type() != baseStruct.Type() {
		return fmt.Errorf("target and base configurations must have the same type")
	}
	
	// Copy non-zero values from base to target where target field is zero
	for i := 0; i < targetStruct.NumField(); i++ {
		targetField := targetStruct.Field(i)
		baseField := baseStruct.Field(i)
		
		// Skip unexported fields
		if !targetField.CanSet() {
			continue
		}
		
		// If target field is zero and base field is not zero, copy from base
		if targetField.IsZero() && !baseField.IsZero() {
			if targetField.Type() == baseField.Type() {
				targetField.Set(baseField)
			}
		}
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