package help

import (
	"errors"
	"strings"
	"testing"
)

func TestNewErrorFormatter(t *testing.T) {
	tests := []struct {
		name        string
		programName string
		useColor    bool
	}{
		{"with color", "test-app", true},
		{"without color", "test-app", false},
		{"empty program name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatter(tt.programName, tt.useColor)
			if formatter == nil {
				t.Fatal("NewErrorFormatter returned nil")
			}

			if formatter.programName != tt.programName {
				t.Errorf("Expected program name '%s', got '%s'", tt.programName, formatter.programName)
			}

			if formatter.useColor != tt.useColor {
				t.Errorf("Expected useColor %v, got %v", tt.useColor, formatter.useColor)
			}
		})
	}
}

func TestErrorFormatter_FormatError(t *testing.T) {
	formatter := NewErrorFormatter("test-app", false) // Disable colors for testing

	tests := []struct {
		name    string
		err     error
		context *ErrorContext
		want    []string // Strings that should be present in output
	}{
		{
			name:    "basic error without context",
			err:     errors.New("something went wrong"),
			context: nil,
			want:    []string{"❌", "something went wrong"},
		},
		{
			name: "unknown command error",
			err:  errors.New("unknown command"),
			context: &ErrorContext{
				Type:        ErrorTypeUnknownCommand,
				Command:     "invalid-cmd",
				Suggestions: []string{"deploy", "build"},
				AllCommands: []string{"deploy", "build", "test"},
			},
			want: []string{"❌ Unknown command:", "invalid-cmd", "💡 Did you mean:", "deploy", "build", "📋 Available commands:", "test-app help"},
		},
		{
			name: "unknown flag error",
			err:  errors.New("unknown flag"),
			context: &ErrorContext{
				Type:        ErrorTypeUnknownFlag,
				Command:     "deploy",
				Flag:        "--invalid-flag",
				Suggestions: []string{"--verbose", "--output"},
				AllFlags:    []string{"--verbose", "--output", "--help"},
			},
			want: []string{"❌ Unknown flag:", "--invalid-flag", "💡 Did you mean:", "--verbose", "🏳️  Available flags:", "test-app help deploy"},
		},
		{
			name: "missing required flag error",
			err:  errors.New("missing required flag"),
			context: &ErrorContext{
				Type:          ErrorTypeMissingRequired,
				Command:       "deploy",
				Flag:          "--name",
				RequiredFlags: []string{"--name", "--target"},
				Examples:      []string{"test-app deploy --name myapp --target prod"},
			},
			want: []string{"❌ Missing required flag:", "--name", "🔴 Required flags:", "✗ (missing)", "📝 Examples:", "test-app help deploy"},
		},
		{
			name: "invalid value error",
			err:  errors.New("invalid value"),
			context: &ErrorContext{
				Type:           ErrorTypeInvalidValue,
				Flag:           "--format",
				Value:          "xml",
				AvailableItems: []string{"json", "yaml", "text"},
				Suggestions:    []string{"yaml"},
			},
			want: []string{"❌ Invalid value for flag", "--format", "xml", "✅ Valid options:", "json", "💡 Did you mean:", "yaml"},
		},
		{
			name: "validation error",
			err:  errors.New("validation failed: email format invalid"),
			context: &ErrorContext{
				Type:     ErrorTypeValidation,
				Examples: []string{"test-app register --email user@example.com"},
			},
			want: []string{"❌ Validation failed:", "validation failed: email format invalid", "📝 Examples:", "user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatError(tt.err, tt.context)

			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("FormatError() output missing '%s'\nFull output:\n%s", want, result)
				}
			}
		})
	}
}

func TestErrorFormatter_Colorize(t *testing.T) {
	tests := []struct {
		name     string
		useColor bool
		color    Color
		text     string
		want     string
	}{
		{
			name:     "with color enabled",
			useColor: true,
			color:    ColorRed,
			text:     "error",
			want:     "\033[31merror\033[0m",
		},
		{
			name:     "with color disabled",
			useColor: false,
			color:    ColorRed,
			text:     "error",
			want:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatter("test-app", tt.useColor)
			result := formatter.colorize(tt.color, tt.text)

			if result != tt.want {
				t.Errorf("colorize() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestNewSuggestionEngine(t *testing.T) {
	engine := NewSuggestionEngine()
	if engine == nil {
		t.Fatal("NewSuggestionEngine returned nil")
	}
}

func TestSuggestionEngine_SuggestCommands(t *testing.T) {
	engine := NewSuggestionEngine()

	tests := []struct {
		name      string
		input     string
		available []string
		want      []string
	}{
		{
			name:      "single character difference",
			input:     "buil",
			available: []string{"build", "test", "deploy"},
			want:      []string{"build"},
		},
		{
			name:      "exact match filtered out",
			input:     "deploy",
			available: []string{"deploy", "develop", "delete"},
			want:      []string{}, // exact match excluded, others too far
		},
		{
			name:      "multiple suggestions sorted by distance",
			input:     "test",
			available: []string{"best", "west", "rest", "testing"},
			want:      []string{"best", "west", "rest"}, // Limit to 3
		},
		{
			name:      "no close matches",
			input:     "xyz",
			available: []string{"deploy", "build", "test"},
			want:      []string{},
		},
		{
			name:      "case insensitive",
			input:     "deploi",
			available: []string{"deploy", "DEVELOP"},
			want:      []string{"deploy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.SuggestCommands(tt.input, tt.available)

			if len(result) != len(tt.want) {
				t.Errorf("SuggestCommands() returned %d suggestions, want %d", len(result), len(tt.want))
			}

			for i, want := range tt.want {
				if i >= len(result) || result[i] != want {
					t.Errorf("SuggestCommands() = %v, want %v", result, tt.want)
					break
				}
			}
		})
	}
}

func TestSuggestionEngine_SuggestFlags(t *testing.T) {
	engine := NewSuggestionEngine()

	tests := []struct {
		name      string
		input     string
		available []string
		want      []string
	}{
		{
			name:      "flag suggestions",
			input:     "--verbos",
			available: []string{"--verbose", "--version", "--help"},
			want:      []string{"--verbose"},
		},
		{
			name:      "short flag suggestions",
			input:     "-v",
			available: []string{"-h", "-v", "-q"},
			want:      []string{"-h", "-q"}, // exact match excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.SuggestFlags(tt.input, tt.available)

			if len(result) != len(tt.want) {
				t.Errorf("SuggestFlags() returned %d suggestions, want %d", len(result), len(tt.want))
			}

			for i, want := range tt.want {
				if i >= len(result) || result[i] != want {
					t.Errorf("SuggestFlags() = %v, want %v", result, tt.want)
					break
				}
			}
		})
	}
}

func TestSuggestionEngine_SuggestValues(t *testing.T) {
	engine := NewSuggestionEngine()

	tests := []struct {
		name      string
		input     string
		available []string
		want      []string
	}{
		{
			name:      "value suggestions with stricter matching",
			input:     "jso",
			available: []string{"json", "yaml", "text"},
			want:      []string{"json"},
		},
		{
			name:      "no suggestions for distant matches",
			input:     "xml",
			available: []string{"json", "yaml", "text"},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.SuggestValues(tt.input, tt.available)

			if len(result) != len(tt.want) {
				t.Errorf("SuggestValues() returned %d suggestions, want %d", len(result), len(tt.want))
			}

			for i, want := range tt.want {
				if i >= len(result) || result[i] != want {
					t.Errorf("SuggestValues() = %v, want %v", result, tt.want)
					break
				}
			}
		})
	}
}

func TestSuggestionEngine_LevenshteinDistance(t *testing.T) {
	engine := NewSuggestionEngine()

	tests := []struct {
		name string
		s1   string
		s2   string
		want int
	}{
		{"identical strings", "hello", "hello", 0},
		{"empty strings", "", "", 0},
		{"one empty", "", "hello", 5},
		{"single character difference", "hello", "hallo", 1},
		{"insertion", "hello", "helloo", 1},
		{"deletion", "hello", "hell", 1},
		{"substitution", "hello", "hxllo", 1},
		{"multiple operations", "kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.levenshteinDistance(tt.s1, tt.s2)
			if result != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.want)
			}
		})
	}
}

func TestNewErrorContext(t *testing.T) {
	builder := NewErrorContext()
	if builder == nil {
		t.Fatal("NewErrorContext returned nil")
	}

	if builder.context == nil {
		t.Fatal("ErrorContextBuilder context is nil")
	}
}

func TestErrorContextBuilder(t *testing.T) {
	builder := NewErrorContext()

	context := builder.
		Type(ErrorTypeUnknownCommand).
		Command("invalid-cmd").
		Flag("--invalid").
		Value("invalid-value").
		Suggestions([]string{"suggestion1", "suggestion2"}).
		AvailableItems([]string{"item1", "item2"}).
		AllCommands([]string{"cmd1", "cmd2"}).
		AllFlags([]string{"--flag1", "--flag2"}).
		RequiredFlags([]string{"--required"}).
		Examples([]string{"example command"}).
		Build()

	if context.Type != ErrorTypeUnknownCommand {
		t.Errorf("Expected Type %v, got %v", ErrorTypeUnknownCommand, context.Type)
	}

	if context.Command != "invalid-cmd" {
		t.Errorf("Expected Command 'invalid-cmd', got '%s'", context.Command)
	}

	if context.Flag != "--invalid" {
		t.Errorf("Expected Flag '--invalid', got '%s'", context.Flag)
	}

	if context.Value != "invalid-value" {
		t.Errorf("Expected Value 'invalid-value', got '%s'", context.Value)
	}

	if len(context.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(context.Suggestions))
	}

	if len(context.AvailableItems) != 2 {
		t.Errorf("Expected 2 available items, got %d", len(context.AvailableItems))
	}

	if len(context.AllCommands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(context.AllCommands))
	}

	if len(context.AllFlags) != 2 {
		t.Errorf("Expected 2 flags, got %d", len(context.AllFlags))
	}

	if len(context.RequiredFlags) != 1 {
		t.Errorf("Expected 1 required flag, got %d", len(context.RequiredFlags))
	}

	if len(context.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(context.Examples))
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that error types are distinct
	types := []ErrorType{
		ErrorTypeGeneric,
		ErrorTypeUnknownCommand,
		ErrorTypeUnknownFlag,
		ErrorTypeMissingRequired,
		ErrorTypeInvalidValue,
		ErrorTypeValidation,
	}

	for i, typeA := range types {
		for j, typeB := range types {
			if i != j && typeA == typeB {
				t.Errorf("Error types at index %d and %d are identical: %v", i, j, typeA)
			}
		}
	}
}

func TestColorConstants(t *testing.T) {
	colors := map[string]Color{
		"ColorReset":  ColorReset,
		"ColorRed":    ColorRed,
		"ColorGreen":  ColorGreen,
		"ColorYellow": ColorYellow,
		"ColorBlue":   ColorBlue,
		"ColorPurple": ColorPurple,
		"ColorCyan":   ColorCyan,
		"ColorWhite":  ColorWhite,
		"ColorBold":   ColorBold,
	}

	for name, color := range colors {
		if string(color) == "" {
			t.Errorf("Color %s is empty", name)
		}
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c int
		want    int
	}{
		{"a is minimum", 1, 2, 3, 1},
		{"b is minimum", 3, 1, 2, 1},
		{"c is minimum", 3, 2, 1, 1},
		{"all equal", 5, 5, 5, 5},
		{"negative numbers", -1, -2, -3, -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b, tt.c)
			if result != tt.want {
				t.Errorf("min(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, result, tt.want)
			}
		})
	}
}
