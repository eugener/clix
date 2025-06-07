package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eugener/clix/internal/output"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter(output.FormatJSON)
	if formatter == nil {
		t.Fatal("NewFormatter returned nil")
	}
}

func TestNewFormatterWithWriter(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatterWithWriter(&buf, output.FormatJSON)
	if formatter == nil {
		t.Fatal("NewFormatterWithWriter returned nil")
	}
}

func TestFormatAndOutput(t *testing.T) {
	// Create a simple data structure
	data := map[string]string{"name": "test", "value": "123"}

	// Test different formats
	formats := []output.Format{
		"json",
		"yaml", 
		"table",
		"text",
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			// Since FormatAndOutput writes to stdout, we'll test FormatToWriter instead
			var buf bytes.Buffer
			err := FormatToWriter(&buf, data, format)
			if err != nil {
				t.Errorf("FormatToWriter failed: %v", err)
			}
			
			output := buf.String()
			if output == "" {
				t.Error("Expected non-empty output")
			}

			// Basic validation for each format
			switch format {
			case "json":
				if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
					t.Error("JSON output doesn't look like JSON")
				}
			case "yaml":
				if !strings.Contains(output, "name:") {
					t.Error("YAML output doesn't look like YAML")
				}
			case "table":
				if !strings.Contains(output, "│") {
					t.Error("Table output doesn't contain table separators")
				}
			case "text":
				// Text format should contain the data in some form
				if !strings.Contains(output, "test") {
					t.Error("Text output doesn't contain expected data")
				}
			}
		})
	}
}

func TestFormatToWriter(t *testing.T) {
	var buf bytes.Buffer
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "Alice",
		Age:  30,
	}

	err := FormatToWriter(&buf, data, "json")
	if err != nil {
		t.Fatalf("FormatToWriter failed: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "30") {
		t.Errorf("Output missing expected data: %s", result)
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	expectedFormats := []output.Format{
		"json",
		"yaml", 
		"table",
		"text",
	}

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

func TestOutputConfig(t *testing.T) {
	config := OutputConfig{
		Format: "json",
	}

	if config.Format != "json" {
		t.Errorf("Expected format JSON, got %s", config.Format)
	}
}

// Test complex data structures
func TestComplexDataFormatting(t *testing.T) {
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	data := []Product{
		{ID: 1, Name: "Laptop", Price: 999.99},
		{ID: 2, Name: "Mouse", Price: 29.99},
	}

	formats := []output.Format{
		"json",
		"yaml",
		"table",
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := FormatToWriter(&buf, data, format)
			if err != nil {
				t.Errorf("FormatToWriter failed for %s: %v", format, err)
			}

			result := buf.String()
			if !strings.Contains(result, "Laptop") {
				t.Errorf("Output missing expected data for format %s: %s", format, result)
			}
		})
	}
}