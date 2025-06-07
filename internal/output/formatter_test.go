package output

import (
	"bytes"
	"strings"
	"testing"
)

// Test data structures
type Person struct {
	Name string
	Age  int
	City string
}

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price float64 `json:"price"`
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	expectedFormats := []Format{FormatJSON, FormatYAML, FormatTable, FormatText}
	
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}
	
	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found", expected)
		}
	}
}

func TestValidFormat(t *testing.T) {
	tests := []struct {
		format string
		valid  bool
	}{
		{"json", true},
		{"yaml", true},
		{"table", true},
		{"text", true},
		{"xml", false},
		{"csv", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := ValidFormat(tt.format)
			if result != tt.valid {
				t.Errorf("ValidFormat(%q) = %v, want %v", tt.format, result, tt.valid)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(&buf, FormatJSON)
	
	if formatter == nil {
		t.Fatal("NewFormatter returned nil")
	}
	
	if formatter.writer != &buf {
		t.Error("Formatter writer not set correctly")
	}
	
	if formatter.format != FormatJSON {
		t.Error("Formatter format not set correctly")
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	tests := []struct {
		name string
		data any
		want []string // Strings that should be present in output
	}{
		{
			name: "simple struct",
			data: Person{Name: "Alice", Age: 30, City: "New York"},
			want: []string{"Alice", "30", "New York", "Name", "Age", "City"},
		},
		{
			name: "slice of structs",
			data: []Person{
				{Name: "Alice", Age: 30, City: "New York"},
				{Name: "Bob", Age: 25, City: "San Francisco"},
			},
			want: []string{"Alice", "Bob", "30", "25"},
		},
		{
			name: "map",
			data: map[string]any{"name": "test", "count": 42},
			want: []string{"name", "test", "count", "42"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, FormatJSON)
			
			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			
			output := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("JSON output missing %q\nFull output:\n%s", want, output)
				}
			}
		})
	}
}

func TestFormatter_FormatYAML(t *testing.T) {
	tests := []struct {
		name string
		data any
		want []string
	}{
		{
			name: "simple struct",
			data: Person{Name: "Alice", Age: 30, City: "New York"},
			want: []string{"name: Alice", "age: 30", "city: New York"},
		},
		{
			name: "slice",
			data: []string{"apple", "banana", "cherry"},
			want: []string{"- apple", "- banana", "- cherry"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, FormatYAML)
			
			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			
			output := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("YAML output missing %q\nFull output:\n%s", want, output)
				}
			}
		})
	}
}

func TestFormatter_FormatTable(t *testing.T) {
	tests := []struct {
		name string
		data any
		want []string
	}{
		{
			name: "slice of structs",
			data: []Person{
				{Name: "Alice", Age: 30, City: "New York"},
				{Name: "Bob", Age: 25, City: "SF"},
			},
			want: []string{"Name", "Age", "City", "Alice", "Bob", "30", "25", "New York", "SF", "│", "┌", "─"},
		},
		{
			name: "single struct",
			data: Person{Name: "Alice", Age: 30, City: "New York"},
			want: []string{"Field", "Value", "Name", "Alice", "Age", "30", "City", "New York", "│", "┌", "─"},
		},
		{
			name: "map",
			data: map[string]string{"name": "test", "status": "active"},
			want: []string{"Key", "Value", "name", "test", "status", "active", "│", "┌", "─"},
		},
		{
			name: "slice of simple values",
			data: []string{"apple", "banana", "cherry"},
			want: []string{"Value", "apple", "banana", "cherry", "│", "┌", "─"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, FormatTable)
			
			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			
			output := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("Table output missing %q\nFull output:\n%s", want, output)
				}
			}
		})
	}
}

func TestFormatter_FormatText(t *testing.T) {
	tests := []struct {
		name string
		data any
		want string
	}{
		{
			name: "string",
			data: "Hello, World!",
			want: "Hello, World!",
		},
		{
			name: "number",
			data: 42,
			want: "42",
		},
		{
			name: "struct",
			data: Person{Name: "Alice", Age: 30},
			want: "Alice",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, FormatText)
			
			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			
			output := buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("Text output missing %q\nFull output:\n%s", tt.want, output)
			}
		})
	}
}

func TestFormatter_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(&buf, Format("xml"))
	
	err := formatter.Format("test data")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}
}

func TestFormatter_EmptyData(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		data   any
	}{
		{"empty slice JSON", FormatJSON, []Person{}},
		{"empty slice YAML", FormatYAML, []Person{}},
		{"empty slice Table", FormatTable, []Person{}},
		{"nil data JSON", FormatJSON, nil},
		{"nil data YAML", FormatYAML, nil},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, tt.format)
			
			err := formatter.Format(tt.data)
			if err != nil {
				t.Errorf("Format() error = %v", err)
			}
			
			// Should not panic and should produce some output
			output := buf.String()
			if output == "" && tt.format != FormatTable {
				t.Error("Expected some output for empty data")
			}
		})
	}
}

func TestFormatter_ComplexStructures(t *testing.T) {
	type NestedStruct struct {
		Products []Product `json:"products"`
		Total    int       `json:"total"`
	}
	
	data := NestedStruct{
		Products: []Product{
			{ID: 1, Name: "Laptop", Price: 999.99},
			{ID: 2, Name: "Mouse", Price: 29.99},
		},
		Total: 2,
	}
	
	formats := []Format{FormatJSON, FormatYAML, FormatTable}
	
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(&buf, format)
			
			err := formatter.Format(data)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			
			output := buf.String()
			if output == "" {
				t.Error("Expected non-empty output")
			}
			
			// Check that essential data is present
			if !strings.Contains(output, "Laptop") {
				t.Error("Output missing 'Laptop'")
			}
		})
	}
}

// Benchmark tests
func BenchmarkFormatter_JSON(b *testing.B) {
	data := []Person{
		{Name: "Alice", Age: 30, City: "New York"},
		{Name: "Bob", Age: 25, City: "San Francisco"},
		{Name: "Charlie", Age: 35, City: "Chicago"},
	}
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		formatter := NewFormatter(&buf, FormatJSON)
		formatter.Format(data)
	}
}

func BenchmarkFormatter_Table(b *testing.B) {
	data := []Person{
		{Name: "Alice", Age: 30, City: "New York"},
		{Name: "Bob", Age: 25, City: "San Francisco"},
		{Name: "Charlie", Age: 35, City: "Chicago"},
	}
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		formatter := NewFormatter(&buf, FormatTable)
		formatter.Format(data)
	}
}