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

// Ultimate CLI showcasing all developer happiness features
type DeployConfig struct {
	Environment string   `posix:"e,env,Deployment environment,choices=dev;staging;prod,required" yaml:"environment" json:"environment"`
	Version     string   `posix:"v,version,Version to deploy,required" yaml:"version" json:"version"`
	Replicas    int      `posix:"r,replicas,Number of replicas,default=3" yaml:"replicas" json:"replicas"`
	Namespace   string   `posix:"n,namespace,Kubernetes namespace,default=default" yaml:"namespace" json:"namespace"`
	DryRun      bool     `posix:"d,dry-run,Dry run mode" yaml:"dry_run" json:"dry_run"`
	ConfigFile  string   `posix:"c,config,Config file path,env=DEPLOY_CONFIG" yaml:"config_file" json:"config_file"`
	Services    []string `posix:",,,positional" yaml:"services" json:"services"`
}

type ServerConfig struct {
	Host        string `posix:"h,host,Server host,default=localhost" yaml:"host" json:"host"`
	Port        int    `posix:"p,port,Server port,default=8080" yaml:"port" json:"port"`
	Debug       bool   `posix:"d,debug,Enable debug mode" yaml:"debug" json:"debug"`
	LogLevel    string `posix:"l,log-level,Log level,choices=debug;info;warn;error;default=info" yaml:"log_level" json:"log_level"`
	Workers     int    `posix:"w,workers,Number of workers,default=4" yaml:"workers" json:"workers"`
	SSLCert     string `posix:"cert,ssl-cert,SSL certificate path,env=SSL_CERT" yaml:"ssl_cert" json:"ssl_cert"`
	SSLKey      string `posix:"key,ssl-key,SSL key path,env=SSL_KEY" yaml:"ssl_key" json:"ssl_key"`
	DatabaseURL string `posix:"db,database,Database URL,env=DATABASE_URL" yaml:"database_url" json:"database_url"`
}

func main() {
	// Create the ultimate CLI with ALL developer happiness features enabled
	application := app.NewApplicationWithOptions(
		config.WithName("ultimate-cli"),
		config.WithVersion("2.0.0"),
		config.WithDescription("🚀 Ultimate CLI showcasing all developer happiness features"),
		config.WithAuthor("Go CLI Framework Team"),
		
		// Error handling & suggestions
		config.WithColoredOutput(true),
		config.WithMaxHelpWidth(120),
		
		// Configuration file support
		config.WithConfigFile("ultimate"),
		config.WithConfigPaths([]string{
			".",
			"$HOME/.config/ultimate-cli",
			"/etc/ultimate-cli",
		}),
		config.WithAutoLoadConfig(true),
		
		// Interactive mode
		config.WithInteractiveMode(true),
		
		// Middleware for robustness
		config.WithRecovery(),
		config.WithLogging(),
		config.WithTimeout(2*time.Minute),
		
		// Lifecycle hooks
		config.WithBeforeAll(func(ctx *core.ExecutionContext) error {
			fmt.Printf("🚀 Ultimate CLI v2.0.0 starting...\n")
			return nil
		}),
		config.WithAfterAll(func(ctx *core.ExecutionContext) error {
			fmt.Printf("✨ Command completed in %v\n", ctx.Duration())
			return nil
		}),
		
		// Enhanced error handling
		config.WithErrorHandler(func(err error) int {
			if err != nil {
				return 1
			}
			return 0
		}),
	)
	
	// Deployment command - showcases config files, choices, required fields, etc.
	deployCmd := core.NewCommand("deploy", "Deploy services with comprehensive configuration support",
		func(ctx context.Context, config DeployConfig) error {
			fmt.Printf("🚀 Deploying to %s environment...\n", config.Environment)
			fmt.Printf("   Version: %s\n", config.Version)
			fmt.Printf("   Replicas: %d\n", config.Replicas)
			fmt.Printf("   Namespace: %s\n", config.Namespace)
			fmt.Printf("   Dry Run: %v\n", config.DryRun)
			
			if config.ConfigFile != "" {
				fmt.Printf("   Config File: %s\n", config.ConfigFile)
			}
			
			if len(config.Services) > 0 {
				fmt.Printf("   Services: %v\n", config.Services)
			} else {
				fmt.Println("   Services: all services")
			}
			
			// Simulate deployment
			fmt.Println()
			fmt.Println("📦 Preparing deployment...")
			time.Sleep(500 * time.Millisecond)
			
			if config.DryRun {
				fmt.Println("🧪 Dry run completed - no actual deployment performed")
			} else {
				fmt.Println("✅ Deployment completed successfully!")
			}
			
			return nil
		})
	
	// Server command - showcases environment variables, defaults, SSL config
	serverCmd := core.NewCommand("serve", "Start server with comprehensive configuration",
		func(ctx context.Context, config ServerConfig) error {
			fmt.Printf("🌐 Starting server...\n")
			fmt.Printf("   Address: %s:%d\n", config.Host, config.Port)
			fmt.Printf("   Debug Mode: %v\n", config.Debug)
			fmt.Printf("   Log Level: %s\n", config.LogLevel)
			fmt.Printf("   Workers: %d\n", config.Workers)
			
			if config.SSLCert != "" && config.SSLKey != "" {
				fmt.Printf("   SSL: enabled (cert: %s, key: %s)\n", config.SSLCert, config.SSLKey)
			} else {
				fmt.Printf("   SSL: disabled\n")
			}
			
			if config.DatabaseURL != "" {
				fmt.Printf("   Database: connected\n")
			}
			
			fmt.Println()
			fmt.Println("✅ Server started successfully!")
			fmt.Printf("   Access at: http%s://%s:%d\n", 
				func() string { if config.SSLCert != "" { return "s" }; return "" }(),
				config.Host, config.Port)
			
			return nil
		})
	
	// Config generation command
	configCmd := core.NewCommand("config", "Generate example configuration files",
		func(ctx context.Context, config struct {
			Command string `posix:"c,command,Command to generate config for,choices=deploy;serve,default=deploy"`
			Format  string `posix:"f,format,Configuration format,choices=yaml;json,default=yaml"`
			Output  string `posix:"o,output,Output file (default: stdout)"`
		}) error {
			data, err := application.GenerateConfigFile(config.Command, config.Format)
			if err != nil {
				return err
			}
			
			output := fmt.Sprintf("# Example configuration for '%s' command\n", config.Command)
			output += fmt.Sprintf("# Save this as ultimate.%s\n\n", config.Format)
			output += string(data)
			
			if config.Output != "" {
				err := os.WriteFile(config.Output, []byte(output), 0644)
				if err != nil {
					return fmt.Errorf("failed to write config file: %w", err)
				}
				fmt.Printf("✅ Configuration file written to: %s\n", config.Output)
			} else {
				fmt.Print(output)
			}
			
			return nil
		})
	
	// Demo command showing error handling and suggestions
	demoCmd := core.NewCommand("demo", "Demonstrate error handling and suggestions",
		func(ctx context.Context, config struct {
			Action string `posix:"a,action,Action to perform,choices=hello;goodbye;test,required"`
			Name   string `posix:"n,name,Your name,required"`
			Count  int    `posix:"c,count,Repeat count,default=1"`
		}) error {
			message := ""
			switch config.Action {
			case "hello":
				message = fmt.Sprintf("Hello, %s!", config.Name)
			case "goodbye":
				message = fmt.Sprintf("Goodbye, %s!", config.Name)
			case "test":
				message = fmt.Sprintf("Testing with %s", config.Name)
			}
			
			for i := 0; i < config.Count; i++ {
				fmt.Printf("%d: %s\n", i+1, message)
			}
			
			return nil
		})
	
	// Register all commands
	application.Register(deployCmd)
	application.Register(serverCmd)
	application.Register(configCmd)
	application.Register(demoCmd)
	
	// Show enhanced help if no arguments
	if len(os.Args) == 1 {
		fmt.Println("🚀 Ultimate CLI Framework Demo")
		fmt.Println()
		fmt.Println("This CLI showcases ALL developer happiness features:")
		fmt.Println()
		fmt.Println("✨ Features Demonstrated:")
		fmt.Println("   🎯 Intelligent error messages with suggestions")
		fmt.Println("   📁 Configuration file support (YAML/JSON)")
		fmt.Println("   🤖 Interactive mode for missing required fields")
		fmt.Println("   🏷️  Environment variable integration")
		fmt.Println("   🎨 Colored output and beautiful help")
		fmt.Println("   ⚡ POSIX-compliant argument parsing")
		fmt.Println("   🛡️  Comprehensive validation with choices")
		fmt.Println("   🔧 Middleware pipeline with recovery/logging")
		fmt.Println()
		fmt.Println("🧪 Try these examples:")
		fmt.Printf("   %s demo                                    # Interactive mode\n", os.Args[0])
		fmt.Printf("   %s deploy                                  # Interactive deployment\n", os.Args[0])
		fmt.Printf("   %s deploy --env prod --version v1.2.3     # Partial config\n", os.Args[0])
		fmt.Printf("   %s serve --debug                          # Server with debug\n", os.Args[0])
		fmt.Printf("   %s config --command deploy --format yaml  # Generate config\n", os.Args[0])
		fmt.Printf("   %s invalid-command                        # See smart errors\n", os.Args[0])
		fmt.Printf("   %s deploy --invalid-flag                  # See flag suggestions\n", os.Args[0])
		fmt.Println()
		fmt.Println("📖 Use 'help <command>' for detailed information about any command")
		fmt.Println()
	}
	
	// Run application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}