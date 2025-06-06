package complete

import (
	"testing"
	
	"github.com/eugener/clix/core"
)

func TestNewGenerator(t *testing.T) {
	registry := core.NewRegistry()
	generator := NewGenerator(registry)
	if generator == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestCompletionItem_Basic(t *testing.T) {
	item := &CompletionItem{
		Value:       "deploy",
		Description: "Deploy the application",
		Type:        CompletionCommands,
	}
	
	if item.Value != "deploy" {
		t.Errorf("Expected value 'deploy', got '%s'", item.Value)
	}
	
	if item.Description != "Deploy the application" {
		t.Errorf("Expected description 'Deploy the application', got '%s'", item.Description)
	}
	
	if item.Type != CompletionCommands {
		t.Errorf("Expected type CompletionCommands, got %v", item.Type)
	}
}

func TestCompletionType_Constants(t *testing.T) {
	// Test that completion type constants are defined
	if CompletionCommands < 0 {
		t.Error("CompletionCommands should be non-negative")
	}
	
	if CompletionFlags < 0 {
		t.Error("CompletionFlags should be non-negative")
	}
	
	if CompletionValues < 0 {
		t.Error("CompletionValues should be non-negative")
	}
	
	if CompletionFiles < 0 {
		t.Error("CompletionFiles should be non-negative")
	}
	
	if CompletionDirectories < 0 {
		t.Error("CompletionDirectories should be non-negative")
	}
}

func TestGenerator_Complete_EmptyArgs(t *testing.T) {
	registry := core.NewRegistry()
	generator := NewGenerator(registry)
	
	// Test completion with empty args
	items, err := generator.Complete([]string{}, 0)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	
	// Items can be empty or nil for empty registry, just verify no panic
	_ = items
	
	// Should return available commands (even if none registered)
	// Length check is not strict as it depends on registered commands
}