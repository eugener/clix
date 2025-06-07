package help

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/eugener/clix/internal/bind"
)

// Test structs for validation testing
type ValidTestConfig struct {
	Name     string `posix:"n,name,Name of the item,required"`
	Verbose  bool   `posix:"v,verbose,Enable verbose output"`
	Output   string `posix:"o,output,Output format,default=text|choices=text;json;yaml"`
	Count    int    `posix:"c,count,Number of items,default=1"`
	Optional string `posix:",optional,Optional field"`
}

type InvalidTestConfig struct {
	RequiredField string `posix:",requiredfield,Required field,required"`
	ChoiceField   string `posix:",choicefield,Choice field,choices=a;b;c"`
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}

	if validator.analyzer == nil {
		t.Error("Validator analyzer is nil")
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		ve   ValidationError
		want string
	}{
		{
			name: "basic error",
			ve: ValidationError{
				Field:   "name",
				Value:   "",
				Message: "field is required",
			},
			want: "validation error for field name: field is required",
		},
		{
			name: "with complex value",
			ve: ValidationError{
				Field:   "choices",
				Value:   "invalid",
				Message: "must be one of: a, b, c",
			},
			want: "validation error for field choices: must be one of: a, b, c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ve.Error()
			if result != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name string
		ves  ValidationErrors
		want string
	}{
		{
			name: "empty errors",
			ves:  ValidationErrors{},
			want: "",
		},
		{
			name: "single error",
			ves: ValidationErrors{
				{Field: "name", Message: "field is required"},
			},
			want: "validation error for field name: field is required",
		},
		{
			name: "multiple errors",
			ves: ValidationErrors{
				{Field: "name", Message: "field is required"},
				{Field: "count", Message: "must be positive"},
			},
			want: "multiple validation errors:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ves.Error()
			if !strings.Contains(result, tt.want) {
				t.Errorf("ValidationErrors.Error() = %v, want to contain %v", result, tt.want)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		config  any
		wantErr bool
		errType string
	}{
		{
			name: "valid struct with all fields set",
			config: ValidTestConfig{
				Name:     "test",
				Verbose:  true,
				Output:   "json",
				Count:    5,
				Optional: "value",
			},
			wantErr: false,
		},
		{
			name: "valid struct with minimal required fields",
			config: ValidTestConfig{
				Name: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid struct missing required field",
			config: ValidTestConfig{
				Verbose: true,
				// Name is missing but required
			},
			wantErr: true,
			errType: "required",
		},
		{
			name: "invalid struct with bad choice",
			config: ValidTestConfig{
				Name:   "test",
				Output: "xml", // Not in choices
			},
			wantErr: true,
			errType: "must be one of",
		},
		{
			name:    "non-struct config",
			config:  "not a struct",
			wantErr: true,
			errType: "struct",
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errType: "struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				if !strings.Contains(err.Error(), tt.errType) {
					t.Errorf("Validator.Validate() error = %v, expected to contain %v", err, tt.errType)
				}
			}
		})
	}
}

func TestValidator_ValidateChoices(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		fieldInfo bind.FieldInfo
		value     any
		wantErr   bool
	}{
		{
			name: "valid choice",
			fieldInfo: bind.FieldInfo{
				Name:    "format",
				Choices: []string{"json", "yaml", "text"},
			},
			value:   "json",
			wantErr: false,
		},
		{
			name: "invalid choice",
			fieldInfo: bind.FieldInfo{
				Name:    "format",
				Choices: []string{"json", "yaml", "text"},
			},
			value:   "xml",
			wantErr: true,
		},
		{
			name: "no choices defined",
			fieldInfo: bind.FieldInfo{
				Name:    "format",
				Choices: nil, // Use nil instead of empty slice to properly test no choices
			},
			value:   "anything",
			wantErr: false,
		},
		{
			name: "numeric value in choices",
			fieldInfo: bind.FieldInfo{
				Name:    "level",
				Choices: []string{"1", "2", "3"},
			},
			value:   1,
			wantErr: false, // Should convert to string for comparison
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateChoices(tt.fieldInfo, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateChoices() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Error("Expected ValidationError")
				}
			}
		})
	}
}

func TestValidator_IsZeroValue(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"empty string", "", true},
		{"non-empty string", "value", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"nil slice", []string(nil), true},
		{"empty slice", []string{}, false}, // Go considers empty slice as not zero
		{"non-empty slice", []string{"item"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := reflect.ValueOf(tt.value)
			result := validator.isZeroValue(value)

			if result != tt.want {
				t.Errorf("isZeroValue() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNewValidatorRegistry(t *testing.T) {
	registry := NewValidatorRegistry()
	if registry == nil {
		t.Fatal("NewValidatorRegistry returned nil")
	}

	if registry.validators == nil {
		t.Error("Registry validators map is nil")
	}
}

func TestValidatorRegistry_Register(t *testing.T) {
	registry := NewValidatorRegistry()

	validator := func(value any) error {
		return nil
	}

	registry.Register("test", validator)

	retrieved, exists := registry.Get("test")
	if !exists {
		t.Error("Validator not found after registration")
	}

	if retrieved == nil {
		t.Error("Retrieved validator is nil")
	}
}

func TestValidatorRegistry_Get(t *testing.T) {
	registry := NewValidatorRegistry()

	// Test getting non-existent validator
	_, exists := registry.Get("nonexistent")
	if exists {
		t.Error("Expected validator to not exist")
	}

	// Test getting existing validator
	validator := func(value any) error {
		return nil
	}
	registry.Register("existing", validator)

	retrieved, exists := registry.Get("existing")
	if !exists {
		t.Error("Expected validator to exist")
	}

	if retrieved == nil {
		t.Error("Retrieved validator should not be nil")
	}
}

func TestEmailValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid email - no @", "test.example.com", true},
		{"invalid email - no domain", "test@", true},
		{"invalid email - no dot", "test@example", true},
		{"non-string value", 123, true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EmailValidator(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("EmailValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestURLValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid http URL", "http://example.com", false},
		{"valid https URL", "https://example.com", false},
		{"invalid URL - no protocol", "example.com", true},
		{"invalid URL - wrong protocol", "ftp://example.com", true},
		{"non-string value", 123, true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := URLValidator(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("URLValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRangeValidator(t *testing.T) {
	validator := RangeValidator(0, 100)

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid int", 50, false},
		{"valid float", 75.5, false},
		{"valid string number", "25", false},
		{"below range", -10, true},
		{"above range", 150, true},
		{"invalid string", "not a number", true},
		{"non-numeric value", []string{}, true},
		{"boundary min", 0, false},
		{"boundary max", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("RangeValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLengthValidator(t *testing.T) {
	validator := LengthValidator(2, 10)

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid length", "hello", false},
		{"min length", "hi", false},
		{"max length", "1234567890", false},
		{"too short", "a", true},
		{"too long", "12345678901", true},
		{"non-string value", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("LengthValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatternValidator(t *testing.T) {
	validator := PatternValidator("alphanumeric")

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid alphanumeric", "abc123", false},
		{"valid letters only", "abcDEF", false},
		{"valid numbers only", "123456", false},
		{"invalid with special chars", "abc-123", true},
		{"invalid with spaces", "abc 123", true},
		{"non-string value", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("PatternValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
