package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"claude-code-test/app"
	"claude-code-test/complete"
	"claude-code-test/config"
	"claude-code-test/core"
)

// Example command configurations with advanced features
type GreetConfig struct {
	Name     string   `posix:"n,name,Name to greet,required"`
	Verbose  bool     `posix:"v,verbose,Enable verbose output"`
	Count    int      `posix:"c,count,Number of greetings,default=1"`
	Style    string   `posix:"s,style,Greeting style,choices=casual;formal;friendly"`
	Output   string   `posix:"o,output,Output file,env=GREET_OUTPUT"`
	Tags     []string `posix:",,,positional"`
}

type FileConfig struct {
	Input    string `posix:"i,input,Input file path,required"`
	Output   string `posix:"o,output,Output file path"`
	Format   string `posix:"f,format,Output format,choices=json;yaml;xml;default=json"`
	Compress bool   `posix:"z,compress,Compress output"`
}

type ServerConfig struct {
	Port     int    `posix:"p,port,Server port,default=8080"`
	Host     string `posix:"h,host,Server host,default=localhost"`
	Debug    bool   `posix:"d,debug,Enable debug mode"`
	LogLevel string `posix:"l,log-level,Log level,choices=debug;info;warn;error;default=info"`
}

func main() {
	// Handle special completion command
	if len(os.Args) > 1 && os.Args[1] == "__complete" {
		// This would be handled by the completion system
		registry := core.NewRegistry()
		registerCommands(registry)
		handler := complete.NewCompletionHandler(registry)
		handler.Handle(os.Args[2:])
		return
	}

	// Create application with functional options
	application := app.NewApplicationWithOptions(
		config.WithName("advanced-cli"),
		config.WithVersion("2.0.0"),
		config.WithDescription("Advanced CLI framework demonstration"),
		config.WithAuthor("Claude Code Framework"),
		config.WithDefaultTimeout(60*time.Second),
		config.WithColoredOutput(true),
		config.WithMaxHelpWidth(100),
		
		// Add middleware
		config.WithRecovery(),
		config.WithLogging(),
		config.WithTimeout(30*time.Second),
		
		// Add hooks
		config.WithBeforeAll(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Info("Application starting", "version", "2.0.0")
			return nil
		}),
		config.WithAfterAll(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Info("Application finished", "duration", ctx.Duration())
			return nil
		}),
		config.WithBeforeEach(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Debug("Command starting", "command", ctx.CommandName)
			return nil
		}),
		
		// Custom error handler
		config.WithErrorHandler(func(err error) int {
			if err != nil {
				slog.Error("Command failed", "error", err)
				return 1
			}
			return 0
		}),
	)

	// Register commands
	registerCommands(application.GetRegistry())

	// Run application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}

func registerCommands(registry *core.Registry) {
	// Greet command with advanced features
	greetCmd := core.NewCommand("greet", "Greet someone with customizable options and styles", 
		func(ctx context.Context, config GreetConfig) error {
			style := config.Style
			if style == "" {
				style = "casual"
			}

			var greeting string
			switch style {
			case "formal":
				greeting = fmt.Sprintf("Good day, %s", config.Name)
			case "friendly":
				greeting = fmt.Sprintf("Hey there, %s!", config.Name)
			default:
				greeting = fmt.Sprintf("Hello, %s!", config.Name)
			}

			for i := 0; i < config.Count; i++ {
				if config.Verbose {
					fmt.Printf("[%s] Greeting #%d: %s\n", style, i+1, greeting)
				} else {
					fmt.Println(greeting)
				}
			}

			if len(config.Tags) > 0 {
				fmt.Printf("Tags: %v\n", config.Tags)
			}

			if config.Output != "" {
				fmt.Printf("Output would be written to: %s\n", config.Output)
			}

			return nil
		})

	// File processing command
	fileCmd := core.NewCommand("process", "Process files with various output formats",
		func(ctx context.Context, config FileConfig) error {
			fmt.Printf("Processing file: %s\n", config.Input)
			fmt.Printf("Output format: %s\n", config.Format)
			
			if config.Output != "" {
				fmt.Printf("Output file: %s\n", config.Output)
			}
			
			if config.Compress {
				fmt.Println("Compression enabled")
			}

			// Simulate processing
			time.Sleep(100 * time.Millisecond)
			fmt.Println("Processing completed successfully!")
			
			return nil
		})

	// Server command with configuration
	serverCmd := core.NewCommand("serve", "Start a server with configurable options",
		func(ctx context.Context, config ServerConfig) error {
			fmt.Printf("Starting server on %s:%d\n", config.Host, config.Port)
			fmt.Printf("Log level: %s\n", config.LogLevel)
			
			if config.Debug {
				fmt.Println("Debug mode enabled")
			}

			// Simulate server startup
			fmt.Println("Server started successfully!")
			fmt.Println("Press Ctrl+C to stop (simulation)")
			
			return nil
		})

	// Completion generation command
	completionCmd := core.NewCommand("completion", "Generate shell completion scripts",
		func(ctx context.Context, config struct {
			Shell string `posix:"s,shell,Shell type,choices=bash;zsh;fish;required"`
		}) error {
			generator := complete.NewGenerator(registry)
			
			var script string
			switch config.Shell {
			case "bash":
				script = generator.GenerateBashCompletion("advanced-cli")
			case "zsh":
				script = generator.GenerateZshCompletion("advanced-cli")
			case "fish":
				script = generator.GenerateFishCompletion("advanced-cli")
			default:
				return fmt.Errorf("unsupported shell: %s", config.Shell)
			}
			
			fmt.Println(script)
			return nil
		})

	// Register all commands
	registry.Register(greetCmd)
	registry.Register(fileCmd)
	registry.Register(serverCmd)
	registry.Register(completionCmd)
}