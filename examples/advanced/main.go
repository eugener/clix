package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/eugener/clix/cli"
)

// Example command configurations with advanced features
type GreetConfig struct {
	Name     string   `flag:"name,n" required:"true" help:"Name to greet"`
	Verbose  bool     `flag:"verbose,v" help:"Enable verbose output"`
	Count    int      `flag:"count,c" default:"1" help:"Number of greetings"`
	Style    string   `flag:"style,s" choices:"casual,formal,friendly" help:"Greeting style"`
	Output   string   `flag:"output,o" env:"GREET_OUTPUT" help:"Output file"`
}

type GreetCommand struct{}

func (c *GreetCommand) Run(ctx context.Context, config GreetConfig) error {
	greeting := "Hello"
	switch config.Style {
	case "formal":
		greeting = "Good day"
	case "friendly":
		greeting = "Hey there"
	case "casual":
		greeting = "Hi"
	}

	message := fmt.Sprintf("%s, %s!", greeting, config.Name)
	
	for i := 0; i < config.Count; i++ {
		if config.Output != "" {
			file, err := os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open output file: %w", err)
			}
			defer file.Close()
			fmt.Fprintln(file, message)
		} else {
			fmt.Println(message)
		}
		
		if config.Verbose {
			slog.Info("Greeting sent", "iteration", i+1, "style", config.Style)
		}
	}
	
	return nil
}

type FileConfig struct {
	Input    string `flag:"input,i" required:"true" help:"Input file path"`
	Output   string `flag:"output,o" help:"Output file path"`
	Format   string `flag:"format,f" choices:"json,yaml,xml" default:"json" help:"Output format"`
	Compress bool   `flag:"compress,z" help:"Compress output"`
}

type ProcessCommand struct{}

func (c *ProcessCommand) Run(ctx context.Context, config FileConfig) error {
	if config.Compress {
		fmt.Printf("Processing %s -> %s (compressed %s)\n", config.Input, config.Output, config.Format)
	} else {
		fmt.Printf("Processing %s -> %s (%s)\n", config.Input, config.Output, config.Format)
	}
	
	// Simulate processing time
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Processing complete!")
	
	return nil
}

type ServerConfig struct {
	Port     int    `flag:"port,p" default:"8080" help:"Server port"`
	Host     string `flag:"host,h" default:"localhost" help:"Server host"`
	Debug    bool   `flag:"debug,d" help:"Enable debug mode"`
	LogLevel string `flag:"log-level,l" choices:"debug,info,warn,error" default:"info" help:"Log level"`
}

type ServerCommand struct{}

func (c *ServerCommand) Run(ctx context.Context, config ServerConfig) error {
	fmt.Printf("Starting server on %s:%d\n", config.Host, config.Port)
	fmt.Printf("Debug mode: %v, Log level: %s\n", config.Debug, config.LogLevel)
	
	// Simulate server running
	fmt.Println("Server is running... (Press Ctrl+C to stop)")
	
	select {
	case <-ctx.Done():
		fmt.Println("Server shutting down...")
		return nil
	case <-time.After(2 * time.Second):
		fmt.Println("Server stopped after demo timeout")
		return nil
	}
}

func main() {
	app := cli.New("advanced-cli").
		Version("2.0.0").
		Description("Advanced CLI framework demonstration with all features").
		Interactive().
		AutoConfig().
		Recovery().
		Logging().
		WithCommands(
			&GreetCommand{},
			&ProcessCommand{},
			&ServerCommand{},
		).
		Build()

	app.RunWithArgs(context.Background())
}