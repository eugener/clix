// Package main demonstrates the traditional struct-based CLI approach
// This example shows the core concepts of the CLI framework using the stable,
// backward-compatible API with struct-based command configuration.
package main

import (
	"context"
	"fmt"
	"os"

	"claude-code-test/app"
	"claude-code-test/config"
	"claude-code-test/core"
)

// Simple command configurations
type HelloConfig struct {
	Name string `posix:"n,name,Your name,required"`
	Age  int    `posix:"a,age,Your age"`
}

type CountConfig struct {
	Start int `posix:"s,start,Start number,default=1"`
	End   int `posix:"e,end,End number,default=10"`
}

func main() {
	// Create app with config file support enabled
	// This demonstrates the traditional approach - fully explicit configuration
	application := app.NewApplicationWithOptions(
		config.WithName("simple-cli"),
		config.WithDescription("Traditional struct-based CLI demonstration"),
		config.WithAutoLoadConfig(true), // Enable YAML/JSON config file loading
		config.WithRecovery(),           // Enable panic recovery
		config.WithLogging(),            // Enable command execution logging
	)

	// Register commands using struct-based configuration
	// Each command gets its own configuration struct with POSIX tags
	application.Register(core.NewCommand("hello", "Say hello with personalized greeting", func(ctx context.Context, config HelloConfig) error {
		greeting := fmt.Sprintf("Hello, %s!", config.Name)
		if config.Age > 0 {
			greeting += fmt.Sprintf(" You are %d years old.", config.Age)
		}
		fmt.Println(greeting)
		return nil
	}))

	application.Register(core.NewCommand("count", "Count numbers in a range", func(ctx context.Context, config CountConfig) error {
		fmt.Printf("Counting from %d to %d:\n", config.Start, config.End)
		for i := config.Start; i <= config.End; i++ {
			fmt.Printf("%d ", i)
		}
		fmt.Println()
		return nil
	}))

	// Run the application
	// The framework will handle argument parsing, config file loading,
	// command execution, and error handling automatically
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}

// Example usage:
//   go run main.go hello --name "Alice" --age 30
//   go run main.go count --start 1 --end 5
//   go run main.go --help