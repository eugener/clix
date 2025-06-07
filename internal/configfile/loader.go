package configfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from various file formats
type Loader struct {
	searchPaths []string
	fileName    string
}

// NewLoader creates a new configuration file loader
func NewLoader(fileName string, searchPaths ...string) *Loader {
	if len(searchPaths) == 0 {
		// Default search paths
		searchPaths = []string{
			".",             // Current directory
			"$HOME/.config", // User config directory
			"/etc",          // System config directory
		}
	}

	return &Loader{
		searchPaths: searchPaths,
		fileName:    fileName,
	}
}

// Load loads configuration from file into the target struct
func (l *Loader) Load(target any) error {
	// Find configuration file
	configPath, err := l.findConfigFile()
	if err != nil {
		return err
	}

	if configPath == "" {
		// No config file found, that's okay
		return nil
	}

	// Read file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(configPath))

	switch ext {
	case ".yaml", ".yml":
		return l.loadYAML(content, target)
	case ".json":
		return l.loadJSON(content, target)
	case ".toml":
		return l.loadTOML(content, target)
	default:
		// Try to detect format from content
		return l.loadWithAutoDetect(content, target)
	}
}

// LoadFromPath loads configuration from a specific file path
func (l *Loader) LoadFromPath(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		return l.loadYAML(content, target)
	case ".json":
		return l.loadJSON(content, target)
	case ".toml":
		return l.loadTOML(content, target)
	default:
		return l.loadWithAutoDetect(content, target)
	}
}

// findConfigFile searches for configuration file in search paths
func (l *Loader) findConfigFile() (string, error) {
	// Expand environment variables in search paths
	expandedPaths := make([]string, len(l.searchPaths))
	for i, path := range l.searchPaths {
		expandedPaths[i] = os.ExpandEnv(path)
	}

	// Try different extensions
	extensions := []string{".yaml", ".yml", ".json", ".toml", ""}

	for _, searchPath := range expandedPaths {
		for _, ext := range extensions {
			configPath := filepath.Join(searchPath, l.fileName+ext)
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}
	}

	return "", nil // No file found
}

// loadYAML loads YAML configuration
func (l *Loader) loadYAML(content []byte, target any) error {
	// First unmarshal to a map to handle CLI-specific struct tags
	var yamlData map[string]any
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Map YAML keys to struct fields using struct tags
	return l.mapToStruct(yamlData, target)
}

// loadJSON loads JSON configuration
func (l *Loader) loadJSON(content []byte, target any) error {
	// First unmarshal to a map to handle CLI-specific struct tags
	var jsonData map[string]any
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Map JSON keys to struct fields using struct tags
	return l.mapToStruct(jsonData, target)
}

// loadTOML loads TOML configuration (placeholder - would need TOML library)
func (l *Loader) loadTOML(content []byte, target any) error {
	// For now, return an error - TOML support would require external library
	return fmt.Errorf("TOML support not yet implemented")
}

// loadWithAutoDetect attempts to detect format and load
func (l *Loader) loadWithAutoDetect(content []byte, target any) error {
	// Try JSON first
	if err := l.loadJSON(content, target); err == nil {
		return nil
	}

	// Try YAML
	if err := l.loadYAML(content, target); err == nil {
		return nil
	}

	return fmt.Errorf("unable to detect configuration file format")
}

// mapToStruct maps configuration data to struct fields using struct tags
func (l *Loader) mapToStruct(data map[string]any, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	targetType := targetStruct.Type()

	// Build mapping from config keys to struct fields
	fieldMap := l.buildFieldMapping(targetType)

	// Set values from config data
	for configKey, configValue := range data {
		if fieldInfo, exists := fieldMap[configKey]; exists {
			if err := l.setFieldValue(targetStruct, fieldInfo, configValue); err != nil {
				return fmt.Errorf("failed to set field %s: %w", fieldInfo.Name, err)
			}
		}
	}

	return nil
}

// FieldMapping contains information about struct field mapping
type FieldMapping struct {
	Name      string
	Index     int
	Type      reflect.Type
	JSONKey   string
	YAMLKey   string
	ConfigKey string
}

// buildFieldMapping builds a mapping from config keys to struct fields
func (l *Loader) buildFieldMapping(structType reflect.Type) map[string]*FieldMapping {
	fieldMap := make(map[string]*FieldMapping)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		mapping := &FieldMapping{
			Name:  field.Name,
			Index: i,
			Type:  field.Type,
		}

		// Check for various struct tags
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			mapping.JSONKey = parts[0]
			fieldMap[mapping.JSONKey] = mapping
		}

		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" && yamlTag != "-" {
			parts := strings.Split(yamlTag, ",")
			mapping.YAMLKey = parts[0]
			fieldMap[mapping.YAMLKey] = mapping
		}

		if configTag := field.Tag.Get("config"); configTag != "" && configTag != "-" {
			mapping.ConfigKey = configTag
			fieldMap[mapping.ConfigKey] = mapping
		}

		// Also check posix tag for CLI compatibility
		if posixTag := field.Tag.Get("posix"); posixTag != "" {
			parts := strings.Split(posixTag, ",")
			if len(parts) > 1 && parts[1] != "" {
				// Use long flag name as config key
				fieldMap[parts[1]] = mapping
			}
		}

		// Default to lowercase field name
		defaultKey := strings.ToLower(field.Name)
		if _, exists := fieldMap[defaultKey]; !exists {
			fieldMap[defaultKey] = mapping
		}

		// Also use exact field name
		if _, exists := fieldMap[field.Name]; !exists {
			fieldMap[field.Name] = mapping
		}
	}

	return fieldMap
}

// setFieldValue sets a field value with type conversion
func (l *Loader) setFieldValue(targetStruct reflect.Value, fieldInfo *FieldMapping, value any) error {
	field := targetStruct.Field(fieldInfo.Index)

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldInfo.Name)
	}

	// Convert value to appropriate type
	convertedValue, err := l.convertValue(value, fieldInfo.Type)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(convertedValue))
	return nil
}

// convertValue converts a value to the target type
func (l *Loader) convertValue(value any, targetType reflect.Type) (any, error) {
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}

	valueType := reflect.TypeOf(value)

	// If types match, return as-is
	if valueType == targetType {
		return value, nil
	}

	// Handle different type conversions
	switch targetType.Kind() {
	case reflect.String:
		return fmt.Sprintf("%v", value), nil

	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strings.ToLower(v) == "true" || v == "1" || v == "yes", nil
		case int, int64:
			return reflect.ValueOf(v).Int() != 0, nil
		case float64:
			return v != 0, nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int:
			return int64(v), nil
		case int64:
			return v, nil
		case float64:
			return int64(v), nil
		case string:
			return parseInt64(v)
		}

	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			return v, nil
		case float32:
			return float64(v), nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return parseFloat64(v)
		}

	case reflect.Slice:
		return l.convertSlice(value, targetType)

	case reflect.Map:
		return l.convertMap(value, targetType)
	}

	return nil, fmt.Errorf("cannot convert %T to %s", value, targetType)
}

// convertSlice converts value to slice type
func (l *Loader) convertSlice(value any, targetType reflect.Type) (any, error) {
	valueReflect := reflect.ValueOf(value)

	// If value is not a slice, wrap it
	if valueReflect.Kind() != reflect.Slice {
		slice := reflect.MakeSlice(targetType, 1, 1)
		elem, err := l.convertValue(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		slice.Index(0).Set(reflect.ValueOf(elem))
		return slice.Interface(), nil
	}

	// Convert each element
	slice := reflect.MakeSlice(targetType, valueReflect.Len(), valueReflect.Len())
	for i := 0; i < valueReflect.Len(); i++ {
		elem, err := l.convertValue(valueReflect.Index(i).Interface(), targetType.Elem())
		if err != nil {
			return nil, err
		}
		slice.Index(i).Set(reflect.ValueOf(elem))
	}

	return slice.Interface(), nil
}

// convertMap converts value to map type
func (l *Loader) convertMap(value any, targetType reflect.Type) (any, error) {
	valueReflect := reflect.ValueOf(value)

	if valueReflect.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot convert %T to map", value)
	}

	targetMap := reflect.MakeMap(targetType)

	for _, key := range valueReflect.MapKeys() {
		val := valueReflect.MapIndex(key)

		convertedKey, err := l.convertValue(key.Interface(), targetType.Key())
		if err != nil {
			return nil, err
		}

		convertedVal, err := l.convertValue(val.Interface(), targetType.Elem())
		if err != nil {
			return nil, err
		}

		targetMap.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedVal))
	}

	return targetMap.Interface(), nil
}

// Helper functions for parsing
func parseInt64(s string) (int64, error) {
	// Simple integer parsing
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseFloat64(s string) (float64, error) {
	// Simple float parsing
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// ConfigGenerator generates example configuration files
type ConfigGenerator struct{}

// NewConfigGenerator creates a new config generator
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// GenerateYAML generates a YAML configuration file from struct
func (cg *ConfigGenerator) GenerateYAML(structType reflect.Type) ([]byte, error) {
	example := cg.generateExampleStruct(structType)
	return yaml.Marshal(example)
}

// GenerateJSON generates a JSON configuration file from struct
func (cg *ConfigGenerator) GenerateJSON(structType reflect.Type) ([]byte, error) {
	example := cg.generateExampleStruct(structType)
	return json.MarshalIndent(example, "", "  ")
}

// generateExampleStruct creates an example configuration structure
func (cg *ConfigGenerator) generateExampleStruct(structType reflect.Type) map[string]any {
	example := make(map[string]any)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		if !field.IsExported() {
			continue
		}

		// Get config key
		configKey := cg.getConfigKey(field)
		if configKey == "" {
			configKey = strings.ToLower(field.Name)
		}

		// Generate example value
		exampleValue := cg.generateExampleValue(field.Type, field.Tag)
		example[configKey] = exampleValue
	}

	return example
}

// getConfigKey gets the configuration key for a field
func (cg *ConfigGenerator) getConfigKey(field reflect.StructField) string {
	// Check various tag sources
	if yamlTag := field.Tag.Get("yaml"); yamlTag != "" && yamlTag != "-" {
		return strings.Split(yamlTag, ",")[0]
	}

	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		return strings.Split(jsonTag, ",")[0]
	}

	if configTag := field.Tag.Get("config"); configTag != "" && configTag != "-" {
		return configTag
	}

	// Check posix tag for CLI compatibility
	if posixTag := field.Tag.Get("posix"); posixTag != "" {
		parts := strings.Split(posixTag, ",")
		if len(parts) > 1 && parts[1] != "" {
			return parts[1]
		}
	}

	return ""
}

// generateExampleValue generates an example value for a type
func (cg *ConfigGenerator) generateExampleValue(fieldType reflect.Type, tag reflect.StructTag) any {
	// Check for default value in posix tag
	if posixTag := tag.Get("posix"); posixTag != "" {
		parts := strings.Split(posixTag, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "default=") {
				return strings.TrimPrefix(part, "default=")
			}
			if strings.HasPrefix(part, "choices=") {
				choices := strings.Split(strings.TrimPrefix(part, "choices="), ";")
				if len(choices) > 0 {
					return choices[0]
				}
			}
		}
	}

	// Generate based on type
	switch fieldType.Kind() {
	case reflect.String:
		return "example_value"
	case reflect.Bool:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return 0
	case reflect.Float32, reflect.Float64:
		return 0.0
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			return []string{"item1", "item2"}
		}
		return []any{}
	case reflect.Map:
		return map[string]any{}
	default:
		return nil
	}
}
