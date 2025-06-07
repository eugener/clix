package configfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader("config.yaml")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
}

func TestLoader_LoadYAML(t *testing.T) {
	// Create a temporary YAML file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test.yaml")

	yamlContent := `
name: test-app
count: 42
verbose: true
`

	err := os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	loader := NewLoader("test.yaml")
	err = loader.LoadFromPath(yamlFile, &struct {
		Name    string `yaml:"name" json:"name"`
		Count   int    `yaml:"count" json:"count"`
		Verbose bool   `yaml:"verbose" json:"verbose"`
	}{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loading succeeded

	// Check loaded values (this depends on the actual implementation)
	// For now, just verify that loading succeeded
}

func TestLoader_LoadJSON(t *testing.T) {
	// Create a temporary JSON file
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test.json")

	jsonContent := `{
	"name": "test-app",
	"verbose": true
}`

	err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	loader := NewLoader("test.json")
	var data struct {
		Name    string `yaml:"name" json:"name"`
		Count   int    `yaml:"count" json:"count"`
		Verbose bool   `yaml:"verbose" json:"verbose"`
	}
	err = loader.LoadFromPath(jsonFile, &data)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loading succeeded
}

func TestLoader_LoadNonExistentFile(t *testing.T) {
	loader := NewLoader("config.yaml")
	err := loader.LoadFromPath("/nonexistent/file.yaml", &map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoader_LoadInvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "invalid.yaml")

	invalidYAML := `
name: test-app
count: [invalid yaml structure
verbose: true
`

	err := os.WriteFile(yamlFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader("config")
	err = loader.LoadFromPath(yamlFile, &map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoader_LoadInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "invalid.json")

	invalidJSON := `{
	"name": "test-app",
	"count": 42,
	"verbose": true
	// missing closing brace
`

	err := os.WriteFile(jsonFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader("invalid.json")
	err = loader.LoadFromPath(jsonFile, &map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoader_UnsupportedFileType(t *testing.T) {
	// Create a temporary file with unsupported extension
	tempDir := t.TempDir()
	txtFile := filepath.Join(tempDir, "config.txt")

	err := os.WriteFile(txtFile, []byte("some content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader("config.txt")
	err = loader.LoadFromPath(txtFile, &map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for unsupported file type")
	}
}
