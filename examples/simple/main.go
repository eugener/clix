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
	application := app.NewApplicationWithOptions(
		config.WithName("simple-cli"),
		config.WithDescription("A simple CLI demonstration"),
		config.WithAutoLoadConfig(true),
		config.WithRecovery(),
		config.WithLogging(),
	)

	// Register commands
	application.Register(core.NewCommand("hello", "Say hello", func(ctx context.Context, config HelloConfig) error {
		greeting := fmt.Sprintf("Hello, %s!", config.Name)
		if config.Age > 0 {
			greeting += fmt.Sprintf(" You are %d years old.", config.Age)
		}
		fmt.Println(greeting)
		return nil
	}))

	application.Register(core.NewCommand("count", "Count numbers", func(ctx context.Context, config CountConfig) error {
		fmt.Printf("Counting from %d to %d:\n", config.Start, config.End)
		for i := config.Start; i <= config.End; i++ {
			fmt.Printf("%d ", i)
		}
		fmt.Println()
		return nil
	}))

	// Run the application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}