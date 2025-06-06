package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"claude-code-test/app"
	"claude-code-test/config"
	"claude-code-test/core"
)

// Example command configuration
type GreetConfig struct {
	Name    string   `posix:"n,name,Name to greet,required"`
	Verbose bool     `posix:"v,verbose,Enable verbose output"`
	Count   int      `posix:"c,count,Number of greetings,default=1"`
	Tags    []string `posix:",,,positional"`
}

// Example command implementation
func main() {
	// Create application using the new app package
	application := app.NewApplicationWithOptions(
		config.WithName("example"),
		config.WithVersion("1.0.0"),
		config.WithDescription("Example CLI application using the new framework"),
		config.WithRecovery(),
		config.WithLogging(),
		config.WithTimeout(30*time.Second),
		config.WithColoredOutput(true),
	)
	
	// Create and register a command
	greetCmd := core.NewCommand("greet", "Greet someone with customizable options", func(ctx context.Context, config GreetConfig) error {
		for i := 0; i < config.Count; i++ {
			if config.Verbose {
				fmt.Printf("Greeting #%d: Hello, %s!\n", i+1, config.Name)
			} else {
				fmt.Printf("Hello, %s!\n", config.Name)
			}
		}
		
		if len(config.Tags) > 0 {
			fmt.Printf("Tags: %v\n", config.Tags)
		}
		
		return nil
	})
	
	if err := application.Register(greetCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register command: %v\n", err)
		os.Exit(1)
	}
	
	// Run the application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}