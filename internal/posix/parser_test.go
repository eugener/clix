package posix

import (
	"testing"
)

func TestNewConfigurableParser(t *testing.T) {
	parser := NewConfigurableParser(nil)
	if parser == nil {
		t.Fatal("NewConfigurableParser returned nil")
	}
}

func TestParser_Parse_SimpleFlags(t *testing.T) {
	parser := NewConfigurableParser(nil)

	args := []string{"--name", "testvalue", "--count", "5", "--verbose"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Parse result is nil")
	}

	// Check long flags
	if result.Flags["name"] != "testvalue" {
		t.Errorf("Expected name 'testvalue', got '%v'", result.Flags["name"])
	}

	if result.Flags["count"] != "5" {
		t.Errorf("Expected count '5', got '%v'", result.Flags["count"])
	}

	if result.Flags["verbose"] != true {
		t.Errorf("Expected verbose true, got '%v'", result.Flags["verbose"])
	}
}

func TestParser_Parse_ShortFlags(t *testing.T) {
	parser := NewConfigurableParser(nil)

	args := []string{"-n", "testvalue", "-c", "5", "-v"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check short flags (assuming they map to the same keys)
	if result.Flags["n"] != "testvalue" {
		t.Errorf("Expected -n 'testvalue', got '%v'", result.Flags["n"])
	}

	if result.Flags["c"] != "5" {
		t.Errorf("Expected -c '5', got '%v'", result.Flags["c"])
	}

	if result.Flags["v"] != true {
		t.Errorf("Expected -v true, got '%v'", result.Flags["v"])
	}
}

func TestParser_Parse_MixedFlags(t *testing.T) {
	parser := NewConfigurableParser(nil)

	args := []string{"--name", "testvalue", "-c", "5"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Flags["name"] != "testvalue" {
		t.Errorf("Expected name 'testvalue', got '%v'", result.Flags["name"])
	}

	if result.Flags["c"] != "5" {
		t.Errorf("Expected -c '5', got '%v'", result.Flags["c"])
	}
}

func TestParser_Parse_PositionalArgs(t *testing.T) {
	parser := NewConfigurableParser(nil)

	args := []string{"command", "arg1", "arg2", "--flag", "value"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Positional) < 2 {
		t.Errorf("Expected at least 2 positional args, got %d", len(result.Positional))
	}

	// First positional should be "command"
	if len(result.Positional) > 0 && result.Positional[0] != "command" {
		t.Errorf("Expected first positional 'command', got '%s'", result.Positional[0])
	}
}

func TestParser_Parse_EmptyArgs(t *testing.T) {
	parser := NewConfigurableParser(nil)

	args := []string{}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed with empty args: %v", err)
	}

	if result == nil {
		t.Fatal("Parse result is nil")
	}

	if len(result.Flags) != 0 {
		t.Errorf("Expected 0 flags with empty args, got %d", len(result.Flags))
	}

	if len(result.Positional) != 0 {
		t.Errorf("Expected 0 positional args with empty args, got %d", len(result.Positional))
	}
}

func TestParser_Parse_InvalidFlag(t *testing.T) {
	parser := NewConfigurableParser(nil)

	// Test malformed flag
	args := []string{"--"}
	result, err := parser.Parse(args)

	// Should not fail, just treat as positional or ignore
	if err != nil {
		t.Logf("Parse returned error (acceptable): %v", err)
	}

	if result == nil {
		t.Fatal("Parse result is nil")
	}
}

func TestParser_Parse_FlagWithoutValue(t *testing.T) {
	parser := NewConfigurableParser(nil)

	// Flag that expects value but doesn't get one
	args := []string{"--name"}
	result, err := parser.Parse(args)

	// This might fail or succeed depending on implementation
	// If it succeeds, name should be empty or true
	if err != nil {
		t.Logf("Parse returned expected error: %v", err)
		return
	}

	if result != nil && result.Flags["name"] == nil {
		t.Log("Flag without value treated as nil")
	}
}

func TestParser_Parse_DoubleDash(t *testing.T) {
	parser := NewConfigurableParser(nil)

	// Everything after -- should be positional
	args := []string{"--flag", "value", "--", "--not-a-flag", "arg"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have the flag before --
	if result.Flags["flag"] != "value" {
		t.Errorf("Expected flag 'value', got '%v'", result.Flags["flag"])
	}

	// Should have positional args after --
	found := false
	for _, pos := range result.Positional {
		if pos == "--not-a-flag" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected '--not-a-flag' in positional args")
	}
}
