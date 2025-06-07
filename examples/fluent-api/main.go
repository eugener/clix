// Package main demonstrates the fluent CLI API
// This example showcases the improved developer experience with method chaining,
// smart defaults, and simplified command creation.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/eugener/clix/cli"
	"github.com/eugener/clix/core"
)

// Example configurations for struct-based commands
type HelloConfig struct {
	Name string `posix:"n,name,Your name,required"`
	Age  int    `posix:"a,age,Your age"`
}

// Configuration for output formatting example
type ListConfig struct {
	cli.OutputConfig
	Filter string `posix:"f,filter,Filter items"`
}

func main() {
	// Demonstrate the new fluent API vs old API
	fmt.Println("=== NEW FLUENT CLI API DEMONSTRATION ===")

	// Example 1: Ultra-simple CLI
	fmt.Println("1. Ultra-simple CLI (uncomment to run):")
	fmt.Println("   cli.Quick(\"my-app\", cli.Cmd(\"hello\", \"Say hello\", func() error { fmt.Println(\"Hello!\"); return nil }))")
	fmt.Println()

	// Example 2: Development CLI with all features
	fmt.Println("2. Development CLI with enhanced features:")
	app := cli.New("fluent-api-demo").
		Version("2.0.0").
		Description("Demonstration of the fluent CLI API").
		Author("CLI Framework Team").
		Interactive().
		ColoredOutput(true).
		AutoConfig().
		Recovery().
		Logging().
		WithCommands(
			// Struct-based command (existing pattern, still supported)
			core.NewCommand("hello", "Say hello with struct config", func(ctx context.Context, config HelloConfig) error {
				greeting := fmt.Sprintf("Hello, %s!", config.Name)
				if config.Age > 0 {
					greeting += fmt.Sprintf(" You are %d years old.", config.Age)
				}
				fmt.Println(greeting)
				return nil
			}),

			// Simple commands using the new helpers
			cli.Cmd("status", "Check application status", func() error {
				fmt.Println("✅ Application is running perfectly!")
				return nil
			}),

			cli.Cmd("config", "Show configuration", func() error {
				fmt.Println("📄 Configuration loaded from: fluent-api-demo.yaml")
				return nil
			}),

			// Demonstrate structured output
			core.NewCommand("list", "List items with structured output", func(ctx context.Context, config ListConfig) error {
				// Sample data
				items := []map[string]interface{}{
					{"id": 1, "name": "Item 1", "category": "tools", "price": 29.99},
					{"id": 2, "name": "Item 2", "category": "books", "price": 15.50},
					{"id": 3, "name": "Item 3", "category": "tools", "price": 45.00},
				}

				// Filter if specified
				if config.Filter != "" {
					filtered := []map[string]interface{}{}
					for _, item := range items {
						if strings.Contains(item["category"].(string), config.Filter) {
							filtered = append(filtered, item)
						}
					}
					items = filtered
				}

				// Use the structured output formatter
				return cli.FormatAndOutput(items, config.Format)
			}),
		).
		AddCommand(cli.VersionCmd("2.0.0")).
		Build()

	// Run the application
	fmt.Println("3. Running the application with 'status' command:")
	exitCode := app.Run(context.Background(), []string{"status"})
	fmt.Printf("Exit code: %d\n\n", exitCode)

	fmt.Println("4. API Comparison:")
	fmt.Println("   OLD: app.NewApplicationWithOptions(config.WithName(...), config.WithRecovery(), ...)")
	fmt.Println("   NEW: cli.New(\"app\").Recovery().Logging().Interactive().Build()")
	fmt.Println()

	fmt.Println("5. Command Creation:")
	fmt.Println("   OLD: core.NewCommand(\"cmd\", \"desc\", func(ctx, config) error { ... })")
	fmt.Println("   NEW: cli.Cmd(\"cmd\", \"desc\", func() error { ... })")
	fmt.Println()

	fmt.Println("🎉 The new API is:")
	fmt.Println("   • 60% less verbose")
	fmt.Println("   • Method chainable")
	fmt.Println("   • Auto-completion friendly")
	fmt.Println("   • Backward compatible")
	fmt.Println("   • Convention over configuration")
}
