package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"claude-code-test/app"
	"claude-code-test/config"
	"claude-code-test/core"
)

// Showcase all framework features
type DatabaseConfig struct {
	Host     string `posix:"h,host,Database host,default=localhost"`
	Port     int    `posix:"p,port,Database port,default=5432"`
	Username string `posix:"u,username,Database username,required,env=DB_USER"`
	Password string `posix:"w,password,Database password,env=DB_PASS"`
	Database string `posix:"d,database,Database name,required"`
	SSL      bool   `posix:"s,ssl,Enable SSL connection"`
	Timeout  int    `posix:"t,timeout,Connection timeout in seconds,default=30"`
}

type ApiConfig struct {
	URL    string   `posix:"u,url,API URL,required"`
	Method string   `posix:"m,method,HTTP method,choices=GET;POST;PUT;DELETE;default=GET"`
	Data   string   `posix:"d,data,Request data"`
	Headers []string `posix:",,,positional"`
}

func main() {
	// Create a production-ready application with all features
	application := app.NewApplicationWithOptions(
		// Basic configuration
		config.WithName("showcase"),
		config.WithVersion("3.0.0"),
		config.WithDescription("Complete CLI framework showcase with all features"),
		config.WithAuthor("Claude Code Framework Team"),
		
		// Advanced configuration
		config.WithDefaultTimeout(2*time.Minute),
		config.WithColoredOutput(true),
		config.WithMaxHelpWidth(120),
		
		// Middleware stack
		config.WithRecovery(),
		config.WithLogging(),
		config.WithTimeout(90*time.Second),
		
		// Lifecycle hooks
		config.WithBeforeAll(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Info("🚀 Application starting", 
				"version", "3.0.0",
				"pid", os.Getpid(),
				"args", ctx.Args)
			return nil
		}),
		config.WithAfterAll(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Info("✅ Application finished", 
				"duration", ctx.Duration(),
				"success", true)
			return nil
		}),
		config.WithBeforeEach(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Debug("⚡ Command starting", 
				"command", ctx.CommandName,
				"timestamp", time.Now().Format(time.RFC3339))
			return nil
		}),
		config.WithAfterEach(func(ctx *core.ExecutionContext) error {
			ctx.Logger.Debug("🏁 Command finished", 
				"command", ctx.CommandName,
				"duration", ctx.Duration())
			return nil
		}),
		
		// Custom error handler
		config.WithErrorHandler(func(err error) int {
			slog.Error("💥 Command failed", "error", err, "timestamp", time.Now())
			return 1
		}),
		
		// Global flags
		config.WithGlobalFlag("verbose", false),
		config.WithGlobalFlag("quiet", false),
	)
	
	// Register comprehensive commands
	registerShowcaseCommands(application)
	
	// Run with context
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}

func registerShowcaseCommands(application *app.Application) {
	// Database connection command with extensive validation
	dbCmd := core.NewCommand("db", "Connect to database with comprehensive options", 
		func(ctx context.Context, config DatabaseConfig) error {
			fmt.Printf("🔗 Connecting to database...\n")
			fmt.Printf("   Host: %s:%d\n", config.Host, config.Port)
			fmt.Printf("   Database: %s\n", config.Database)
			fmt.Printf("   Username: %s\n", config.Username)
			fmt.Printf("   SSL: %v\n", config.SSL)
			fmt.Printf("   Timeout: %ds\n", config.Timeout)
			
			// Simulate connection
			time.Sleep(100 * time.Millisecond)
			fmt.Println("✅ Connected successfully!")
			
			return nil
		})
	
	// API request command with choices and validation
	apiCmd := core.NewCommand("api", "Make API requests with validation and choices",
		func(ctx context.Context, config ApiConfig) error {
			fmt.Printf("🌐 Making API request...\n")
			fmt.Printf("   URL: %s\n", config.URL)
			fmt.Printf("   Method: %s\n", config.Method)
			
			if config.Data != "" {
				fmt.Printf("   Data: %s\n", config.Data)
			}
			
			if len(config.Headers) > 0 {
				fmt.Printf("   Headers: %v\n", config.Headers)
			}
			
			// Simulate API call
			time.Sleep(200 * time.Millisecond)
			fmt.Println("✅ Request completed!")
			
			return nil
		})
	
	// File processing command with complex validation
	processCmd := core.NewCommand("process", "Process files with advanced validation",
		func(ctx context.Context, config struct {
			Input      string   `posix:"i,input,Input file,required"`
			Output     string   `posix:"o,output,Output file"`
			Format     string   `posix:"f,format,Output format,choices=json;yaml;xml;csv;default=json"`
			Compress   bool     `posix:"z,compress,Compress output"`
			Concurrent int      `posix:"c,concurrent,Concurrent workers,default=4"`
			Filters    []string `posix:",,,positional"`
		}) error {
			fmt.Printf("📁 Processing file: %s\n", config.Input)
			fmt.Printf("   Format: %s\n", config.Format)
			fmt.Printf("   Workers: %d\n", config.Concurrent)
			
			if config.Output != "" {
				fmt.Printf("   Output: %s\n", config.Output)
			}
			
			if config.Compress {
				fmt.Println("   Compression: enabled")
			}
			
			if len(config.Filters) > 0 {
				fmt.Printf("   Filters: %v\n", config.Filters)
			}
			
			// Simulate processing
			for i := 0; i < config.Concurrent; i++ {
				fmt.Printf("   Worker %d: processing...\n", i+1)
				time.Sleep(50 * time.Millisecond)
			}
			
			fmt.Println("✅ Processing completed!")
			return nil
		})
	
	// Monitoring command with environment variables
	monitorCmd := core.NewCommand("monitor", "System monitoring with environment integration",
		func(ctx context.Context, config struct {
			Interval   int    `posix:"i,interval,Check interval in seconds,default=5"`
			Threshold  int    `posix:"t,threshold,Alert threshold,default=80"`
			LogFile    string `posix:"l,log,Log file path,env=MONITOR_LOG"`
			AlertEmail string `posix:"e,email,Alert email,env=ALERT_EMAIL"`
			Continuous bool   `posix:"c,continuous,Run continuously"`
		}) error {
			fmt.Printf("📊 Starting system monitor...\n")
			fmt.Printf("   Interval: %ds\n", config.Interval)
			fmt.Printf("   Threshold: %d%%\n", config.Threshold)
			
			if config.LogFile != "" {
				fmt.Printf("   Log file: %s\n", config.LogFile)
			}
			
			if config.AlertEmail != "" {
				fmt.Printf("   Alert email: %s\n", config.AlertEmail)
			}
			
			fmt.Printf("   Continuous: %v\n", config.Continuous)
			
			// Simulate monitoring
			checks := 3
			if config.Continuous {
				checks = 1 // Just one cycle for demo
			}
			
			for i := 0; i < checks; i++ {
				fmt.Printf("   Check %d: CPU 45%%, Memory 67%%, Disk 23%%\n", i+1)
				time.Sleep(time.Duration(config.Interval/3) * time.Second)
			}
			
			fmt.Println("✅ Monitoring completed!")
			return nil
		})
	
	// Configuration command showcasing defaults and choices
	configCmd := core.NewCommand("config", "Manage application configuration",
		func(ctx context.Context, config struct {
			Action string `posix:"a,action,Configuration action,choices=get;set;list;reset;required"`
			Key    string `posix:"k,key,Configuration key"`
			Value  string `posix:"v,value,Configuration value"`
			Format string `posix:"f,format,Output format,choices=json;yaml;table;default=table"`
		}) error {
			fmt.Printf("⚙️  Configuration %s\n", config.Action)
			
			switch config.Action {
			case "get":
				if config.Key == "" {
					return fmt.Errorf("key is required for get action")
				}
				fmt.Printf("   Key: %s = example_value\n", config.Key)
			case "set":
				if config.Key == "" || config.Value == "" {
					return fmt.Errorf("key and value are required for set action")
				}
				fmt.Printf("   Setting: %s = %s\n", config.Key, config.Value)
			case "list":
				fmt.Printf("   Format: %s\n", config.Format)
				fmt.Println("   Available configurations:")
				fmt.Println("   - database.host = localhost")
				fmt.Println("   - database.port = 5432")
				fmt.Println("   - api.timeout = 30")
			case "reset":
				fmt.Println("   Resetting all configurations to defaults")
			}
			
			fmt.Println("✅ Configuration updated!")
			return nil
		})
	
	// Register all commands
	application.Register(dbCmd)
	application.Register(apiCmd)
	application.Register(processCmd)
	application.Register(monitorCmd)
	application.Register(configCmd)
}