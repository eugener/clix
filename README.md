# Go CLI Framework

A modern, type-safe, and feature-rich CLI framework for Go that leverages the latest language features including generics, structured logging, and advanced reflection.

## 🚀 Features

- **Type-Safe Commands**: Generic `Command[T]` interface with compile-time type checking
- **POSIX Compliance**: Full POSIX argument parsing with bundled flags, `--` handling, and `=` syntax
- **Zero Dependencies**: Built using only Go standard library
- **Middleware Pipeline**: Composable execution pipeline with built-in recovery, logging, and timeout
- **Auto-Generated Help**: Rich help text generated from struct tags with customizable templates
- **Shell Completion**: Built-in support for bash, zsh, and fish auto-completion
- **Environment Integration**: Automatic environment variable binding
- **Advanced Validation**: Comprehensive validation with custom validators and choices
- **Modern Go**: Uses generics, slog, context, and other Go 1.21+ features

## 📦 Installation

```bash
go get github.com/yourorg/go-cli-framework
```

## 🏃 Quick Start

### Simple CLI (3 lines)

```go
package main

import (
    "context"
    "fmt"
    "os"
    "github.com/yourorg/go-cli-framework/app"
    "github.com/yourorg/go-cli-framework/core"
)

type HelloConfig struct {
    Name string `posix:"n,name,Your name,required"`
    Age  int    `posix:"a,age,Your age"`
}

func main() {
    application := app.QuickApp("mycli", "A simple CLI example",
        core.NewCommand("hello", "Say hello", func(ctx context.Context, config HelloConfig) error {
            fmt.Printf("Hello, %s!", config.Name)
            if config.Age > 0 {
                fmt.Printf(" You are %d years old.", config.Age)
            }
            fmt.Println()
            return nil
        }),
    )
    
    exitCode := application.RunWithArgs(context.Background())
    os.Exit(exitCode)
}
```

### Usage

```bash
$ mycli hello --name Alice --age 25
Hello, Alice! You are 25 years old.

$ mycli hello -n Bob
Hello, Bob!

$ mycli help hello
Say hello

Usage:
  mycli hello [options]

Options:
  -a, --age <int>
        Your age
  -n, --name <string>
        Your name (required)
```

## 📖 Complete Documentation

### Table of Contents

1. [Command Definition](#command-definition)
2. [Struct Tags](#struct-tags)
3. [Application Configuration](#application-configuration)
4. [Middleware](#middleware)
5. [Help System](#help-system)
6. [Validation](#validation)
7. [Shell Completion](#shell-completion)
8. [Advanced Features](#advanced-features)
9. [Examples](#examples)

## Command Definition

### Basic Command

```go
type MyConfig struct {
    Input  string `posix:"i,input,Input file,required"`
    Output string `posix:"o,output,Output file"`
    Verbose bool  `posix:"v,verbose,Enable verbose output"`
}

cmd := core.NewCommand("process", "Process files", 
    func(ctx context.Context, config MyConfig) error {
        if config.Verbose {
            fmt.Printf("Processing %s -> %s\n", config.Input, config.Output)
        }
        // Your logic here
        return nil
    })
```

### Generic Type Safety

The framework uses Go generics to provide compile-time type safety:

```go
// This is type-safe - config parameter is guaranteed to be MyConfig
func(ctx context.Context, config MyConfig) error {
    // config.Input is guaranteed to be a string
    // config.Verbose is guaranteed to be a bool
    // All validation rules are already applied
    return nil
}
```

## Struct Tags

The framework uses the `posix` struct tag to define command-line arguments:

### Tag Format

```
`posix:"short,long,description,flags"`
```

### Tag Examples

```go
type Config struct {
    // Basic flag
    Name string `posix:"n,name,Your name"`
    
    // Required flag
    Email string `posix:"e,email,Email address,required"`
    
    // Flag with default value
    Port int `posix:"p,port,Server port,default=8080"`
    
    // Flag with choices
    Format string `posix:"f,format,Output format,choices=json;yaml;xml"`
    
    // Environment variable binding
    Token string `posix:"t,token,API token,env=API_TOKEN"`
    
    // Hidden flag (not shown in help)
    Debug bool `posix:"d,debug,Debug mode,hidden"`
    
    // Positional arguments
    Files []string `posix:",,,positional"`
}
```

### Available Flags

- `required` - Field must be provided
- `default=value` - Default value if not provided
- `choices=a;b;c` - Restrict to specific values
- `env=VAR_NAME` - Bind to environment variable
- `hidden` - Hide from help text
- `positional` - Positional argument (for slices)

## Application Configuration

### Quick Applications

```go
// Development-friendly app
app := app.DevApp("myapp", "Description", commands...)

// Production-ready app
app := app.ProdApp("myapp", "Description", commands...)

// Minimal app
app := app.QuickApp("myapp", "Description", commands...)
```

### Advanced Configuration

```go
app := app.NewApplicationWithOptions(
    config.WithName("myapp"),
    config.WithVersion("1.0.0"),
    config.WithDescription("My awesome CLI"),
    config.WithAuthor("Your Name"),
    
    // Middleware
    config.WithRecovery(),
    config.WithLogging(),
    config.WithTimeout(30*time.Second),
    
    // Help customization
    config.WithColoredOutput(true),
    config.WithMaxHelpWidth(100),
    
    // Lifecycle hooks
    config.WithBeforeAll(func(ctx *core.ExecutionContext) error {
        log.Println("App starting")
        return nil
    }),
    config.WithAfterEach(func(ctx *core.ExecutionContext) error {
        log.Printf("Command %s took %v", ctx.CommandName, ctx.Duration())
        return nil
    }),
    
    // Error handling
    config.WithErrorHandler(func(err error) int {
        log.Printf("Error: %v", err)
        return 1
    }),
)
```

### Builder Pattern

```go
app := app.NewApplicationBuilder().
    Name("myapp").
    Version("1.0.0").
    Description("My CLI").
    Recovery().
    Logging().
    Timeout(30*time.Second).
    AddCommand(cmd1).
    AddCommand(cmd2).
    Build()
```

## Middleware

### Built-in Middleware

```go
// Panic recovery
config.WithRecovery()

// Request logging
config.WithLogging()

// Command timeout
config.WithTimeout(30*time.Second)
```

### Custom Middleware

```go
func MetricsMiddleware(next core.ExecuteFunc) core.ExecuteFunc {
    return func(ctx *core.ExecutionContext) error {
        start := time.Now()
        err := next(ctx)
        duration := time.Since(start)
        
        // Record metrics
        recordCommandMetrics(ctx.CommandName, duration, err)
        
        return err
    }
}

// Use it
app := app.NewApplicationWithOptions(
    config.WithMiddleware(MetricsMiddleware),
)
```

## Help System

### Automatic Help Generation

Help text is automatically generated from struct tags:

```bash
$ myapp help command
Process files

Usage:
  myapp process [options] [FILES...]

Options:
  -f, --format <string>
        Output format (choices: json, yaml, xml) (default: json)
  -i, --input <string>
        Input file (required)
  -o, --output <string>
        Output file
  -v, --verbose
        Enable verbose output

Arguments:
  FILES <[]string>
        Files to process
```

### Custom Help Templates

```go
helpConfig := help.DefaultHelpConfig("myapp")
helpConfig.UsageTemplate = `
{{.Description}}

Usage: {{.Usage}}

{{if .Flags}}Options:{{range .Flags}}
  {{.Short}}, {{.Long}}  {{.Description}}{{end}}
{{end}}
`

app := app.NewApplicationWithOptions(
    config.WithHelpConfig(helpConfig),
)
```

## Validation

### Built-in Validation

- Required fields
- Type validation
- Choice validation
- Default value application

### Custom Validators

```go
validator := help.NewValidatorRegistry()
validator.Register("email", help.EmailValidator)
validator.Register("url", help.URLValidator)
validator.Register("range", help.RangeValidator(1, 100))

// Use in validation
validator.Validate(config)
```

### Example Validation Errors

```bash
$ myapp process --format invalid
Error: validation failed: field Format must be one of: json, yaml, xml

$ myapp process
Error: validation failed: required field Input is missing
```

## Shell Completion

### Generate Completion Scripts

```go
// Add completion command to your app
completionCmd := core.NewCommand("completion", "Generate shell completion",
    func(ctx context.Context, config struct {
        Shell string `posix:"s,shell,Shell type,choices=bash;zsh;fish,required"`
    }) error {
        generator := complete.NewGenerator(registry)
        
        switch config.Shell {
        case "bash":
            fmt.Print(generator.GenerateBashCompletion("myapp"))
        case "zsh":
            fmt.Print(generator.GenerateZshCompletion("myapp"))
        case "fish":
            fmt.Print(generator.GenerateFishCompletion("myapp"))
        }
        return nil
    })
```

### Install Completions

```bash
# Bash
source <(myapp completion --shell bash)

# Zsh  
myapp completion --shell zsh > ~/.zsh/completions/_myapp

# Fish
myapp completion --shell fish > ~/.config/fish/completions/myapp.fish
```

### Smart Completions

The framework provides intelligent completions:

- Command names
- Flag names (short and long)
- Flag values based on choices
- File paths for string arguments
- Boolean values (true/false)

## Advanced Features

### Environment Variables

```go
type Config struct {
    Token string `posix:"t,token,API token,env=API_TOKEN"`
    Debug bool   `posix:"d,debug,Debug mode,env=DEBUG"`
}
```

Usage:
```bash
# Environment variable is used if flag not provided
$ API_TOKEN=secret myapp command

# Flag overrides environment variable
$ API_TOKEN=secret myapp command --token override
```

### Positional Arguments

```go
type Config struct {
    Required string   `posix:"r,required,Required flag,required"`
    Files    []string `posix:",,,positional"`
}
```

Usage:
```bash
$ myapp command --required value file1.txt file2.txt file3.txt
```

### Bundled Short Flags

```bash
# These are equivalent
$ myapp command -v -d -f json
$ myapp command -vdf json
$ myapp command -vd --format json
```

### POSIX Compliance

- `--` stops flag parsing
- `-` is treated as a filename (stdin/stdout)
- `--flag=value` syntax supported
- Short flags can be bundled
- Long flags can use `=` or space

### Context Integration

```go
func myCommand(ctx context.Context, config MyConfig) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Use context for timeouts, cancellation, etc.
    return performWork(ctx, config)
}
```

## Examples

### File Processing CLI

```go
package main

import (
    "context"
    "fmt"
    "os"
    "github.com/yourorg/go-cli-framework/app"
    "github.com/yourorg/go-cli-framework/config"
    "github.com/yourorg/go-cli-framework/core"
)

type ProcessConfig struct {
    Input      string   `posix:"i,input,Input file,required"`
    Output     string   `posix:"o,output,Output file"`
    Format     string   `posix:"f,format,Output format,choices=json;yaml;xml;default=json"`
    Compress   bool     `posix:"z,compress,Compress output"`
    Workers    int      `posix:"w,workers,Number of workers,default=4"`
    Filters    []string `posix:",,,positional"`
}

func main() {
    app := app.NewApplicationWithOptions(
        config.WithName("fileproc"),
        config.WithVersion("1.0.0"),
        config.WithDescription("File processing utility"),
        config.WithRecovery(),
        config.WithLogging(),
    )
    
    processCmd := core.NewCommand("process", "Process files with options",
        func(ctx context.Context, config ProcessConfig) error {
            fmt.Printf("Processing %s -> %s\n", config.Input, config.Output)
            fmt.Printf("Format: %s, Workers: %d\n", config.Format, config.Workers)
            
            if config.Compress {
                fmt.Println("Compression enabled")
            }
            
            if len(config.Filters) > 0 {
                fmt.Printf("Filters: %v\n", config.Filters)
            }
            
            // Your processing logic here
            return nil
        })
    
    app.Register(processCmd)
    
    exitCode := app.RunWithArgs(context.Background())
    os.Exit(exitCode)
}
```

### API Client CLI

```go
type APIConfig struct {
    URL     string            `posix:"u,url,API URL,required"`
    Method  string            `posix:"m,method,HTTP method,choices=GET;POST;PUT;DELETE;default=GET"`
    Headers map[string]string `posix:"H,header,HTTP headers"`
    Data    string            `posix:"d,data,Request data"`
    Token   string            `posix:"t,token,Auth token,env=API_TOKEN"`
}

apiCmd := core.NewCommand("request", "Make API requests",
    func(ctx context.Context, config APIConfig) error {
        client := &http.Client{Timeout: 30 * time.Second}
        
        req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, 
            strings.NewReader(config.Data))
        if err != nil {
            return err
        }
        
        if config.Token != "" {
            req.Header.Set("Authorization", "Bearer "+config.Token)
        }
        
        resp, err := client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        // Handle response
        return nil
    })
```

### Database CLI

```go
type DBConfig struct {
    Host     string `posix:"h,host,Database host,default=localhost"`
    Port     int    `posix:"p,port,Database port,default=5432"`
    Database string `posix:"d,database,Database name,required"`
    Username string `posix:"u,username,Username,required,env=DB_USER"`
    Password string `posix:"w,password,Password,env=DB_PASS"`
    SSL      bool   `posix:"s,ssl,Enable SSL"`
    Query    string `posix:"q,query,SQL query"`
}

dbCmd := core.NewCommand("query", "Execute database queries",
    func(ctx context.Context, config DBConfig) error {
        connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s",
            config.Host, config.Port, config.Username, config.Database)
        
        if config.Password != "" {
            connStr += " password=" + config.Password
        }
        
        if config.SSL {
            connStr += " sslmode=require"
        }
        
        // Connect and execute query
        return nil
    })
```

## 🏗️ Architecture

### Package Structure

```
go-cli-framework/
├── app/           # Application wrapper and lifecycle management
├── bind/          # Struct tag reflection and value binding
├── complete/      # Shell completion generation
├── config/        # Functional options configuration
├── core/          # Core interfaces and execution engine
├── help/          # Help generation and validation
└── posix/         # POSIX-compliant argument parser
```

### Key Components

- **Command[T]**: Generic interface for type-safe commands
- **Registry**: Command registration and lookup
- **Executor**: Command execution with middleware pipeline
- **Parser**: POSIX-compliant argument parsing
- **Binder**: Struct tag-based value binding
- **Generator**: Help text and completion generation
- **Validator**: Comprehensive validation system

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by modern CLI frameworks like Cobra and urfave/cli
- Built with Go 1.21+ features including generics and slog
- Designed for type safety and developer experience

## 📊 Performance

- Zero-allocation argument parsing in hot paths
- Minimal reflection usage with smart caching
- Efficient middleware pipeline
- Fast help text generation

## 🔒 Security

- Input validation and sanitization
- Safe reflection usage
- No unsafe operations
- Environment variable protection

---

**Happy CLI building! 🚀**