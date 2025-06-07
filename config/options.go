package config

import (
	"log/slog"
	"time"

	"github.com/eugener/clix/core"
	"github.com/eugener/clix/internal/help"
	"github.com/eugener/clix/internal/output"
)

// CLIConfig holds the configuration for a CLI application
type CLIConfig struct {
	Name        string
	Version     string
	Description string
	Author      string
	
	// Help configuration
	HelpConfig *help.HelpConfig
	
	// Execution configuration
	DefaultTimeout time.Duration
	Logger         *slog.Logger
	
	// Middleware
	Middleware []core.Middleware
	
	// Global flags
	GlobalFlags map[string]interface{}
	
	// Configuration file settings
	ConfigFile     string
	ConfigPaths    []string
	AutoLoadConfig bool
	
	// Interactive mode settings
	InteractiveMode bool
	
	// Output formatting settings
	OutputFormat     output.Format
	OutputFormatter  *output.Formatter
	
	// Error handling
	ErrorHandler func(error) int
	
	// Hooks
	BeforeAll  func(*core.ExecutionContext) error
	AfterAll   func(*core.ExecutionContext) error
	BeforeEach func(*core.ExecutionContext) error
	AfterEach  func(*core.ExecutionContext) error
}

// Option represents a functional option for configuring the CLI
type Option func(*CLIConfig)

// WithName sets the CLI application name
func WithName(name string) Option {
	return func(c *CLIConfig) {
		c.Name = name
		if c.HelpConfig != nil {
			c.HelpConfig.ProgramName = name
		}
	}
}

// WithVersion sets the CLI application version
func WithVersion(version string) Option {
	return func(c *CLIConfig) {
		c.Version = version
		if c.HelpConfig != nil {
			c.HelpConfig.Version = version
		}
	}
}

// WithDescription sets the CLI application description
func WithDescription(description string) Option {
	return func(c *CLIConfig) {
		c.Description = description
		if c.HelpConfig != nil {
			c.HelpConfig.Description = description
		}
	}
}

// WithAuthor sets the CLI application author
func WithAuthor(author string) Option {
	return func(c *CLIConfig) {
		c.Author = author
	}
}

// WithLogger sets the logger for the CLI
func WithLogger(logger *slog.Logger) Option {
	return func(c *CLIConfig) {
		c.Logger = logger
	}
}

// WithDefaultTimeout sets the default command timeout
func WithDefaultTimeout(timeout time.Duration) Option {
	return func(c *CLIConfig) {
		c.DefaultTimeout = timeout
	}
}

// WithMiddleware adds middleware to the CLI
func WithMiddleware(middleware ...core.Middleware) Option {
	return func(c *CLIConfig) {
		c.Middleware = append(c.Middleware, middleware...)
	}
}

// WithRecovery adds panic recovery middleware
func WithRecovery() Option {
	return func(c *CLIConfig) {
		c.Middleware = append(c.Middleware, core.RecoveryMiddleware)
	}
}

// WithLogging adds logging middleware
func WithLogging() Option {
	return func(c *CLIConfig) {
		c.Middleware = append(c.Middleware, core.LoggingMiddleware)
	}
}

// WithTimeout adds timeout middleware with the specified duration
func WithTimeout(timeout time.Duration) Option {
	return func(c *CLIConfig) {
		c.Middleware = append(c.Middleware, core.TimeoutMiddleware(timeout))
	}
}

// WithHelpConfig sets custom help configuration
func WithHelpConfig(helpConfig *help.HelpConfig) Option {
	return func(c *CLIConfig) {
		c.HelpConfig = helpConfig
	}
}

// WithColoredOutput enables or disables colored output
func WithColoredOutput(enabled bool) Option {
	return func(c *CLIConfig) {
		if c.HelpConfig == nil {
			c.HelpConfig = help.DefaultHelpConfig(c.Name)
		}
		c.HelpConfig.ColorEnabled = enabled
	}
}

// WithMaxHelpWidth sets the maximum width for help text
func WithMaxHelpWidth(width int) Option {
	return func(c *CLIConfig) {
		if c.HelpConfig == nil {
			c.HelpConfig = help.DefaultHelpConfig(c.Name)
		}
		c.HelpConfig.MaxWidth = width
	}
}

// WithErrorHandler sets a custom error handler
func WithErrorHandler(handler func(error) int) Option {
	return func(c *CLIConfig) {
		c.ErrorHandler = handler
	}
}

// WithBeforeAll sets a hook to run before all commands
func WithBeforeAll(hook func(*core.ExecutionContext) error) Option {
	return func(c *CLIConfig) {
		c.BeforeAll = hook
	}
}

// WithAfterAll sets a hook to run after all commands
func WithAfterAll(hook func(*core.ExecutionContext) error) Option {
	return func(c *CLIConfig) {
		c.AfterAll = hook
	}
}

// WithBeforeEach sets a hook to run before each command
func WithBeforeEach(hook func(*core.ExecutionContext) error) Option {
	return func(c *CLIConfig) {
		c.BeforeEach = hook
	}
}

// WithAfterEach sets a hook to run after each command
func WithAfterEach(hook func(*core.ExecutionContext) error) Option {
	return func(c *CLIConfig) {
		c.AfterEach = hook
	}
}

// WithGlobalFlag adds a global flag that applies to all commands
func WithGlobalFlag(name string, value interface{}) Option {
	return func(c *CLIConfig) {
		if c.GlobalFlags == nil {
			c.GlobalFlags = make(map[string]interface{})
		}
		c.GlobalFlags[name] = value
	}
}

// WithConfigFile sets the configuration file name
func WithConfigFile(filename string) Option {
	return func(c *CLIConfig) {
		c.ConfigFile = filename
		c.AutoLoadConfig = true
	}
}

// WithConfigPaths sets custom search paths for configuration files
func WithConfigPaths(paths []string) Option {
	return func(c *CLIConfig) {
		c.ConfigPaths = paths
	}
}

// WithAutoLoadConfig enables or disables automatic config loading
func WithAutoLoadConfig(enabled bool) Option {
	return func(c *CLIConfig) {
		c.AutoLoadConfig = enabled
	}
}

// WithInteractiveMode enables or disables interactive prompting for missing required fields
func WithInteractiveMode(enabled bool) Option {
	return func(c *CLIConfig) {
		c.InteractiveMode = enabled
	}
}

// WithOutputFormat sets the default output format for commands
func WithOutputFormat(format output.Format) Option {
	return func(c *CLIConfig) {
		c.OutputFormat = format
	}
}

// WithOutputFormatter sets a custom output formatter
func WithOutputFormatter(formatter *output.Formatter) Option {
	return func(c *CLIConfig) {
		c.OutputFormatter = formatter
	}
}

// DefaultConfig returns a default CLI configuration
func DefaultConfig() *CLIConfig {
	return &CLIConfig{
		Name:           "cli",
		Version:        "1.0.0",
		Description:    "",
		DefaultTimeout: 30 * time.Second,
		Logger:         slog.Default(),
		Middleware:     []core.Middleware{},
		GlobalFlags:    make(map[string]interface{}),
		ConfigFile:      "",
		ConfigPaths:     []string{},
		AutoLoadConfig:  false,
		InteractiveMode: false,
		OutputFormat:    output.FormatText,
		OutputFormatter: nil,
		ErrorHandler: func(err error) int {
			if err != nil {
				return 1
			}
			return 0
		},
	}
}

// Apply applies all options to the configuration
func (c *CLIConfig) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	
	// Ensure help config is set
	if c.HelpConfig == nil {
		c.HelpConfig = help.DefaultHelpConfig(c.Name)
		c.HelpConfig.Version = c.Version
		c.HelpConfig.Description = c.Description
	}
}

// Builder provides a fluent interface for building CLI configurations
type Builder struct {
	config *CLIConfig
}

// NewBuilder creates a new configuration builder
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultConfig(),
	}
}

// Name sets the CLI application name
func (b *Builder) Name(name string) *Builder {
	b.config.Apply(WithName(name))
	return b
}

// Version sets the CLI application version
func (b *Builder) Version(version string) *Builder {
	b.config.Apply(WithVersion(version))
	return b
}

// Description sets the CLI application description
func (b *Builder) Description(description string) *Builder {
	b.config.Apply(WithDescription(description))
	return b
}

// Author sets the CLI application author
func (b *Builder) Author(author string) *Builder {
	b.config.Apply(WithAuthor(author))
	return b
}

// Logger sets the logger for the CLI
func (b *Builder) Logger(logger *slog.Logger) *Builder {
	b.config.Apply(WithLogger(logger))
	return b
}

// DefaultTimeout sets the default command timeout
func (b *Builder) DefaultTimeout(timeout time.Duration) *Builder {
	b.config.Apply(WithDefaultTimeout(timeout))
	return b
}

// Middleware adds middleware to the CLI
func (b *Builder) Middleware(middleware ...core.Middleware) *Builder {
	b.config.Apply(WithMiddleware(middleware...))
	return b
}

// Recovery adds panic recovery middleware
func (b *Builder) Recovery() *Builder {
	b.config.Apply(WithRecovery())
	return b
}

// Logging adds logging middleware
func (b *Builder) Logging() *Builder {
	b.config.Apply(WithLogging())
	return b
}

// Timeout adds timeout middleware
func (b *Builder) Timeout(timeout time.Duration) *Builder {
	b.config.Apply(WithTimeout(timeout))
	return b
}

// ColoredOutput enables or disables colored output
func (b *Builder) ColoredOutput(enabled bool) *Builder {
	b.config.Apply(WithColoredOutput(enabled))
	return b
}

// MaxHelpWidth sets the maximum width for help text
func (b *Builder) MaxHelpWidth(width int) *Builder {
	b.config.Apply(WithMaxHelpWidth(width))
	return b
}

// ErrorHandler sets a custom error handler
func (b *Builder) ErrorHandler(handler func(error) int) *Builder {
	b.config.Apply(WithErrorHandler(handler))
	return b
}

// BeforeAll sets a hook to run before all commands
func (b *Builder) BeforeAll(hook func(*core.ExecutionContext) error) *Builder {
	b.config.Apply(WithBeforeAll(hook))
	return b
}

// AfterAll sets a hook to run after all commands
func (b *Builder) AfterAll(hook func(*core.ExecutionContext) error) *Builder {
	b.config.Apply(WithAfterAll(hook))
	return b
}

// BeforeEach sets a hook to run before each command
func (b *Builder) BeforeEach(hook func(*core.ExecutionContext) error) *Builder {
	b.config.Apply(WithBeforeEach(hook))
	return b
}

// AfterEach sets a hook to run after each command
func (b *Builder) AfterEach(hook func(*core.ExecutionContext) error) *Builder {
	b.config.Apply(WithAfterEach(hook))
	return b
}

// GlobalFlag adds a global flag
func (b *Builder) GlobalFlag(name string, value interface{}) *Builder {
	b.config.Apply(WithGlobalFlag(name, value))
	return b
}

// Build returns the configured CLI config
func (b *Builder) Build() *CLIConfig {
	return b.config
}

// Preset configurations for common use cases

// DevelopmentConfig returns a configuration suitable for development
func DevelopmentConfig() *CLIConfig {
	config := DefaultConfig()
	config.Apply(
		WithLogging(),
		WithRecovery(),
		WithColoredOutput(true),
		WithDefaultTimeout(60*time.Second),
	)
	return config
}

// ProductionConfig returns a configuration suitable for production
func ProductionConfig() *CLIConfig {
	config := DefaultConfig()
	config.Apply(
		WithRecovery(),
		WithColoredOutput(false),
		WithDefaultTimeout(30*time.Second),
	)
	return config
}

// MinimalConfig returns a minimal configuration
func MinimalConfig() *CLIConfig {
	return DefaultConfig()
}