package bind

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// FieldInfo contains metadata about a struct field
type FieldInfo struct {
	Name        string
	Type        reflect.Type
	Short       string
	Long        string
	Description string
	Required    bool
	Default     string
	Choices     []string
	Hidden      bool
	Positional  bool
	Environment string
	Validator   func(any) error
}

// StructMetadata contains all field information for a struct
type StructMetadata struct {
	Fields      []FieldInfo
	FieldMap    map[string]*FieldInfo
	ShortMap    map[string]*FieldInfo
	Positional  []*FieldInfo
	Environment map[string]*FieldInfo
}

// Analyzer provides reflection-based struct analysis
type Analyzer struct {
	tagName string
}

// NewAnalyzer creates a new struct analyzer
func NewAnalyzer(tagName string) *Analyzer {
	if tagName == "" {
		tagName = "posix"
	}
	return &Analyzer{tagName: tagName}
}

// Analyze extracts metadata from a struct type
func (a *Analyzer) Analyze(structType reflect.Type) (*StructMetadata, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", structType.Kind())
	}

	metadata := &StructMetadata{
		Fields:      make([]FieldInfo, 0),
		FieldMap:    make(map[string]*FieldInfo),
		ShortMap:    make(map[string]*FieldInfo),
		Positional:  make([]*FieldInfo, 0),
		Environment: make(map[string]*FieldInfo),
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldInfo, err := a.parseField(field)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %w", field.Name, err)
		}

		if fieldInfo == nil {
			continue // Skip fields without tags
		}

		metadata.Fields = append(metadata.Fields, *fieldInfo)

		// Build lookup maps
		if fieldInfo.Positional {
			metadata.Positional = append(metadata.Positional, fieldInfo)
		} else {
			metadata.FieldMap[fieldInfo.Long] = fieldInfo
			if fieldInfo.Short != "" {
				if _, exists := metadata.ShortMap[fieldInfo.Short]; exists {
					return nil, fmt.Errorf("duplicate short flag: -%s", fieldInfo.Short)
				}
				metadata.ShortMap[fieldInfo.Short] = fieldInfo
			}
		}

		if fieldInfo.Environment != "" {
			metadata.Environment[fieldInfo.Environment] = fieldInfo
		}
	}

	return metadata, nil
}

// parseField extracts FieldInfo from a reflect.StructField
func (a *Analyzer) parseField(field reflect.StructField) (*FieldInfo, error) {
	tag := field.Tag.Get(a.tagName)
	if tag == "" || tag == "-" {
		return nil, nil // Skip fields without tags
	}

	info := &FieldInfo{
		Name: field.Name,
		Type: field.Type,
		Long: strings.ToLower(field.Name), // Default long name
	}

	// Parse tag: "short,long,description,flags"
	parts := strings.Split(tag, ",")

	// Short flag
	if len(parts) > 0 && parts[0] != "" {
		info.Short = parts[0]
	}

	// Long flag
	if len(parts) > 1 && parts[1] != "" {
		info.Long = parts[1]
	}

	// Description
	if len(parts) > 2 {
		info.Description = parts[2]
	}

	// Flags
	if len(parts) > 3 {
		flagStr := parts[3]
		if err := a.parseFlags(info, flagStr); err != nil {
			return nil, err
		}
	}

	// Validate field configuration
	if err := a.validateField(info); err != nil {
		return nil, err
	}

	return info, nil
}

// parseFlags parses the flags portion of the tag
func (a *Analyzer) parseFlags(info *FieldInfo, flagStr string) error {
	flags := strings.Split(flagStr, "|")

	for _, flag := range flags {
		flag = strings.TrimSpace(flag)

		switch {
		case flag == "required":
			info.Required = true
		case flag == "hidden":
			info.Hidden = true
		case flag == "positional":
			info.Positional = true
		case strings.HasPrefix(flag, "default="):
			info.Default = strings.TrimPrefix(flag, "default=")
		case strings.HasPrefix(flag, "env="):
			info.Environment = strings.TrimPrefix(flag, "env=")
		case strings.HasPrefix(flag, "choices="):
			choicesStr := strings.TrimPrefix(flag, "choices=")
			info.Choices = strings.Split(choicesStr, ";")
		default:
			return fmt.Errorf("unknown flag: %s", flag)
		}
	}

	return nil
}

// validateField validates the field configuration
func (a *Analyzer) validateField(info *FieldInfo) error {
	// Positional fields can't have short/long flags
	if info.Positional && (info.Short != "" || info.Long != strings.ToLower(info.Name)) {
		return fmt.Errorf("positional field %s cannot have short/long flags", info.Name)
	}

	// Required validation
	if info.Required && info.Default != "" {
		return fmt.Errorf("field %s cannot be both required and have a default", info.Name)
	}

	// Type-specific validation
	switch info.Type.Kind() {
	case reflect.Bool:
		if info.Default != "" && info.Default != "true" && info.Default != "false" {
			return fmt.Errorf("boolean field %s has invalid default: %s", info.Name, info.Default)
		}
	case reflect.Slice:
		if !info.Positional {
			return fmt.Errorf("slice field %s must be positional", info.Name)
		}
	}

	return nil
}

// Binder provides value binding functionality
type Binder struct {
	analyzer *Analyzer
}

// NewBinder creates a new value binder
func NewBinder(tagName string) *Binder {
	return &Binder{
		analyzer: NewAnalyzer(tagName),
	}
}

// BindValues binds parsed values to a struct
func (b *Binder) BindValues(target any, values map[string]any, positional []string) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	targetType := targetStruct.Type()

	metadata, err := b.analyzer.Analyze(targetType)
	if err != nil {
		return err
	}

	// Set flag values
	for flagName, value := range values {
		// Try to find by long name first
		fieldInfo, exists := metadata.FieldMap[flagName]
		if !exists {
			// Try to find by short name
			if shortFieldInfo, shortExists := metadata.ShortMap[flagName]; shortExists {
				fieldInfo = shortFieldInfo
				exists = true
			}
		}

		if !exists {
			continue // Skip unknown flags
		}

		field := targetStruct.FieldByName(fieldInfo.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		if err := b.setValue(field, fieldInfo.Type, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldInfo.Name, err)
		}
	}

	// Set positional values
	for i, fieldInfo := range metadata.Positional {
		field := targetStruct.FieldByName(fieldInfo.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// Handle slice types (remaining arguments)
		if fieldInfo.Type.Kind() == reflect.Slice {
			remaining := positional[i:]
			if err := b.setSliceValue(field, fieldInfo.Type, remaining); err != nil {
				return fmt.Errorf("error setting positional field %s: %w", fieldInfo.Name, err)
			}
			break
		}

		// Handle single positional argument
		if i < len(positional) {
			if err := b.setValue(field, fieldInfo.Type, positional[i]); err != nil {
				return fmt.Errorf("error setting positional field %s: %w", fieldInfo.Name, err)
			}
		}
	}

	// Apply defaults
	if err := b.applyDefaults(targetStruct, metadata); err != nil {
		return err
	}

	return nil
}

// setValue sets a single value on a reflect.Value
func (b *Binder) setValue(field reflect.Value, fieldType reflect.Type, value any) error {
	// Convert value to appropriate type
	convertedValue, err := b.convertValue(value, fieldType)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(convertedValue))
	return nil
}

// setSliceValue sets a slice value
func (b *Binder) setSliceValue(field reflect.Value, fieldType reflect.Type, values []string) error {
	elemType := fieldType.Elem()
	slice := reflect.MakeSlice(fieldType, len(values), len(values))

	for i, value := range values {
		convertedValue, err := b.convertValue(value, elemType)
		if err != nil {
			return err
		}
		slice.Index(i).Set(reflect.ValueOf(convertedValue))
	}

	field.Set(slice)
	return nil
}

// convertValue converts a value to the target type
func (b *Binder) convertValue(value any, targetType reflect.Type) (any, error) {
	// If value is already the correct type, return as-is
	valueType := reflect.TypeOf(value)
	if valueType == targetType {
		return value, nil
	}

	// Convert string values
	if valueType.Kind() == reflect.String {
		return b.convertFromString(value.(string), targetType)
	}

	// Try direct conversion
	valueReflect := reflect.ValueOf(value)
	if valueReflect.Type().ConvertibleTo(targetType) {
		return valueReflect.Convert(targetType).Interface(), nil
	}

	return nil, fmt.Errorf("cannot convert %T to %s", value, targetType)
}

// convertFromString converts string values to target types
func (b *Binder) convertFromString(value string, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 64)
		return int(v), err
	case reflect.Int8:
		v, err := strconv.ParseInt(value, 10, 8)
		return int8(v), err
	case reflect.Int16:
		v, err := strconv.ParseInt(value, 10, 16)
		return int16(v), err
	case reflect.Int32:
		v, err := strconv.ParseInt(value, 10, 32)
		return int32(v), err
	case reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Uint:
		v, err := strconv.ParseUint(value, 10, 64)
		return uint(v), err
	case reflect.Uint8:
		v, err := strconv.ParseUint(value, 10, 8)
		return uint8(v), err
	case reflect.Uint16:
		v, err := strconv.ParseUint(value, 10, 16)
		return uint16(v), err
	case reflect.Uint32:
		v, err := strconv.ParseUint(value, 10, 32)
		return uint32(v), err
	case reflect.Uint64:
		return strconv.ParseUint(value, 10, 64)
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		return float32(v), err
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType.Kind())
	}
}

// applyDefaults applies default values to unset fields
func (b *Binder) applyDefaults(targetStruct reflect.Value, metadata *StructMetadata) error {
	for _, fieldInfo := range metadata.Fields {
		if fieldInfo.Default == "" {
			continue
		}

		field := targetStruct.FieldByName(fieldInfo.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// Only apply default if field is zero value
		if !field.IsZero() {
			continue
		}

		defaultValue, err := b.convertFromString(fieldInfo.Default, fieldInfo.Type)
		if err != nil {
			return fmt.Errorf("invalid default value for field %s: %w", fieldInfo.Name, err)
		}

		field.Set(reflect.ValueOf(defaultValue))
	}

	return nil
}
