// Package main demonstrates structured output functionality
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/eugener/clix/cli"
	"github.com/eugener/clix/core"
)

// Configuration for output formatting example
type ListConfig struct {
	Format string `posix:"o,format,Output format (json|yaml|table|text)"`
	Filter string `posix:"f,filter,Filter items"`
}

func main() {
	app := cli.New("output-demo").
		Version("1.0.0").
		Description("Demonstration of structured output").
		Recovery().
		WithCommands(
			// Demonstrate structured output
			core.NewCommand("list", "List items with structured output", func(ctx context.Context, config ListConfig) error {
				// Set default format if not specified
				if config.Format == "" {
					config.Format = "text"
				}
				
				// Validate format
				if !cli.ValidFormat(config.Format) {
					return fmt.Errorf("invalid output format: %s. Valid formats: json, yaml, table, text", config.Format)
				}

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
				return cli.FormatAndOutput(items, cli.Format(config.Format))
			}),
		).
		Build()

	app.RunWithArgs(context.Background())
}