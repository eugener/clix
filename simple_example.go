package main

import (
	"context"
	"fmt"
	"os"

	"claude-code-test/app"
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
	// Create a quick app with minimal setup
	application := app.QuickApp("simple-cli", "A simple CLI demonstration",
		// Hello command
		core.NewCommand("hello", "Say hello", func(ctx context.Context, config HelloConfig) error {
			greeting := fmt.Sprintf("Hello, %s!", config.Name)
			if config.Age > 0 {
				greeting += fmt.Sprintf(" You are %d years old.", config.Age)
			}
			fmt.Println(greeting)
			return nil
		}),

		// Count command
		core.NewCommand("count", "Count numbers", func(ctx context.Context, config CountConfig) error {
			fmt.Printf("Counting from %d to %d:\n", config.Start, config.End)
			for i := config.Start; i <= config.End; i++ {
				fmt.Printf("%d ", i)
			}
			fmt.Println()
			return nil
		}),
	)

	// Run the application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}