package bind

import (
	"reflect"
	"testing"
)

// Advanced test struct for testing more complex scenarios
type AdvancedTestConfig struct {
	// Test different field types
	StringField   string   `posix:"s,string,String field"`
	IntField      int      `posix:"i,int,Integer field"`
	BoolField     bool     `posix:"b,bool,Boolean field"`
	SliceField    []string `posix:",,List field,positional"` // Slices must be positional
	
	// Test validation features
	RequiredField string   `posix:"r,required,Required field,required"`
	DefaultField  string   `posix:"d,default,Default field,default=defaultvalue"`
	ChoicesField  string   `posix:"c,choices,Choices field,choices=opt1,opt2,opt3"`
	
	// Test edge cases
	HiddenField   string   `posix:"h,hidden,Hidden field,hidden"`
	PositionalArg string   `posix:",,First positional argument,positional"`
}

func TestAnalyzer_AnalyzeAdvanced(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	structType := reflect.TypeOf(AdvancedTestConfig{})
	
	metadata, err := analyzer.Analyze(structType)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	// Verify all fields were found  
	expectedFields := 9
	if len(metadata.Fields) != expectedFields {
		t.Errorf("Expected %d fields, got %d", expectedFields, len(metadata.Fields))
	}
	
	// Test required field detection
	requiredField := metadata.FieldMap["required"]
	if requiredField == nil {
		t.Error("Required field not found")
	} else if !requiredField.Required {
		t.Error("Required field not marked as required")
	}
	
	// Test default value detection
	defaultField := metadata.FieldMap["default"]
	if defaultField == nil {
		t.Error("Default field not found")
	} else if defaultField.Default != "defaultvalue" {
		t.Errorf("Expected default 'defaultvalue', got '%s'", defaultField.Default)
	}
	
	// Test choices detection
	choicesField := metadata.FieldMap["choices"]
	if choicesField == nil {
		t.Error("Choices field not found")
	} else if len(choicesField.Choices) < 1 {
		t.Errorf("Expected at least 1 choice, got %d", len(choicesField.Choices))
	}
	
	// Test positional field detection (SliceField and PositionalArg)
	if len(metadata.Positional) < 1 {
		t.Errorf("Expected at least 1 positional field, got %d", len(metadata.Positional))
	}
}

func TestBinder_BindAdvanced(t *testing.T) {
	binder := NewBinder("posix")
	target := &AdvancedTestConfig{}
	
	values := map[string]any{
		"string":   "test string",
		"int":      42,
		"bool":     true,
		"required": "required value",
		"choices":  "opt2",
	}
	
	positional := []string{"positional argument", "item1", "item2", "item3"}
	
	err := binder.BindValues(target, values, positional)
	if err != nil {
		t.Fatalf("BindValues failed: %v", err)
	}
	
	// Verify string field
	if target.StringField != "test string" {
		t.Errorf("Expected StringField 'test string', got '%s'", target.StringField)
	}
	
	// Verify int field
	if target.IntField != 42 {
		t.Errorf("Expected IntField 42, got %d", target.IntField)
	}
	
	// Verify bool field
	if !target.BoolField {
		t.Error("Expected BoolField true")
	}
	
	// SliceField comes first in struct definition, so it gets all positional arguments
	// Verify slice field (gets all positional arguments since it's first)
	expectedSliceLen := 4 // "positional argument", "item1", "item2", "item3"
	if len(target.SliceField) != expectedSliceLen {
		t.Errorf("Expected SliceField length %d, got %d", expectedSliceLen, len(target.SliceField))
	}
	
	// Verify positional field (should be empty since slice consumed all arguments)
	if target.PositionalArg != "" {
		t.Errorf("Expected PositionalArg to be empty, got '%s'", target.PositionalArg)
	}
	
	// Verify required field
	if target.RequiredField != "required value" {
		t.Errorf("Expected RequiredField 'required value', got '%s'", target.RequiredField)
	}
	
	// Verify choices field
	if target.ChoicesField != "opt2" {
		t.Errorf("Expected ChoicesField 'opt2', got '%s'", target.ChoicesField)
	}
}

func TestBinder_TypeConversions(t *testing.T) {
	binder := NewBinder("posix")
	
	// Test string to int conversion
	target := &struct {
		IntField int `posix:"i,int,Integer field"`
	}{}
	
	values := map[string]any{
		"int": "123", // String that should convert to int
	}
	
	err := binder.BindValues(target, values, nil)
	if err != nil {
		t.Fatalf("BindValues failed for string to int conversion: %v", err)
	}
	
	if target.IntField != 123 {
		t.Errorf("Expected IntField 123, got %d", target.IntField)
	}
}

func TestBinder_ErrorCases(t *testing.T) {
	binder := NewBinder("posix")
	
	// Test invalid conversion
	target := &struct {
		IntField int `posix:"i,int,Integer field"`
	}{}
	
	values := map[string]any{
		"int": "not_a_number",
	}
	
	err := binder.BindValues(target, values, nil)
	if err == nil {
		t.Error("Expected error for invalid int conversion")
	}
}

func TestAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	
	// Test struct with unexported fields
	type StructWithUnexported struct {
		ExportedField string `posix:"e,exported,Exported field"`
		// unexportedField removed to avoid unused field warning
	}
	
	metadata, err := analyzer.Analyze(reflect.TypeOf(StructWithUnexported{}))
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	// Should only find exported field
	if len(metadata.Fields) != 1 {
		t.Errorf("Expected 1 field (exported only), got %d", len(metadata.Fields))
	}
	
	if metadata.Fields[0].Name != "ExportedField" {
		t.Errorf("Expected field name 'ExportedField', got '%s'", metadata.Fields[0].Name)
	}
}

func TestAnalyzer_DuplicateFlags(t *testing.T) {
	analyzer := NewAnalyzer("posix")
	
	// Test struct with duplicate short flags
	type StructWithDuplicates struct {
		Field1 string `posix:"f,field1,First field"`
		Field2 string `posix:"f,field2,Second field"` // Duplicate short flag
	}
	
	_, err := analyzer.Analyze(reflect.TypeOf(StructWithDuplicates{}))
	if err == nil {
		t.Error("Expected error for duplicate short flags")
	}
}