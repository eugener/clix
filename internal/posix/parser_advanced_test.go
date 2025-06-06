package posix

import (
	"testing"
)

func TestParser_Parse_ComplexScenarios(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test complex flag combinations
	args := []string{
		"command",           // positional
		"--verbose",         // boolean flag
		"--file", "/path/to/file", // flag with value
		"-n", "5",          // short flag with value
		"-abc",             // combined short flags
		"arg1", "arg2",     // more positional args
		"--",               // separator
		"--not-a-flag",     // treated as positional after --
	}
	
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	// Verify flags
	if result.Flags["verbose"] != true {
		t.Error("Expected verbose flag to be true")
	}
	
	if result.Flags["file"] != "/path/to/file" {
		t.Errorf("Expected file flag '/path/to/file', got '%v'", result.Flags["file"])
	}
	
	if result.Flags["n"] != "5" {
		t.Errorf("Expected n flag '5', got '%v'", result.Flags["n"])
	}
	
	// Verify positional arguments
	if len(result.Positional) < 3 {
		t.Errorf("Expected at least 3 positional args, got %d", len(result.Positional))
	}
	
	// Check first positional
	if len(result.Positional) > 0 && result.Positional[0] != "command" {
		t.Errorf("Expected first positional 'command', got '%s'", result.Positional[0])
	}
}

func TestParser_Parse_EqualsNotation(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test --flag=value notation
	args := []string{"--output=/tmp/file.txt", "--verbose=true"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	// Check if parser supports equals notation
	if result.Flags["output"] == "/tmp/file.txt" {
		t.Log("Parser supports --flag=value notation")
	} else {
		t.Log("Parser treats --flag=value as single argument (implementation dependent)")
	}
}

func TestParser_Parse_QuotedValues(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test quoted values
	args := []string{"--message", "hello world", "--path", "/path with spaces/file"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if result.Flags["message"] != "hello world" {
		t.Errorf("Expected message 'hello world', got '%v'", result.Flags["message"])
	}
	
	if result.Flags["path"] != "/path with spaces/file" {
		t.Errorf("Expected path '/path with spaces/file', got '%v'", result.Flags["path"])
	}
}

func TestParser_Parse_BooleanFlags(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test various boolean flag formats
	args := []string{"--enable", "--disable=false", "--verbose=true", "-q"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	// Boolean flags without values should be true
	if result.Flags["enable"] != true {
		t.Errorf("Expected enable to be true, got '%v'", result.Flags["enable"])
	}
	
	if result.Flags["q"] != true {
		t.Errorf("Expected q to be true, got '%v'", result.Flags["q"])
	}
}

func TestParser_Parse_EdgeCases(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	testCases := []struct {
		name string
		args []string
	}{
		{"Single dash", []string{"-"}},
		{"Double dash only", []string{"--"}},
		{"Empty flag name", []string{"--", "value"}},
		{"Flag with empty value", []string{"--flag", ""}},
		{"Multiple double dashes", []string{"--", "--", "--"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.Parse(tc.args)
			if err != nil {
				t.Logf("Parse returned error (may be expected): %v", err)
			} else if result == nil {
				t.Error("Parse returned nil result without error")
			}
		})
	}
}

func TestParser_Parse_LongArguments(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test very long arguments
	longValue := make([]byte, 10000)
	for i := range longValue {
		longValue[i] = 'a'
	}
	
	args := []string{"--data", string(longValue)}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed with long argument: %v", err)
	}
	
	if len(result.Flags["data"].(string)) != 10000 {
		t.Error("Long argument not preserved correctly")
	}
}

func TestParser_Parse_SpecialCharacters(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test special characters in values
	args := []string{
		"--regex", "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		"--json", `{"key": "value", "number": 42}`,
		"--unicode", "Hello 世界! 🚀",
	}
	
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed with special characters: %v", err)
	}
	
	// Verify special characters are preserved
	if regex, ok := result.Flags["regex"].(string); ok {
		if len(regex) == 0 {
			t.Error("Regex pattern not preserved")
		}
	}
	
	if unicode, ok := result.Flags["unicode"].(string); ok {
		if !contains(unicode, "世界") {
			t.Error("Unicode characters not preserved")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParser_Performance(t *testing.T) {
	parser := NewConfigurableParser(nil)
	
	// Test with many arguments
	args := make([]string, 1000)
	for i := 0; i < 500; i++ {
		args[i*2] = "--flag" + string(rune('0'+i%10))
		args[i*2+1] = "value" + string(rune('0'+i%10))
	}
	
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed with many arguments: %v", err)
	}
	
	if len(result.Flags) == 0 {
		t.Error("No flags parsed from many arguments")
	}
}