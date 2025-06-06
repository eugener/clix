# CLI Framework Examples

This directory contains examples demonstrating different aspects of the CLI framework.

## Quick Start Examples

### 1. Simple Example (`simple/`)
Basic CLI application using the traditional struct-based approach.
- Demonstrates struct configuration with POSIX tags
- Shows automatic config file loading and merging
- Perfect for learning the core concepts

```bash
cd simple && go run main.go hello --name "World"
```

### 2. Fluent API (`fluent-api/`)
Modern fluent API demonstration showing the improved developer experience.
- Method chaining for application configuration
- Simplified command creation helpers
- Comparison with the traditional approach

```bash
cd fluent-api && go run main.go
```

## Feature-Specific Examples

### 3. Configuration Loading (`config/`)
Advanced configuration management with multiple sources.
- YAML and JSON config file support
- Environment variable integration
- CLI argument precedence over config files

### 4. Interactive Mode (`interactive/`)
Interactive prompting for missing required fields.
- Automatic fallback to prompts when CLI args are missing
- Type-aware validation and suggestions
- User-friendly error messages

### 5. Advanced Features (`advanced/`)
Comprehensive example showing all framework capabilities.
- Middleware (logging, recovery, timeout)
- Hooks (before/after execution)
- Shell completion generation
- Complex command configurations

## API Comparison

### Traditional Approach (Backward Compatible)
```go
app := app.NewApplicationWithOptions(
    config.WithName("my-app"),
    config.WithRecovery(),
    config.WithLogging(),
    config.WithInteractiveMode(true),
)
app.Register(core.NewCommand("hello", "Say hello", handler))
```

### New Fluent API (Recommended)
```go
cli.New("my-app").
    Recovery().
    Logging().
    Interactive().
    WithCommands(cli.Cmd("hello", "Say hello", handler)).
    RunWithArgs(context.Background())
```

## Running Examples

Each example is a standalone Go module:

```bash
# Run any example
cd <example-name>
go run main.go [command] [flags]

# Show help for any example
go run main.go --help

# Show help for a specific command
go run main.go <command> --help
```

## Common Patterns

- **Simple Commands**: Use `cli.Cmd()` for commands with no arguments
- **Struct-Based Commands**: Use `core.NewCommand()` with struct configuration for complex flag handling
- **Configuration**: Enable `AutoConfig()` for automatic YAML/JSON config file loading
- **Development**: Use `cli.Dev()` for enhanced development experience
- **Production**: Use `cli.Prod()` for optimized production settings