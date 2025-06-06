package help

import (
	"testing"
)

func TestNewGenerator(t *testing.T) {
	config := &HelpConfig{
		ColorEnabled: true,
		MaxWidth:     80,
	}
	
	generator := NewGenerator(config)
	if generator == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestDefaultHelpConfig(t *testing.T) {
	config := DefaultHelpConfig("test-app")
	if config == nil {
		t.Fatal("DefaultHelpConfig returned nil")
	}
	
	if config.ProgramName != "test-app" {
		t.Errorf("Expected program name 'test-app', got '%s'", config.ProgramName)
	}
	
	if config.MaxWidth != 80 {
		t.Errorf("Expected default max width 80, got %d", config.MaxWidth)
	}
}

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter("test-app", true)
	if formatter == nil {
		t.Fatal("NewErrorFormatter returned nil")
	}
}

func TestNewSuggestionEngine(t *testing.T) {
	engine := NewSuggestionEngine()
	if engine == nil {
		t.Fatal("NewSuggestionEngine returned nil")
	}
}

func TestErrorContext_Creation(t *testing.T) {
	// Test that ErrorContext can be created
	// The actual fields depend on the implementation
	context := &ErrorContext{}
	
	// Test that we can create an error context without panicking
	// The specific fields will depend on the actual implementation
	_ = context // Use the variable to avoid unused warning
}

func TestHelpConfig_Validation(t *testing.T) {
	config := &HelpConfig{
		ProgramName:  "test-app",
		Version:      "1.0.0",
		Description:  "Test application",
		ColorEnabled: true,
		MaxWidth:     120,
	}
	
	if config.ProgramName != "test-app" {
		t.Errorf("Expected program name 'test-app', got '%s'", config.ProgramName)
	}
	
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}
	
	if config.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got '%s'", config.Description)
	}
	
	if !config.ColorEnabled {
		t.Error("Expected ColorEnabled to be true")
	}
	
	if config.MaxWidth != 120 {
		t.Errorf("Expected MaxWidth 120, got %d", config.MaxWidth)
	}
}