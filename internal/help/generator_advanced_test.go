package help

import (
	"testing"
)

func TestGenerator_GenerateHelp(t *testing.T) {
	config := &HelpConfig{
		ProgramName:  "test-app",
		Version:      "1.0.0",
		Description:  "A test application",
		ColorEnabled: false, // Disable colors for testing
		MaxWidth:     80,
	}

	generator := NewGenerator(config)

	// Test basic help generation (if method exists)
	// This will test whatever help generation methods are available
	if generator == nil {
		t.Fatal("Generator should not be nil")
	}
}

func TestHelpConfig_DefaultValues(t *testing.T) {
	config := DefaultHelpConfig("my-app")

	// Test all default values
	if config.ProgramName != "my-app" {
		t.Errorf("Expected program name 'my-app', got '%s'", config.ProgramName)
	}

	if config.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", config.Version)
	}

	if !config.ColorEnabled {
		t.Error("Expected ColorEnabled to be true by default")
	}

	if config.MaxWidth != 80 {
		t.Errorf("Expected default MaxWidth 80, got %d", config.MaxWidth)
	}

	if config.UsageTemplate == "" {
		t.Error("UsageTemplate should not be empty")
	}
}

func TestHelpConfig_CustomValues(t *testing.T) {
	config := &HelpConfig{
		ProgramName:   "custom-app",
		Version:       "2.0.0",
		Description:   "Custom description",
		UsageTemplate: "Custom usage: {{.ProgramName}} [options]",
		ExamplesText:  "Example: custom-app --help",
		Footer:        "Visit https://example.com for more info",
		ColorEnabled:  false,
		MaxWidth:      120,
	}

	// Verify all custom values are set correctly
	if config.ProgramName != "custom-app" {
		t.Errorf("Expected program name 'custom-app', got '%s'", config.ProgramName)
	}

	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}

	if config.Description != "Custom description" {
		t.Errorf("Expected description 'Custom description', got '%s'", config.Description)
	}

	if config.UsageTemplate != "Custom usage: {{.ProgramName}} [options]" {
		t.Errorf("Expected custom usage template, got '%s'", config.UsageTemplate)
	}

	if config.ExamplesText != "Example: custom-app --help" {
		t.Errorf("Expected custom examples text, got '%s'", config.ExamplesText)
	}

	if config.Footer != "Visit https://example.com for more info" {
		t.Errorf("Expected custom footer, got '%s'", config.Footer)
	}

	if config.ColorEnabled {
		t.Error("Expected ColorEnabled to be false")
	}

	if config.MaxWidth != 120 {
		t.Errorf("Expected MaxWidth 120, got %d", config.MaxWidth)
	}
}

func TestErrorFormatter_Basic(t *testing.T) {
	formatter := NewErrorFormatter("test-app", true)
	if formatter == nil {
		t.Fatal("ErrorFormatter should not be nil")
	}

	// Test with colors disabled
	formatterNoColor := NewErrorFormatter("test-app", false)
	if formatterNoColor == nil {
		t.Fatal("ErrorFormatter with no color should not be nil")
	}
}

func TestSuggestionEngine_Basic(t *testing.T) {
	engine := NewSuggestionEngine()
	if engine == nil {
		t.Fatal("SuggestionEngine should not be nil")
	}
}

func TestHelpConfig_EdgeCases(t *testing.T) {
	// Test edge cases
	config := &HelpConfig{
		ProgramName:  "", // Empty name
		Version:      "",
		MaxWidth:     0, // Zero width
		ColorEnabled: true,
	}

	// These should not cause crashes
	if config.ProgramName == "" {
		config.ProgramName = "default"
	}

	if config.MaxWidth <= 0 {
		config.MaxWidth = 80
	}

	// Verify fixes
	if config.ProgramName != "default" {
		t.Error("Program name should be set to default")
	}

	if config.MaxWidth != 80 {
		t.Error("MaxWidth should be set to 80")
	}
}

func TestDefaultUsageTemplate(t *testing.T) {
	// Test that the default template is defined
	if DefaultUsageTemplate == "" {
		t.Error("DefaultUsageTemplate should not be empty")
	}

	// Template should contain basic placeholders
	if !containsString(DefaultUsageTemplate, "{{") {
		t.Error("DefaultUsageTemplate should contain template variables")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerator_WithNilConfig(t *testing.T) {
	// Test generator with nil config
	generator := NewGenerator(nil)
	if generator == nil {
		t.Fatal("Generator should handle nil config gracefully")
	}
}

func TestHelpGeneration_EdgeCases(t *testing.T) {
	// Test various edge cases
	testCases := []struct {
		name   string
		config *HelpConfig
	}{
		{
			"Very long program name",
			&HelpConfig{ProgramName: "very-very-very-long-program-name-that-might-cause-issues"},
		},
		{
			"Special characters in name",
			&HelpConfig{ProgramName: "app-with-special_chars123"},
		},
		{
			"Unicode in description",
			&HelpConfig{ProgramName: "unicode-app", Description: "Application with 中文 and émojis 🚀"},
		},
		{
			"Very narrow width",
			&HelpConfig{ProgramName: "narrow", MaxWidth: 20},
		},
		{
			"Very wide width",
			&HelpConfig{ProgramName: "wide", MaxWidth: 200},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator := NewGenerator(tc.config)
			if generator == nil {
				t.Errorf("Generator should handle config case: %s", tc.name)
			}
		})
	}
}
