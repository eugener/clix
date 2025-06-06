package bind

import (
	"reflect"
	"testing"
)

// Test struct for bind tests
type BindTestConfig struct {
	Name     string `posix:"n,name,Name parameter,required"`
	Count    int    `posix:"c,count,Count parameter,default=1"`
	Verbose  bool   `posix:"v,verbose,Enable verbose output"`
	Optional string `posix:"o,optional,Optional parameter"`
}

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	if analyzer == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
	
	if analyzer.tagName != "posix" {
		t.Errorf("Expected tag name 'posix', got '%s'", analyzer.tagName)
	}
}

func TestNewAnalyzer_DefaultTag(t *testing.T) {
	analyzer := NewAnalyzer("")
	if analyzer.tagName != "posix" {
		t.Errorf("Expected default tag name 'posix', got '%s'", analyzer.tagName)
	}
}

func TestAnalyzer_Analyze(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	structType := reflect.TypeOf(BindTestConfig{})
	
	metadata, err := analyzer.Analyze(structType)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if metadata == nil {
		t.Fatal("Metadata is nil")
	}
	
	if len(metadata.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(metadata.Fields))
	}
	
	// Check name field
	nameField := metadata.FieldMap["name"]
	if nameField == nil {
		t.Error("Name field not found in FieldMap")
	} else {
		if nameField.Short != "n" {
			t.Errorf("Expected short flag 'n', got '%s'", nameField.Short)
		}
		if nameField.Long != "name" {
			t.Errorf("Expected long flag 'name', got '%s'", nameField.Long)
		}
		if !nameField.Required {
			t.Error("Name field should be required")
		}
	}
	
	// Check count field with default
	countField := metadata.FieldMap["count"]
	if countField == nil {
		t.Error("Count field not found in FieldMap")
	} else {
		if countField.Default != "1" {
			t.Errorf("Expected default '1', got '%s'", countField.Default)
		}
	}
}

func TestAnalyzer_Analyze_PointerType(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	structType := reflect.TypeOf(&BindTestConfig{})
	
	metadata, err := analyzer.Analyze(structType)
	if err != nil {
		t.Fatalf("Analyze failed with pointer type: %v", err)
	}
	
	if len(metadata.Fields) != 4 {
		t.Errorf("Expected 4 fields with pointer type, got %d", len(metadata.Fields))
	}
}

func TestAnalyzer_Analyze_InvalidType(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	
	// Test with non-struct type
	_, err := analyzer.Analyze(reflect.TypeOf("string"))
	if err == nil {
		t.Error("Expected error for non-struct type")
	}
}

func TestAnalyzer_Analyze_NoTags(t *testing.T) {
	type NoTagsStruct struct {
		Field1 string
		Field2 int
	}
	
	analyzer := NewAnalyzer("posix")
	metadata, err := analyzer.Analyze(reflect.TypeOf(NoTagsStruct{}))
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	// Should have no fields since they don't have tags
	if len(metadata.Fields) != 0 {
		t.Errorf("Expected 0 fields for struct without tags, got %d", len(metadata.Fields))
	}
}

func TestNewBinder(t *testing.T) {
	binder := NewBinder("posix")
	if binder == nil {
		t.Fatal("NewBinder returned nil")
	}
}

func TestBinder_BindValues(t *testing.T) {
	binder := NewBinder("posix")
	target := &BindTestConfig{}
	
	values := map[string]any{
		"name":    "testname",
		"count":   5,
		"verbose": true,
	}
	
	err := binder.BindValues(target, values, nil)
	if err != nil {
		t.Fatalf("BindValues failed: %v", err)
	}
	
	if target.Name != "testname" {
		t.Errorf("Expected Name 'testname', got '%s'", target.Name)
	}
	
	if target.Count != 5 {
		t.Errorf("Expected Count 5, got %d", target.Count)
	}
	
	if !target.Verbose {
		t.Error("Expected Verbose true")
	}
}

func TestBinder_BindValues_InvalidTarget(t *testing.T) {
	binder := NewBinder("posix")
	target := "not a struct pointer"
	
	values := map[string]any{
		"name": "testname",
	}
	
	err := binder.BindValues(target, values, nil)
	if err == nil {
		t.Error("Expected error for invalid target type")
	}
}

func TestBinder_BindValues_Positional(t *testing.T) {
	type PositionalConfig struct {
		Arg1 string `posix:",arg1,First argument,positional"`
		Arg2 string `posix:",arg2,Second argument,positional"`
	}
	
	binder := NewBinder("posix")
	target := &PositionalConfig{}
	
	positional := []string{"value1", "value2"}
	
	err := binder.BindValues(target, nil, positional)
	if err != nil {
		t.Fatalf("BindValues failed with positional args: %v", err)
	}
	
	if target.Arg1 != "value1" {
		t.Errorf("Expected Arg1 'value1', got '%s'", target.Arg1)
	}
	
	if target.Arg2 != "value2" {
		t.Errorf("Expected Arg2 'value2', got '%s'", target.Arg2)
	}
}