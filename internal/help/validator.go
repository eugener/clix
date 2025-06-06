package help

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/eugener/clix/internal/bind"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   any
	Message string
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return ""
	}
	
	if len(ves) == 1 {
		return ves[0].Error()
	}
	
	var msgs []string
	for _, ve := range ves {
		msgs = append(msgs, ve.Error())
	}
	
	return fmt.Sprintf("multiple validation errors:\n- %s", strings.Join(msgs, "\n- "))
}

// Validator provides comprehensive validation functionality
type Validator struct {
	analyzer *bind.Analyzer
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		analyzer: bind.NewAnalyzer("posix"),
	}
}

// Validate validates a parsed configuration struct
func (v *Validator) Validate(config any) error {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}
	
	if configValue.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct")
	}
	
	metadata, err := v.analyzer.Analyze(configValue.Type())
	if err != nil {
		return err
	}
	
	var errors ValidationErrors
	
	// Validate each field
	for _, fieldInfo := range metadata.Fields {
		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() {
			continue
		}
		
		if err := v.validateField(fieldInfo, field); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errors = append(errors, *ve)
			} else {
				errors = append(errors, ValidationError{
					Field:   fieldInfo.Name,
					Value:   field.Interface(),
					Message: err.Error(),
				})
			}
		}
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}

// validateField validates a single field
func (v *Validator) validateField(fieldInfo bind.FieldInfo, field reflect.Value) error {
	value := field.Interface()
	
	// Check if required field is set
	if fieldInfo.Required && v.isZeroValue(field) {
		return &ValidationError{
			Field:   fieldInfo.Name,
			Value:   value,
			Message: "field is required",
		}
	}
	
	// Skip validation for zero values of optional fields
	if !fieldInfo.Required && v.isZeroValue(field) {
		return nil
	}
	
	// Validate choices
	if len(fieldInfo.Choices) > 0 {
		if err := v.validateChoices(fieldInfo, value); err != nil {
			return err
		}
	}
	
	// Type-specific validation
	switch fieldInfo.Type.Kind() {
	case reflect.String:
		return v.validateString(fieldInfo, value.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.validateInt(fieldInfo, field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.validateUint(fieldInfo, field.Uint())
	case reflect.Float32, reflect.Float64:
		return v.validateFloat(fieldInfo, field.Float())
	case reflect.Slice:
		return v.validateSlice(fieldInfo, field)
	}
	
	return nil
}

// validateChoices validates that a value is in the allowed choices
func (v *Validator) validateChoices(fieldInfo bind.FieldInfo, value any) error {
	valueStr := fmt.Sprintf("%v", value)
	
	for _, choice := range fieldInfo.Choices {
		if valueStr == choice {
			return nil
		}
	}
	
	return &ValidationError{
		Field:   fieldInfo.Name,
		Value:   value,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(fieldInfo.Choices, ", ")),
	}
}

// validateString validates string fields
func (v *Validator) validateString(fieldInfo bind.FieldInfo, value string) error {
	// Add string-specific validation rules here
	// For example: min/max length, regex patterns, etc.
	
	return nil
}

// validateInt validates integer fields
func (v *Validator) validateInt(fieldInfo bind.FieldInfo, value int64) error {
	// Add integer-specific validation rules here
	// For example: min/max values, ranges, etc.
	
	return nil
}

// validateUint validates unsigned integer fields
func (v *Validator) validateUint(fieldInfo bind.FieldInfo, value uint64) error {
	// Add unsigned integer-specific validation rules here
	
	return nil
}

// validateFloat validates float fields
func (v *Validator) validateFloat(fieldInfo bind.FieldInfo, value float64) error {
	// Add float-specific validation rules here
	
	return nil
}

// validateSlice validates slice fields
func (v *Validator) validateSlice(fieldInfo bind.FieldInfo, field reflect.Value) error {
	// Validate each element in the slice
	for i := 0; i < field.Len(); i++ {
		elem := field.Index(i)
		
		// Create a temporary field info for the element
		elemFieldInfo := fieldInfo
		elemFieldInfo.Type = fieldInfo.Type.Elem()
		
		if err := v.validateField(elemFieldInfo, elem); err != nil {
			return &ValidationError{
				Field:   fmt.Sprintf("%s[%d]", fieldInfo.Name, i),
				Value:   elem.Interface(),
				Message: err.Error(),
			}
		}
	}
	
	return nil
}

// isZeroValue checks if a value is the zero value for its type
func (v *Validator) isZeroValue(field reflect.Value) bool {
	return field.IsZero()
}

// CustomValidator represents a custom validation function
type CustomValidator func(value any) error

// ValidatorRegistry manages custom validators
type ValidatorRegistry struct {
	validators map[string]CustomValidator
}

// NewValidatorRegistry creates a new validator registry
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[string]CustomValidator),
	}
}

// Register registers a custom validator
func (vr *ValidatorRegistry) Register(name string, validator CustomValidator) {
	vr.validators[name] = validator
}

// Get retrieves a validator by name
func (vr *ValidatorRegistry) Get(name string) (CustomValidator, bool) {
	validator, exists := vr.validators[name]
	return validator, exists
}

// Built-in validators

// EmailValidator validates email addresses
func EmailValidator(value any) error {
	email, ok := value.(string)
	if !ok {
		return fmt.Errorf("email validator requires string value")
	}
	
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

// URLValidator validates URLs
func URLValidator(value any) error {
	url, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL validator requires string value")
	}
	
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	
	return nil
}

// RangeValidator validates that a numeric value is within a range
func RangeValidator(min, max float64) CustomValidator {
	return func(value any) error {
		var num float64
		
		switch v := value.(type) {
		case int:
			num = float64(v)
		case int64:
			num = float64(v)
		case float64:
			num = v
		case float32:
			num = float64(v)
		case string:
			var err error
			num, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("cannot parse number: %v", err)
			}
		default:
			return fmt.Errorf("range validator requires numeric value")
		}
		
		if num < min || num > max {
			return fmt.Errorf("value %.2f must be between %.2f and %.2f", num, min, max)
		}
		
		return nil
	}
}

// LengthValidator validates string length
func LengthValidator(min, max int) CustomValidator {
	return func(value any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("length validator requires string value")
		}
		
		length := len(str)
		if length < min || length > max {
			return fmt.Errorf("length %d must be between %d and %d", length, min, max)
		}
		
		return nil
	}
}

// PatternValidator validates against a regex pattern
func PatternValidator(pattern string) CustomValidator {
	return func(value any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("pattern validator requires string value")
		}
		
		// Simple pattern matching - in a real implementation,
		// you'd use regexp package
		if pattern == "alphanumeric" {
			for _, r := range str {
				if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
					return fmt.Errorf("value must be alphanumeric")
				}
			}
		}
		
		return nil
	}
}