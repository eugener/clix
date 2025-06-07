package interactive

import (
	"reflect"
	"strings"
	"testing"

	"github.com/eugener/clix/internal/bind"
)

// Test config structs
type TestConfig struct {
	Name     string   `short:"n" long:"name" description:"Name of the item" required:"true"`
	Verbose  bool     `short:"v" long:"verbose" description:"Enable verbose output"`
	Output   string   `short:"o" long:"output" description:"Output format" default:"text" choices:"text,json,yaml"`
	Count    int      `short:"c" long:"count" description:"Number of items" default:"1"`
	Tags     []string `long:"tag" description:"Tags to add"`
	Optional string   `long:"optional" description:"Optional field"`
}

type RequiredFieldsConfig struct {
	Database string `required:"true" description:"Database connection string"`
	Port     int    `required:"true" description:"Port number"`
	Debug    bool   `description:"Enable debug mode"`
}

func TestNewPrompter(t *testing.T) {
	prompter := NewPrompter()
	if prompter == nil {
		t.Fatal("NewPrompter returned nil")
	}

	if prompter.scanner == nil {
		t.Error("Prompter scanner is nil")
	}

	if prompter.analyzer == nil {
		t.Error("Prompter analyzer is nil")
	}
}

func TestPrompter_ValidateInput(t *testing.T) {
	prompter := NewPrompter()

	tests := []struct {
		name      string
		input     string
		fieldInfo bind.FieldInfo
		wantErr   bool
		errMsg    string
	}{
		{
			name:  "required field with empty input",
			input: "",
			fieldInfo: bind.FieldInfo{
				Required: true,
				Type:     reflect.TypeOf(""), // Add default type
			},
			wantErr: true,
			errMsg:  "required",
		},
		{
			name:  "optional field with empty input",
			input: "",
			fieldInfo: bind.FieldInfo{
				Required: false,
				Type:     reflect.TypeOf(""), // Add default type
			},
			wantErr: false,
		},
		{
			name:  "valid choice",
			input: "json",
			fieldInfo: bind.FieldInfo{
				Choices: []string{"json", "yaml", "text"},
				Type:    reflect.TypeOf(""), // Add default type
			},
			wantErr: false,
		},
		{
			name:  "invalid choice",
			input: "xml",
			fieldInfo: bind.FieldInfo{
				Choices: []string{"json", "yaml", "text"},
				Type:    reflect.TypeOf(""), // Add default type
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name:  "case insensitive choice",
			input: "JSON",
			fieldInfo: bind.FieldInfo{
				Choices: []string{"json", "yaml", "text"},
				Type:    reflect.TypeOf(""), // Add default type
			},
			wantErr: false,
		},
		{
			name:  "valid boolean true",
			input: "true",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(true),
			},
			wantErr: false,
		},
		{
			name:  "valid boolean false",
			input: "false",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(true),
			},
			wantErr: false,
		},
		{
			name:  "invalid boolean",
			input: "maybe",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(true),
			},
			wantErr: true,
			errMsg:  "must be true or false",
		},
		{
			name:  "valid integer",
			input: "42",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0),
			},
			wantErr: false,
		},
		{
			name:  "invalid integer",
			input: "not-a-number",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0),
			},
			wantErr: true,
			errMsg:  "must be a valid number",
		},
		{
			name:  "valid float",
			input: "3.14",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0.0),
			},
			wantErr: false,
		},
		{
			name:  "invalid float",
			input: "not-a-float",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0.0),
			},
			wantErr: true,
			errMsg:  "must be a valid decimal number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := prompter.validateInput(tt.input, tt.fieldInfo)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateInput() error = %v, expected to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestPrompter_SetFieldValue(t *testing.T) {
	prompter := NewPrompter()

	tests := []struct {
		name      string
		fieldType reflect.Type
		value     string
		wantValue any
		wantErr   bool
	}{
		{
			name:      "string field",
			fieldType: reflect.TypeOf(""),
			value:     "test-value",
			wantValue: "test-value",
			wantErr:   false,
		},
		{
			name:      "empty string field",
			fieldType: reflect.TypeOf(""),
			value:     "",
			wantValue: "",
			wantErr:   false,
		},
		{
			name:      "boolean field true",
			fieldType: reflect.TypeOf(true),
			value:     "true",
			wantValue: true,
			wantErr:   false,
		},
		{
			name:      "boolean field false",
			fieldType: reflect.TypeOf(true),
			value:     "false",
			wantValue: false,
			wantErr:   false,
		},
		{
			name:      "int field",
			fieldType: reflect.TypeOf(0),
			value:     "42",
			wantValue: int64(42),
			wantErr:   false,
		},
		{
			name:      "invalid int",
			fieldType: reflect.TypeOf(0),
			value:     "not-a-number",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "float field",
			fieldType: reflect.TypeOf(0.0),
			value:     "3.14",
			wantValue: 3.14,
			wantErr:   false,
		},
		{
			name:      "invalid float",
			fieldType: reflect.TypeOf(0.0),
			value:     "not-a-float",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "string slice",
			fieldType: reflect.TypeOf([]string{}),
			value:     "a,b,c",
			wantValue: []string{"a", "b", "c"},
			wantErr:   false,
		},
		{
			name:      "string slice with spaces",
			fieldType: reflect.TypeOf([]string{}),
			value:     "a, b , c",
			wantValue: []string{"a", "b", "c"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.New(tt.fieldType).Elem()
			err := prompter.setFieldValue(field, tt.fieldType, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var got any
				switch tt.fieldType.Kind() {
				case reflect.String:
					got = field.String()
				case reflect.Bool:
					got = field.Bool()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					got = field.Int()
				case reflect.Float32, reflect.Float64:
					got = field.Float()
				case reflect.Slice:
					if tt.fieldType.Elem().Kind() == reflect.String {
						slice := make([]string, field.Len())
						for i := 0; i < field.Len(); i++ {
							slice[i] = field.Index(i).String()
						}
						got = slice
					}
				}

				if !reflect.DeepEqual(got, tt.wantValue) {
					t.Errorf("setFieldValue() got = %v, want %v", got, tt.wantValue)
				}
			}
		})
	}
}

func TestPrompter_ParseBool(t *testing.T) {
	prompter := NewPrompter()

	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		// True values
		{"true", "true", true, false},
		{"t", "t", true, false},
		{"yes", "yes", true, false},
		{"y", "y", true, false},
		{"1", "1", true, false},
		{"on", "on", true, false},
		{"TRUE", "TRUE", true, false}, // Case insensitive

		// False values
		{"false", "false", false, false},
		{"f", "f", false, false},
		{"no", "no", false, false},
		{"n", "n", false, false},
		{"0", "0", false, false},
		{"off", "off", false, false},
		{"empty", "", false, false},
		{"FALSE", "FALSE", false, false}, // Case insensitive

		// Invalid values
		{"invalid", "maybe", false, true},
		{"invalid2", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prompter.parseBool(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrompter_GetTypeHint(t *testing.T) {
	prompter := NewPrompter()

	tests := []struct {
		name      string
		fieldInfo bind.FieldInfo
		want      string
	}{
		{
			name: "boolean type",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(true),
			},
			want: "true/false",
		},
		{
			name: "int type",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0),
			},
			want: "number",
		},
		{
			name: "float type",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(0.0),
			},
			want: "decimal",
		},
		{
			name: "string with choices",
			fieldInfo: bind.FieldInfo{
				Type:    reflect.TypeOf(""),
				Choices: []string{"a", "b", "c"},
			},
			want: "choice",
		},
		{
			name: "string without choices",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf(""),
			},
			want: "text",
		},
		{
			name: "slice type",
			fieldInfo: bind.FieldInfo{
				Type: reflect.TypeOf([]string{}),
			},
			want: "comma-separated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prompter.getTypeHint(tt.fieldInfo)
			if got != tt.want {
				t.Errorf("getTypeHint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrompter_BuildPrompt(t *testing.T) {
	prompter := NewPrompter()

	tests := []struct {
		name      string
		fieldInfo bind.FieldInfo
		want      []string // Strings that should be present
	}{
		{
			name: "basic required field",
			fieldInfo: bind.FieldInfo{
				Name:        "username",
				Description: "Your username",
				Required:    true,
				Type:        reflect.TypeOf(""), // Add default type
			},
			want: []string{"Your username", "*", ":"},
		},
		{
			name: "field with choices",
			fieldInfo: bind.FieldInfo{
				Name:    "format",
				Choices: []string{"json", "yaml"},
				Type:    reflect.TypeOf(""), // Add default type
			},
			want: []string{"format", "[json, yaml]", ":"},
		},
		{
			name: "field with default",
			fieldInfo: bind.FieldInfo{
				Name:    "port",
				Default: "8080",
				Type:    reflect.TypeOf(""), // Add default type
			},
			want: []string{"port", "(default: 8080)", ":"},
		},
		{
			name: "field with type hint",
			fieldInfo: bind.FieldInfo{
				Name: "count",
				Type: reflect.TypeOf(0),
			},
			want: []string{"count", "(number)", ":"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prompter.buildPrompt(tt.fieldInfo)

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("buildPrompt() output missing '%s'\\nFull output: %s", want, got)
				}
			}
		})
	}
}

func TestPrompter_SupportsColor(t *testing.T) {
	prompter := NewPrompter()

	// Test that function doesn't panic and returns a boolean
	result := prompter.supportsColor()
	if result != true && result != false {
		t.Error("supportsColor() should return a boolean")
	}
}

func TestNewConfirmPrompter(t *testing.T) {
	prompter := NewConfirmPrompter()
	if prompter == nil {
		t.Fatal("NewConfirmPrompter returned nil")
	}

	if prompter.scanner == nil {
		t.Error("ConfirmPrompter scanner is nil")
	}
}

func TestNewSelectPrompter(t *testing.T) {
	prompter := NewSelectPrompter()
	if prompter == nil {
		t.Fatal("NewSelectPrompter returned nil")
	}

	if prompter.scanner == nil {
		t.Error("SelectPrompter scanner is nil")
	}
}

func TestSelectPrompter_SelectWithNoOptions(t *testing.T) {
	prompter := NewSelectPrompter()

	_, err := prompter.Select("Choose an option", []string{})
	if err == nil {
		t.Error("Select() should return error for empty options")
	}

	if !strings.Contains(err.Error(), "no options") {
		t.Errorf("Expected error about no options, got: %v", err)
	}
}

func TestNewMultilinePrompter(t *testing.T) {
	prompter := NewMultilinePrompter()
	if prompter == nil {
		t.Fatal("NewMultilinePrompter returned nil")
	}

	if prompter.scanner == nil {
		t.Error("MultilinePrompter scanner is nil")
	}
}

func TestNewSmartPrompterDetailed(t *testing.T) {
	prompter := NewSmartPrompter()
	if prompter == nil {
		t.Fatal("NewSmartPrompter returned nil")
	}

	if prompter.basic == nil {
		t.Error("SmartPrompter basic is nil")
	}

	if prompter.confirm == nil {
		t.Error("SmartPrompter confirm is nil")
	}

	if prompter.select_ == nil {
		t.Error("SmartPrompter select_ is nil")
	}

	if prompter.multiline == nil {
		t.Error("SmartPrompter multiline is nil")
	}
}

// Test edge cases and error scenarios
func TestPrompter_EdgeCases(t *testing.T) {
	prompter := NewPrompter()

	t.Run("unsupported field type", func(t *testing.T) {
		field := reflect.New(reflect.TypeOf(complex64(1))).Elem()
		err := prompter.setFieldValue(field, reflect.TypeOf(complex64(1)), "1+2i")

		if err == nil {
			t.Error("Expected error for unsupported field type")
		}

		if !strings.Contains(err.Error(), "unsupported field type") {
			t.Errorf("Expected unsupported field type error, got: %v", err)
		}
	})

	t.Run("invalid boolean for parseBool", func(t *testing.T) {
		_, err := prompter.parseBool("invalid-bool")
		if err == nil {
			t.Error("Expected error for invalid boolean")
		}
	})

	t.Run("validateInput with complex field types", func(t *testing.T) {
		// Test validation with different reflect.Type kinds
		fieldInfo := bind.FieldInfo{
			Type: reflect.TypeOf(int8(0)),
		}

		err := prompter.validateInput("200", fieldInfo)
		if err != nil {
			t.Errorf("Unexpected error for int8: %v", err)
		}
	})
}

func TestPrompter_Integration(t *testing.T) {
	// These tests verify the overall flow without actual user input
	prompter := NewPrompter()

	t.Run("PromptForMissing with invalid target", func(t *testing.T) {
		// Test with non-pointer
		err := prompter.PromptForMissing("not a pointer")
		if err == nil {
			t.Error("Expected error for non-pointer target")
		}

		// Test with pointer to non-struct
		str := "not a struct"
		err = prompter.PromptForMissing(&str)
		if err == nil {
			t.Error("Expected error for pointer to non-struct")
		}
	})
}
