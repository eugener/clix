package main

import (
	"context"
	"fmt"
	"os"

	"github.com/eugener/clix/app"
	"github.com/eugener/clix/config"
	"github.com/eugener/clix/core"
)

// Configuration-aware command
type ServerConfig struct {
	Host     string `posix:"h,host,Server host" yaml:"host" json:"host"`
	Port     int    `posix:"p,port,Server port,default=8080" yaml:"port" json:"port"`
	Debug    bool   `posix:"d,debug,Enable debug mode" yaml:"debug" json:"debug"`
	LogLevel string `posix:"l,log-level,Log level,choices=debug;info;warn;error" yaml:"log_level" json:"log_level"`
	Workers  int    `posix:"w,workers,Number of workers,default=4" yaml:"workers" json:"workers"`
}

func main() {
	// Create application with config file support
	application := app.NewApplicationWithOptions(
		config.WithName("configurable-cli"),
		config.WithVersion("1.0.0"),
		config.WithDescription("CLI with configuration file support"),
		config.WithRecovery(),
		config.WithLogging(),
		
		// Enable configuration file support
		config.WithConfigFile("config"), // Will look for config.yaml, config.json, etc.
		config.WithConfigPaths([]string{
			".",
			"$HOME/.config/configurable-cli",
			"/etc/configurable-cli",
		}),
		config.WithAutoLoadConfig(true),
	)
	
	// Server command
	serverCmd := core.NewCommand("serve", "Start the server with configuration support",
		func(ctx context.Context, config ServerConfig) error {
			fmt.Printf("🚀 Starting server...\n")
			fmt.Printf("   Host: %s\n", config.Host)
			fmt.Printf("   Port: %d\n", config.Port)
			fmt.Printf("   Debug: %v\n", config.Debug)
			fmt.Printf("   Log Level: %s\n", config.LogLevel)
			fmt.Printf("   Workers: %d\n", config.Workers)
			
			fmt.Println("✅ Server configuration loaded successfully!")
			return nil
		})
	
	// Config generation command
	configCmd := core.NewCommand("config", "Generate example configuration files",
		func(ctx context.Context, configData struct {
			Format  string `posix:"f,format,Configuration format,choices=yaml;json;default=yaml"`
			Command string `posix:"c,command,Command to generate config for,default=serve"`
		}) error {
			data, err := application.GenerateConfigFile(configData.Command, configData.Format)
			if err != nil {
				return err
			}
			
			fmt.Printf("# Example configuration for '%s' command\n", configData.Command)
			fmt.Printf("# Save this as config.%s in the current directory\n\n", configData.Format)
			fmt.Print(string(data))
			
			return nil
		})
	
	// Register commands
	application.Register(serverCmd)
	application.Register(configCmd)
	
	// Run application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}