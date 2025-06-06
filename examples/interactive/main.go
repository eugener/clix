package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/eugener/clix/app"
	"github.com/eugener/clix/config"
	"github.com/eugener/clix/core"
)

// Interactive command configuration
type UserConfig struct {
	Name     string `posix:"n,name,Your full name,required"`
	Email    string `posix:"e,email,Email address,required"`
	Age      int    `posix:"a,age,Your age,required"`
	Role     string `posix:"r,role,Your role,choices=admin;user;guest,required"`
	Bio      string `posix:"b,bio,Short bio"`
	Active   bool   `posix:"active,,Account active"`
	Tags     []string `posix:",,,positional"`
}

type DatabaseConfig struct {
	Host     string `posix:"h,host,Database host,required"`
	Port     int    `posix:"p,port,Database port,default=5432"`
	Database string `posix:"d,database,Database name,required"`
	Username string `posix:"u,username,Database username,required"`
	Password string `posix:"w,password,Database password,required"`
	SSL      bool   `posix:"s,ssl,Enable SSL connection"`
}

func main() {
	// Create application with interactive mode enabled
	application := app.NewApplicationWithOptions(
		config.WithName("interactive-cli"),
		config.WithVersion("1.0.0"),
		config.WithDescription("CLI with interactive mode for missing required fields"),
		config.WithRecovery(),
		config.WithLogging(),
		
		// Enable interactive mode
		config.WithInteractiveMode(true),
		
		// Also enable config file support for completeness
		config.WithConfigFile("config"),
		config.WithAutoLoadConfig(true),
	)
	
	// User creation command
	userCmd := core.NewCommand("user", "Create a new user with interactive prompts",
		func(ctx context.Context, config UserConfig) error {
			fmt.Printf("🎉 Creating user...\n")
			fmt.Printf("   Name: %s\n", config.Name)
			fmt.Printf("   Email: %s\n", config.Email)
			fmt.Printf("   Age: %d\n", config.Age)
			fmt.Printf("   Role: %s\n", config.Role)
			
			if config.Bio != "" {
				fmt.Printf("   Bio: %s\n", config.Bio)
			}
			
			fmt.Printf("   Active: %v\n", config.Active)
			
			if len(config.Tags) > 0 {
				fmt.Printf("   Tags: %v\n", config.Tags)
			}
			
			fmt.Println("✅ User created successfully!")
			return nil
		})
	
	// Database connection command
	dbCmd := core.NewCommand("connect", "Connect to database with interactive prompts",
		func(ctx context.Context, config DatabaseConfig) error {
			fmt.Printf("🔗 Connecting to database...\n")
			fmt.Printf("   Host: %s:%d\n", config.Host, config.Port)
			fmt.Printf("   Database: %s\n", config.Database)
			fmt.Printf("   Username: %s\n", config.Username)
			fmt.Printf("   Password: %s\n", strings.Repeat("*", len(config.Password)))
			fmt.Printf("   SSL: %v\n", config.SSL)
			
			fmt.Println("✅ Database connection configured!")
			return nil
		})
	
	// Demo command showing interactive mode
	demoCmd := core.NewCommand("demo", "Demonstrate interactive mode features",
		func(ctx context.Context, config struct {
			Message string `posix:"m,message,Welcome message,required"`
			Count   int    `posix:"c,count,Repeat count,default=1"`
		}) error {
			for i := 0; i < config.Count; i++ {
				fmt.Printf("%d: %s\n", i+1, config.Message)
			}
			return nil
		})
	
	// Register commands
	application.Register(userCmd)
	application.Register(dbCmd)
	application.Register(demoCmd)
	
	// Show usage information
	if len(os.Args) == 1 {
		fmt.Println("🤖 Interactive CLI Example")
		fmt.Println()
		fmt.Println("This CLI demonstrates interactive mode - when you miss required fields,")
		fmt.Println("it will automatically prompt you for the missing information!")
		fmt.Println()
		fmt.Println("Try these commands:")
		fmt.Printf("  %s user                    # Will prompt for all required fields\n", os.Args[0])
		fmt.Printf("  %s user --name Alice       # Will prompt for remaining required fields\n", os.Args[0])
		fmt.Printf("  %s connect                 # Will prompt for database connection info\n", os.Args[0])
		fmt.Printf("  %s demo                    # Will prompt for message\n", os.Args[0])
		fmt.Println()
	}
	
	// Run application
	ctx := context.Background()
	exitCode := application.RunWithArgs(ctx)
	os.Exit(exitCode)
}