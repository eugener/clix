package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/eugener/clix/core"
)

func TestCommandAliases(t *testing.T) {
	// Create a command with aliases using the unified API
	executed := false
	cmd := CmdWithAliases("deploy", "Deploy the application", []string{"d", "dep"}, func() error {
		executed = true
		return nil
	})

	// Create registry and register command
	registry := core.NewRegistry()
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Test that main command works
	executed = false
	err = registry.Execute(context.Background(), "deploy", struct{}{})
	if err != nil {
		t.Errorf("Failed to execute main command: %v", err)
	}
	if !executed {
		t.Error("Main command was not executed")
	}

	// Test that first alias works
	executed = false
	err = registry.Execute(context.Background(), "d", struct{}{})
	if err != nil {
		t.Errorf("Failed to execute alias 'd': %v", err)
	}
	if !executed {
		t.Error("Alias 'd' was not executed")
	}

	// Test that second alias works
	executed = false
	err = registry.Execute(context.Background(), "dep", struct{}{})
	if err != nil {
		t.Errorf("Failed to execute alias 'dep': %v", err)
	}
	if !executed {
		t.Error("Alias 'dep' was not executed")
	}

	// Test that non-existent command fails
	err = registry.Execute(context.Background(), "nonexistent", struct{}{})
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}

func TestCmdWithAliases(t *testing.T) {
	executed := false
	cmd := CmdWithAliases("list", "List items", []string{"ls", "l"}, func() error {
		executed = true
		return nil
	})

	registry := core.NewRegistry()
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// Test main command
	executed = false
	err = registry.Execute(context.Background(), "list", struct{}{})
	if err != nil {
		t.Errorf("Failed to execute main command: %v", err)
	}
	if !executed {
		t.Error("Main command was not executed")
	}

	// Test aliases
	for _, alias := range []string{"ls", "l"} {
		executed = false
		err = registry.Execute(context.Background(), alias, struct{}{})
		if err != nil {
			t.Errorf("Failed to execute alias '%s': %v", alias, err)
		}
		if !executed {
			t.Errorf("Alias '%s' was not executed", alias)
		}
	}
}

func TestAliasConflict(t *testing.T) {
	registry := core.NewRegistry()

	// Register first command
	cmd1 := CmdWithAliases("deploy", "Deploy the application", []string{"d"}, func() error { return nil })

	err := registry.Register(cmd1)
	if err != nil {
		t.Fatalf("Failed to register first command: %v", err)
	}

	// Try to register second command with conflicting alias
	cmd2 := CmdWithAliases("delete", "Delete something", []string{"d"}, func() error { return nil })

	err = registry.Register(cmd2)
	if err == nil {
		t.Error("Expected error for conflicting alias, got nil")
	}
	if !strings.Contains(err.Error(), "alias d already registered") {
		t.Errorf("Expected conflict error message, got: %v", err)
	}
}

func TestHelpWithAliases(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create CLI with aliases
	app := New("test-app").
		WithCommands(
			CmdWithAliases("deploy", "Deploy the application", []string{"d", "dep"}, func() error {
				fmt.Println("Deploying...")
				return nil
			}),
		).
		Build()

	// Set args and show help
	oldArgs := os.Args
	os.Args = []string{"test-app", "help"}
	defer func() { os.Args = oldArgs }()

	app.RunWithArgs(context.Background())

	// Restore stdout and read output
	_ = w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read help output: %v", err)
	}
	helpText := string(output)

	// Check that aliases are shown in help
	if !strings.Contains(helpText, "deploy, d, dep") {
		t.Errorf("Help text should contain aliases, got: %s", helpText)
	}
}
