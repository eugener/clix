# Modern Go CLI Framework

A powerful, type-safe, and developer-friendly CLI framework for Go with fluent API, automatic configuration management, and comprehensive developer experience features.

## 🚀 Features

### Developer Experience
- **Fluent API**: Method chaining for intuitive application building
- **Smart Defaults**: Convention over configuration with sensible presets
- **Interactive Mode**: Automatic prompting for missing required fields
- **Intelligent Errors**: Context-aware error messages with suggestions
- **Auto-completion**: Shell completion for bash, zsh, and fish

### Framework Capabilities
- **Type-Safe Commands**: Generic `Command[T]` interface with compile-time type checking
- **POSIX Compliance**: Full POSIX argument parsing with advanced flag handling
- **Configuration Management**: YAML/JSON config files with CLI override precedence
- **Environment Integration**: Automatic environment variable binding
- **Middleware Pipeline**: Composable execution with recovery, logging, and timeout
- **Modern Go**: Uses generics, slog, context, and Go 1.21+ features

## 📦 Installation

```bash
go get github.com/eugener/clix
```

## 🏃 Quick Start

### Ultra-Simple CLI (1 line)

```go
package main

import "github.com/eugener/clix/cli"

func main() {
    cli.Quick("my-app",
        cli.Cmd("hello", "Say hello", func() error {
            fmt.Println("Hello, World!")
            return nil
        }),
    )
}
```

### Fluent API (Recommended)

```go
package main

import (
    "context"
    "github.com/eugener/clix/cli"
)

func main() {
    cli.New("my-app").
        Version("1.0.0").
        Description("My awesome CLI").
        Interactive().        // Prompt for missing fields
        AutoConfig().        // Load config files automatically
        Recovery().          // Handle panics gracefully
        WithCommands(
            cli.Cmd("deploy", "Deploy application", deployHandler),
            cli.VersionCmd("1.0.0"),
        ).
        RunWithArgs(context.Background())
}
```

### Advanced Configuration

```go
package main

import (
    "context"
    "fmt"
    "github.com/eugener/clix/cli"
    "github.com/eugener/clix/core"
)

type DeployConfig struct {
    Environment string   `posix:"e,env,Environment,choices=dev;staging;prod;required"`
    Version     string   `posix:"v,version,Version to deploy,required"`
    Replicas    int      `posix:"r,replicas,Number of replicas,default=3"`
    DryRun      bool     `posix:",dry-run,Perform a dry run"`
}

func main() {
    cli.New("deploy-tool").
        Version("2.0.0").
        Description("Application deployment tool").
        Interactive().
        AutoConfig().
        WithCommands(
            core.NewCommand("deploy", "Deploy application", 
                func(ctx context.Context, config DeployConfig) error {
                    if config.DryRun {
                        fmt.Printf("DRY RUN: Would deploy %s to %s\n", 
                            config.Version, config.Environment)
                    } else {
                        fmt.Printf("Deploying %s to %s with %d replicas\n", 
                            config.Version, config.Environment, config.Replicas)
                    }
                    return nil
                }),
        ).
        RunWithArgs(context.Background())
}
```

## 📋 Configuration Management

The framework supports multiple configuration sources with proper precedence:

**CLI Arguments > Config Files > Environment Variables > Defaults**

### Config File (deploy-tool.yaml)
```yaml
environment: "staging"
version: "1.0.0"
replicas: 5
```

### Usage Examples
```bash
# Uses config file values
./deploy-tool deploy

# CLI args override config file
./deploy-tool deploy --env prod --replicas 10

# Interactive mode prompts for missing required fields
./deploy-tool deploy  # Will prompt for missing env and version
```

## 🎯 Presets for Common Scenarios

```go
// Development: interactive, colors, recovery, logging
cli.Dev("my-app", commands...)

// Production: logging, recovery, no colors, optimized
cli.Prod("my-app", commands...)

// Minimal: just basic recovery
cli.Quick("my-app", commands...)
```

## 🔧 Advanced Features

### Middleware and Hooks

```go
app := cli.New("my-app").
    Recovery().                    // Panic recovery
    Logging().                     // Command execution logging
    Timeout(30 * time.Second).     // Command timeout
    BeforeAll(startupHook).        // Run before any command
    AfterAll(cleanupHook).         // Run after any command
    BeforeEach(commandSetup).      // Run before each command
    AfterEach(commandTeardown)     // Run after each command
```

### Environment Variables

```go
type Config struct {
    APIKey    string `posix:"k,key,API key,env=API_KEY,required"`
    LogLevel  string `posix:"l,log,Log level,env=LOG_LEVEL,default=info"`
    Database  string `posix:"d,db,Database URL,env=DATABASE_URL"`
}
```

### Validation and Choices

```go
type Config struct {
    Environment string `posix:"e,env,Environment,choices=dev;staging;prod;required"`
    Port        int    `posix:"p,port,Port number,default=8080"`
    Workers     int    `posix:"w,workers,Worker count,default=4"`
}
```

## 📚 Examples

The `examples/` directory contains comprehensive demonstrations:

- **simple/**: Traditional struct-based approach
- **fluent-api/**: Modern fluent API showcase
- **config/**: Configuration file management
- **interactive/**: Interactive prompting features
- **advanced/**: Complete feature demonstration

```bash
cd examples/fluent-api
go run main.go --help
```

## 🏗️ Architecture

### Clean Package Structure

**Public API (what you import):**
- **`cli/`**: Fluent API with method chaining and smart defaults (recommended)
- **`core/`**: Advanced struct-based command configuration
- **`app/`**: Traditional application builder (backward compatible)  
- **`config/`**: Configuration options and presets

**Implementation Details:**
- **`internal/`**: All implementation details (parsing, help, prompting, etc.)

### Typical Import Patterns

**Modern approach (recommended):**
```go
import "github.com/eugener/clix/cli"          // 90% of use cases
import "github.com/eugener/clix/core"         // For struct-based commands
```

**Traditional approach (backward compatible):**
```go
import "github.com/eugener/clix/app"          // Traditional builder
import "github.com/eugener/clix/config"       // Configuration options  
import "github.com/eugener/clix/core"         // Command registration
```

### Backward Compatibility

The framework maintains full backward compatibility:

```go
// Traditional approach (still supported)
app := app.NewApplicationWithOptions(
    config.WithName("my-app"),
    config.WithRecovery(),
    config.WithLogging(),
)
app.Register(core.NewCommand("cmd", "description", handler))

// New fluent approach (recommended)
cli.New("my-app").Recovery().Logging().WithCommands(
    cli.Cmd("cmd", "description", handler),
).Build()
```

## 🧪 Testing

The framework provides testing utilities for CLI applications:

```go
func TestMyCommand(t *testing.T) {
    app := cli.New("test-app").WithCommands(myCommand)
    
    // Test command execution
    exitCode := app.Run(context.Background(), []string{"my-command", "--flag", "value"})
    assert.Equal(t, 0, exitCode)
}
```

## 📈 Performance

- **Zero allocations** in hot paths
- **Lazy evaluation** of help text and completions
- **Efficient parsing** with minimal string operations
- **Concurrent-safe** command registration and execution

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

Inspired by modern CLI frameworks like Cobra, urfave/cli, and Kingpin, but designed specifically for Go's type system and modern language features.
