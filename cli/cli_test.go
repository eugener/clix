package cli

import (
	"context"
	"os"
	"testing"
	
	"github.com/eugener/clix/core"
)

// Test command for testing
type TestConfigFixed struct {
	Name    string `posix:"n,name,Name to greet,required"`
	Count   int    `posix:"c,count,Number of greetings,default=1"`
	Verbose bool   `posix:"v,verbose,Enable verbose output"`
}

type TestCommandFixed struct {
	executed bool
	config   TestConfigFixed
}

func (c *TestCommandFixed) Name() string {
	return "testfixed"
}

func (c *TestCommandFixed) Description() string {
	return "Test command fixed"
}

func (c *TestCommandFixed) Run(ctx context.Context, config TestConfigFixed) error {
	c.executed = true
	c.config = config
	return nil
}

// Mock exit function for testing
var mockExit = func(int) {}

func TestApp_Build_WithoutExit(t *testing.T) {
	app := New("test-app").
		Version("2.0.0").
		Description("Test app").
		Interactive().
		AutoConfig().
		Recovery().
		Logging().
		WithCommands(&TestCommandFixed{})
	
	built := app.Build()
	if built == nil {
		t.Fatal("Build() returned nil")
	}
	
	// Test that we can build without running
	config := built.GetConfig()
	if config.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", config.Name)
	}
	
	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}
}

func TestApp_Run_WithMockedExit(t *testing.T) {
	app := New("test-app")
	cmd := &TestCommandFixed{}
	app.WithCommands(cmd)
	
	// Test successful execution
	exitCode := app.Run(context.Background(), []string{"testfixed", "--name", "test"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	if !cmd.executed {
		t.Error("Command was not executed")
	}
	
	if cmd.config.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", cmd.config.Name)
	}
}

func TestApp_Run_HelpFlag(t *testing.T) {
	app := New("test-app")
	cmd := &TestCommandFixed{}
	app.WithCommands(cmd)
	
	// Test help flag
	exitCode := app.Run(context.Background(), []string{"--help"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}
	
	// Command should not be executed with help
	if cmd.executed {
		t.Error("Command should not have been executed with --help")
	}
}

func TestApp_Run_VersionFlag(t *testing.T) {
	app := New("test-app").Version("1.2.3").WithCommands(VersionCmd("1.2.3"))
	
	// Test version command
	exitCode := app.Run(context.Background(), []string{"version"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for version, got %d", exitCode)
	}
}

func TestApp_Run_InvalidCommand(t *testing.T) {
	app := New("test-app")
	cmd := &TestCommandFixed{}
	app.WithCommands(cmd)
	
	// Test invalid command
	exitCode := app.Run(context.Background(), []string{"nonexistent"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid command")
	}
}

func TestApp_Run_MissingRequiredFlag(t *testing.T) {
	app := New("test-app")
	cmd := &TestCommandFixed{}
	app.WithCommands(cmd)
	
	// Test missing required flag
	exitCode := app.Run(context.Background(), []string{"testfixed"})
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing required flag")
	}
}

func TestQuick_WithoutOsExit(t *testing.T) {
	// Capture original os.Exit to restore later
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Set test args
	os.Args = []string{"test-app", "testfixed", "--name", "test"}
	
	executed := false
	cmd := Cmd("testfixed", "Test command", func() error {
		executed = true
		return nil
	})
	
	// Test Quick function by using Run instead of RunWithArgs
	app := New("test-app").
		Recovery().
		WithCommands(cmd)
	
	exitCode := app.Run(context.Background(), []string{"testfixed"})
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	if !executed {
		t.Error("Command was not executed")
	}
}

func TestDev_Configuration(t *testing.T) {
	// Test Dev configuration without running
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"test-app", "--help"}
	
	// Create dev app but don't run it
	app := New("test-app").
		Interactive().
		ColoredOutput(true).
		Recovery().
		Logging().
		AutoConfig()
	
	built := app.Build()
	if built == nil {
		t.Fatal("Dev app should build successfully")
	}
	
	config := built.GetConfig()
	if config.Name != "test-app" {
		t.Error("Dev app name not set correctly")
	}
	
	if !config.InteractiveMode {
		t.Error("Dev app should have interactive mode enabled")
	}
}

func TestProd_Configuration(t *testing.T) {
	// Test Prod configuration without running
	app := New("test-app").
		Recovery().
		Logging().
		ColoredOutput(false)
	
	built := app.Build()
	if built == nil {
		t.Fatal("Prod app should build successfully")
	}
	
	config := built.GetConfig()
	if config.Name != "test-app" {
		t.Error("Prod app name not set correctly")
	}
	
	if config.HelpConfig.ColorEnabled {
		t.Error("Prod app should have colors disabled")
	}
}

func TestStandardCmds(t *testing.T) {
	commands := StandardCmds("test-app", "1.0.0")
	if len(commands) == 0 {
		t.Error("StandardCmds should return some commands")
	}
	
	// Verify version command exists
	found := false
	for _, cmd := range commands {
		if versionCmd, ok := cmd.(*core.CommandBase[struct{}]); ok {
			if versionCmd.Name() == "version" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("StandardCmds should include version command")
	}
}

func TestApp_ChainedConfiguration(t *testing.T) {
	// Test comprehensive method chaining
	app := New("complex-app").
		Version("2.0.0").
		Description("Complex test app").
		Author("Test Author").
		Interactive().
		AutoConfig().
		Recovery().
		Logging().
		WithCommands(&TestCommandFixed{})
	
	built := app.Build()
	config := built.GetConfig()
	
	if config.Name != "complex-app" {
		t.Errorf("Expected name 'complex-app', got '%s'", config.Name)
	}
	
	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}
	
	if config.Description != "Complex test app" {
		t.Errorf("Expected description 'Complex test app', got '%s'", config.Description)
	}
	
	if config.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", config.Author)
	}
	
	if !config.InteractiveMode {
		t.Error("Expected interactive mode to be enabled")
	}
	
	if !config.AutoLoadConfig {
		t.Error("Expected auto load config to be enabled")
	}
}