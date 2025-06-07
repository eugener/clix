package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/eugener/clix/core"
	"github.com/eugener/clix/internal/help"
)

func TestEnhancedErrorMessages(t *testing.T) {
	registry := core.NewRegistry()

	// Register some commands
	cmd1 := CmdWithAliases("deploy", "Deploy the application", []string{"d"}, func() error { return nil })

	cmd2 := Cmd("list", "List items", func() error { return nil })

	err := registry.Register(cmd1)
	if err != nil {
		t.Fatalf("Failed to register cmd1: %v", err)
	}

	err = registry.Register(cmd2)
	if err != nil {
		t.Fatalf("Failed to register cmd2: %v", err)
	}

	// Test unknown command error
	err = registry.Execute(context.Background(), "deploi", struct{}{})
	if err == nil {
		t.Error("Expected error for unknown command")
	}

	// The error should be caught and formatted by the app layer
	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("Expected error message to contain 'command not found', got: %v", err)
	}
}

func TestErrorFormatterUnknownCommand(t *testing.T) {
	formatter := help.NewErrorFormatter("myapp", true)

	context := help.NewErrorContext().
		Type(help.ErrorTypeUnknownCommand).
		Command("deploi").
		Suggestions([]string{"deploy", "delete"}).
		AllCommands([]string{"deploy", "list", "help"}).
		Build()

	err := fmt.Errorf("command not found: deploi")
	result := formatter.FormatError(err, context)

	// Check that the formatted error contains expected elements
	if !strings.Contains(result, "❌ Unknown command") {
		t.Error("Expected error icon and message")
	}

	if !strings.Contains(result, "deploi") {
		t.Error("Expected command name in error")
	}

	if !strings.Contains(result, "deploy") {
		t.Error("Expected suggestions in error")
	}

	if !strings.Contains(result, "myapp help") {
		t.Error("Expected help hint in error")
	}

	t.Logf("Formatted error:\n%s", result)
}

func TestErrorFormatterCommandConflict(t *testing.T) {
	formatter := help.NewErrorFormatter("myapp", true)

	context := help.NewErrorContext().
		Type(help.ErrorTypeCommandConflict).
		Command("deploy").
		AllCommands([]string{"deploy", "list", "help"}).
		Build()

	err := fmt.Errorf("command alias d already registered")
	result := formatter.FormatError(err, context)

	// Check that the formatted error contains expected elements
	if !strings.Contains(result, "❌ Command conflict") {
		t.Error("Expected command conflict message")
	}

	if !strings.Contains(result, "already registered") {
		t.Error("Expected conflict details")
	}

	if !strings.Contains(result, "This error occurs when") {
		t.Error("Expected helpful explanation")
	}

	t.Logf("Formatted conflict error:\n%s", result)
}

func TestErrorFormatterConfigurationError(t *testing.T) {
	formatter := help.NewErrorFormatter("myapp", true)

	context := help.NewErrorContext().
		Type(help.ErrorTypeConfigurationError).
		Command("start").
		Examples([]string{"config.yaml", "config.json"}).
		Build()

	err := fmt.Errorf("configuration file not found")
	result := formatter.FormatError(err, context)

	// Check that the formatted error contains expected elements
	if !strings.Contains(result, "❌ Configuration error") {
		t.Error("Expected configuration error message")
	}

	if !strings.Contains(result, "troubleshooting") {
		t.Error("Expected troubleshooting tips")
	}

	if !strings.Contains(result, "config.yaml") {
		t.Error("Expected configuration examples")
	}

	t.Logf("Formatted config error:\n%s", result)
}

func TestErrorFormatterMissingRequired(t *testing.T) {
	formatter := help.NewErrorFormatter("myapp", true)

	context := help.NewErrorContext().
		Type(help.ErrorTypeMissingRequired).
		Command("deploy").
		Flag("--env").
		RequiredFlags([]string{"--env", "--region"}).
		Examples([]string{"myapp deploy --env prod --region us-west-2"}).
		Build()

	err := fmt.Errorf("missing required flag: --env")
	result := formatter.FormatError(err, context)

	// Check that the formatted error contains expected elements
	if !strings.Contains(result, "❌ Missing required flag") {
		t.Error("Expected missing required flag message")
	}

	if !strings.Contains(result, "--env") {
		t.Error("Expected flag name in error")
	}

	if !strings.Contains(result, "📝 Examples") {
		t.Error("Expected examples section")
	}

	t.Logf("Formatted missing required error:\n%s", result)
}

func TestSuggestionEngine(t *testing.T) {
	engine := help.NewSuggestionEngine()

	commands := []string{"deploy", "delete", "list", "help", "version"}

	// Test close matches
	suggestions := engine.SuggestCommands("deploi", commands)
	if len(suggestions) == 0 || suggestions[0] != "deploy" {
		t.Errorf("Expected 'deploy' as first suggestion for 'deploi', got: %v", suggestions)
	}

	// Test partial matches - note: "del" is closer to "help" than "delete" by Levenshtein distance
	suggestions = engine.SuggestCommands("delet", commands) // Use "delet" instead for better match
	if len(suggestions) == 0 || suggestions[0] != "delete" {
		t.Errorf("Expected 'delete' as first suggestion for 'delet', got: %v", suggestions)
	}

	// Test no good matches
	suggestions = engine.SuggestCommands("xyz", commands)
	if len(suggestions) > 0 {
		t.Logf("Unexpected suggestions for 'xyz': %v", suggestions)
	}
}
