# Modern Go CLI Framework

A powerful, type-safe, and developer-friendly CLI framework for Go with fluent API, automatic configuration management, and comprehensive developer experience features.

## 🆕 What's New in v2.0

**✨ Professional Visual Experience**
- **Beautiful Unicode Tables**: Rich table formatting with proper borders and alignment
- **Progress Indicators**: Real-time progress bars with ETA calculation and completion tracking  
- **Multiple Spinner Styles**: 6 predefined spinner animations for different use cases
- **Structured Output**: JSON, YAML, Table, and Text formats with single API call
- **Command Aliases**: Intuitive shortcuts for frequently used commands
- **Enhanced Error Messages**: Beautiful, contextual error messages with suggestions

**🚀 Key Improvements**
- **Enhanced User Experience**: Visual feedback makes long-running commands professional
- **Better Error Handling**: Context-aware error messages with smart suggestions and beautiful formatting
- **Command Shortcuts**: Aliases support for improved productivity (e.g., `deploy`, `d`, `dep`)
- **Better Integration**: Progress indicators work seamlessly with structured output
- **Developer Friendly**: Simple APIs with powerful customization options
- **Production Ready**: All features thoroughly tested with comprehensive examples

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

### ✨ New in v2.0: Visual & Output Features
- **🎨 Structured Output**: JSON, YAML, Table, and Text formats with beautiful Unicode tables
- **📊 Progress Indicators**: Progress bars and spinners with ETA calculation and customizable styles
- **🎯 Rich UI Components**: Professional visual feedback for long-running operations
- **⚡ Command Aliases**: Support for command shortcuts and alternative names
- **🚨 Enhanced Error Messages**: Beautiful, contextual error messages with smart suggestions

## 📦 Installation

```bash
go install github.com/eugener/clix@latest
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

### Advanced Configuration with Modern Features

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/eugener/clix/cli"
    "github.com/eugener/clix/core"
)

type DeployConfig struct {
    Environment string   `posix:"e,env,Environment,choices=dev;staging;prod;required"`
    Version     string   `posix:"v,version,Version to deploy,required"`
    Replicas    int      `posix:"r,replicas,Number of replicas,default=3"`
    DryRun      bool     `posix:",dry-run,Perform a dry run"`
    Format      string   `posix:"f,format,Output format,default=table"`
}

func main() {
    cli.New("deploy-tool").
        Version("2.0.0").
        Description("Application deployment tool with modern UI").
        Interactive().
        AutoConfig().
        TableOutput().  // Set default output format
        WithCommands(
            core.NewCommand("deploy", "Deploy application with progress tracking", 
                func(ctx context.Context, config DeployConfig) error {
                    // Create progress bar for deployment steps
                    pb := cli.NewProgressBar("Deploying application", 5)
                    pb.Start()
                    defer pb.Finish()

                    // Simulate deployment steps
                    steps := []string{
                        "Validating configuration...",
                        "Building container...",
                        "Pushing to registry...",
                        "Updating deployment...",
                        "Verifying rollout...",
                    }

                    for i, step := range steps {
                        pb.UpdateTitle(step)
                        time.Sleep(time.Second) // Simulate work
                        pb.Update(i + 1)
                    }

                    // Generate deployment results
                    result := map[string]interface{}{
                        "environment": config.Environment,
                        "version":     config.Version,
                        "replicas":    config.Replicas,
                        "status":      "success",
                        "deployed_at": time.Now().Format(time.RFC3339),
                    }

                    // Output results in requested format
                    return cli.FormatAndOutput(result, cli.Format(config.Format))
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

### ✨ Structured Output Support

Generate beautiful output in multiple formats with a single API:

```go
// Support multiple output formats
type ListConfig struct {
    Format string `posix:"f,format,Output format (json|yaml|table|text),default=table"`
}

func listHandler(ctx context.Context, config ListConfig) error {
    data := []map[string]interface{}{
        {"id": 1, "name": "Server 1", "status": "running", "cpu": "45%"},
        {"id": 2, "name": "Server 2", "status": "stopped", "cpu": "0%"},
    }
    
    // Automatically format output based on user preference
    return cli.FormatAndOutput(data, cli.Format(config.Format))
}

// Fluent API output configuration
app := cli.New("my-app").
    TableOutput().      // Default to table format
    JSONOutput().       // Or default to JSON
    YAMLOutput().       // Or default to YAML
```

**Output Examples:**

```bash
# Beautiful Unicode table (default)
./app list
┌────┬──────────┬─────────┬─────┐
│ id │ name     │ status  │ cpu │
├────┼──────────┼─────────┼─────┤
│ 1  │ Server 1 │ running │ 45% │
│ 2  │ Server 2 │ stopped │ 0%  │
└────┴──────────┴─────────┴─────┘

# JSON for automation
./app list --format json
[
  {"id": 1, "name": "Server 1", "status": "running", "cpu": "45%"},
  {"id": 2, "name": "Server 2", "status": "stopped", "cpu": "0%"}
]

# YAML for configuration
./app list --format yaml
- id: 1
  name: Server 1
  status: running
  cpu: 45%
```

### 📊 Progress Indicators & UI Components

Add professional visual feedback for long-running operations:

```go
// Progress bars with ETA calculation
func processFiles(files []string) error {
    pb := cli.NewProgressBar("Processing files", len(files))
    pb.Start()
    defer pb.Finish()
    
    for i, file := range files {
        // Do work
        processFile(file)
        pb.Update(i + 1)
    }
    return nil
}

// Spinners for unknown duration tasks
func connectToAPI() error {
    spinner := cli.NewSpinner("Connecting to API...")
    spinner.Start()
    defer spinner.Stop()
    
    // Phases of work with title updates
    spinner.UpdateTitle("Authenticating...")
    authenticate()
    
    spinner.UpdateTitle("Fetching data...")
    return fetchData()
}

// Multiple spinner styles available
spinner := cli.NewSpinner("Loading...", 
    cli.WithSpinnerFrames(cli.SpinnerCircle),     // ◐ ◓ ◑ ◒
    cli.WithSpinnerFrames(cli.SpinnerArrows),     // ← ↖ ↑ ↗ → ↘ ↓ ↙
    cli.WithSpinnerFrames(cli.SpinnerDots),       // ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
)

// Convenience wrappers for automatic progress handling
handler := cli.WithProgress("Processing data", 100, func(config T, pb *cli.ProgressBar) error {
    // Progress bar automatically managed
    for i := 0; i < 100; i++ {
        // do work
        pb.Update(i + 1)
    }
    return nil
})
```

### ⚡ Command Aliases

Create intuitive shortcuts for frequently used commands:

```go
// Using CommandBuilder (fluent API)
cmd := cli.NewCommandBuilder("deploy", "Deploy the application").
    WithAliases("d", "dep").
    WithHandler(func() error {
        fmt.Println("Deploying...")
        return nil
    }).
    Build()

// Using helper function
cmd := cli.CmdWithAliases("list", "List items", []string{"ls", "l"}, func() error {
    // List items
    return nil
})

// Multiple aliases for complex commands
cmd := cli.NewCommandBuilder("kubernetes-deploy", "Deploy to Kubernetes").
    WithAliases("k8s-deploy", "k8s", "deploy").
    WithHandler(deployHandler).
    Build()
```

**Automatic Help Integration:**
```bash
# Help automatically shows aliases
$ myapp help
Commands:
  deploy, d, dep          Deploy the application
  list, ls, l            List items
  kubernetes-deploy, k8s  Deploy to Kubernetes

# All aliases work identically
$ myapp deploy --env prod
$ myapp d --env prod      # Same as above
$ myapp dep --env prod    # Same as above
```

### 🚨 Enhanced Error Messages

Get beautiful, contextual error messages with smart suggestions:

```bash
# Unknown command with suggestions
$ myapp deploi
❌ Unknown command: 'deploi'

💡 Did you mean:
   → deploy
   → delete

📋 Available commands:
   deploy, d, dep
   list, ls, l
   help

💡 Try 'myapp help' to see available commands

# Missing required flags with examples
$ myapp deploy
❌ Missing required flag: --env

🔴 Required flags:
   ✗ (missing) --env
   ✗ --region

📝 Examples:
   $ myapp deploy --env prod --region us-west-2
   $ myapp help deploy

💬 Use 'myapp help deploy' for detailed usage

# Configuration errors with troubleshooting
$ myapp start --config invalid.yaml
❌ Configuration error: configuration file not found

💡 Configuration troubleshooting:
   • Check configuration file syntax (YAML/JSON)
   • Verify file permissions
   • Ensure required configuration values are set

📝 Configuration examples:
   config.yaml
   config.json

💬 Use 'myapp help start' for detailed help
```

**Error Types Supported:**
- **Unknown Commands**: Smart suggestions using Levenshtein distance
- **Missing Required Fields**: Clear indication of what's needed
- **Command Conflicts**: Helpful explanations for alias conflicts
- **Configuration Errors**: Actionable troubleshooting steps
- **Invalid Values**: Context-aware validation messages

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
- **fluent-api/**: Modern fluent API showcase with structured output
- **config/**: Configuration file management
- **interactive/**: Interactive prompting features
- **advanced/**: Complete feature demonstration with aliases and error handling
- **output-demo/**: Structured output formats demonstration
- **progress-demo/**: Progress bars and spinners showcase

```bash
# Try the new visual features
cd examples/progress-demo
go run main.go process --count 10 --delay 200ms
go run main.go export --format table --items 5

cd examples/output-demo  
go run main.go list --format json
go run main.go list --format table

# Test command aliases and error messages
cd examples/advanced
go run main.go deploy --env prod    # Main command
go run main.go d --env prod         # Alias
go run main.go deploi               # See enhanced error message
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
