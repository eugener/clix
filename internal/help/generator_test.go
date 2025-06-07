package help

import (
	"reflect"
	"strings"
	"testing"
)

// Test config struct for testing
type TestConfig struct {
	Name     string   `short:"n" long:"name" description:"Name of the item" required:"true"`
	Verbose  bool     `short:"v" long:"verbose" description:"Enable verbose output"`
	Output   string   `short:"o" long:"output" description:"Output format" default:"text" choices:"text,json,yaml"`
	Count    int      `short:"c" long:"count" description:"Number of items" default:"1"`
	Tags     []string `long:"tag" description:"Tags to add"`
	Hidden   string   `long:"hidden" description:"Hidden field" hidden:"true"`
	File     string   `pos:"0" description:"Input file" required:"true"`
	Optional string   `pos:"1" description:"Optional argument"`
}

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config *HelpConfig
		want   bool
	}{
		{
			name: "with valid config",
			config: &HelpConfig{
				ProgramName:  "test-app",
				ColorEnabled: true,
				MaxWidth:     80,
			},
			want: true,
		},
		{
			name:   "with nil config",
			config: nil,
			want:   true, // Should use default config
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.config)
			if (generator != nil) != tt.want {
				t.Errorf("NewGenerator() = %v, want %v", generator != nil, tt.want)
			}
			if generator != nil {
				if generator.config == nil {
					t.Error("Generator config is nil")
				}
				if generator.analyzer == nil {
					t.Error("Generator analyzer is nil")
				}
			}
		})
	}
}

func TestDefaultHelpConfig(t *testing.T) {
	tests := []struct {
		name        string
		programName string
		wantName    string
		wantVersion string
		wantWidth   int
		wantColor   bool
	}{
		{
			name:        "basic program name",
			programName: "test-app",
			wantName:    "test-app",
			wantVersion: "1.0.0",
			wantWidth:   80,
			wantColor:   true,
		},
		{
			name:        "empty program name",
			programName: "",
			wantName:    "",
			wantVersion: "1.0.0",
			wantWidth:   80,
			wantColor:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultHelpConfig(tt.programName)
			if config == nil {
				t.Fatal("DefaultHelpConfig returned nil")
			}
			
			if config.ProgramName != tt.wantName {
				t.Errorf("Expected program name '%s', got '%s'", tt.wantName, config.ProgramName)
			}
			
			if config.Version != tt.wantVersion {
				t.Errorf("Expected version '%s', got '%s'", tt.wantVersion, config.Version)
			}
			
			if config.MaxWidth != tt.wantWidth {
				t.Errorf("Expected max width %d, got %d", tt.wantWidth, config.MaxWidth)
			}
			
			if config.ColorEnabled != tt.wantColor {
				t.Errorf("Expected ColorEnabled %v, got %v", tt.wantColor, config.ColorEnabled)
			}
			
			if config.UsageTemplate == "" {
				t.Error("Expected non-empty UsageTemplate")
			}
		})
	}
}

func TestGenerateMainHelp(t *testing.T) {
	tests := []struct {
		name     string
		config   *HelpConfig
		commands map[string]CommandInfo
		want     []string // Strings that should be present in output
	}{
		{
			name: "empty commands",
			config: &HelpConfig{
				ProgramName: "test-app",
				Description: "Test application",
			},
			commands: map[string]CommandInfo{},
			want:     []string{"Usage:", "test-app <command> [options]", "Test application", "Global Options:"},
		},
		{
			name: "with commands",
			config: &HelpConfig{
				ProgramName: "test-app",
				Description: "Test application",
			},
			commands: map[string]CommandInfo{
				"deploy": {
					Name:        "deploy",
					Description: "Deploy the application",
					ConfigType:  reflect.TypeOf(TestConfig{}),
				},
				"build": {
					Name:        "build",
					Description: "Build the application",
					ConfigType:  reflect.TypeOf(TestConfig{}),
				},
			},
			want: []string{"Commands:", "build", "deploy", "Deploy the application", "Build the application"},
		},
		{
			name: "with footer",
			config: &HelpConfig{
				ProgramName: "test-app",
				Footer:      "For more information, visit https://example.com",
			},
			commands: map[string]CommandInfo{},
			want:     []string{"For more information, visit https://example.com"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.config)
			result := generator.GenerateMainHelp(tt.commands)
			
			if result == "" {
				t.Error("GenerateMainHelp returned empty string")
			}
			
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("GenerateMainHelp() output missing '%s'\nFull output:\n%s", want, result)
				}
			}
		})
	}
}

func TestGenerateCommandHelp(t *testing.T) {
	tests := []struct {
		name    string
		config  *HelpConfig
		cmdName string
		cmdInfo CommandInfo
		want    []string
		wantErr bool
	}{
		{
			name: "basic command",
			config: &HelpConfig{
				ProgramName:   "test-app",
				UsageTemplate: DefaultUsageTemplate,
			},
			cmdName: "deploy",
			cmdInfo: CommandInfo{
				Name:        "deploy",
				Description: "Deploy the application",
				ConfigType:  reflect.TypeOf(TestConfig{}),
				Examples:    []string{"test-app deploy --name myapp file.yml"},
			},
			want: []string{
				"Deploy the application",
				"Usage:",
				"test-app deploy",
				"--name",
				"--verbose",
				"Examples:",
				"test-app deploy --name myapp file.yml",
			},
			wantErr: false,
		},
		{
			name: "invalid config type",
			config: &HelpConfig{
				ProgramName:   "test-app",
				UsageTemplate: DefaultUsageTemplate,
			},
			cmdName: "invalid",
			cmdInfo: CommandInfo{
				Name:       "invalid",
				ConfigType: reflect.TypeOf(0), // Invalid type (int instead of struct)
			},
			want:    nil,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.config)
			result, err := generator.GenerateCommandHelp(tt.cmdName, tt.cmdInfo)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCommandHelp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				for _, want := range tt.want {
					if !strings.Contains(result, want) {
						t.Errorf("GenerateCommandHelp() output missing '%s'\nFull output:\n%s", want, result)
					}
				}
			}
		})
	}
}

func TestGetTypeString(t *testing.T) {
	tests := []struct {
		name string
		type_ reflect.Type
		want  string
	}{
		{"string", reflect.TypeOf(""), "string"},
		{"bool", reflect.TypeOf(true), "bool"},
		{"int", reflect.TypeOf(0), "int"},
		{"int64", reflect.TypeOf(int64(0)), "int"},
		{"uint", reflect.TypeOf(uint(0)), "uint"},
		{"float64", reflect.TypeOf(0.0), "float"},
		{"[]string", reflect.TypeOf([]string{}), "[]string"},
		{"[]int", reflect.TypeOf([]int{}), "[]int"},
	}
	
	generator := NewGenerator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.getTypeString(tt.type_)
			if result != tt.want {
				t.Errorf("getTypeString() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestFormatFlag(t *testing.T) {
	tests := []struct {
		name string
		flag FlagHelp
		want []string // Strings that should be present
	}{
		{
			name: "flag with short and long",
			flag: FlagHelp{
				Short:       "n",
				Long:        "name",
				Description: "Name of the item",
				Type:        "string",
				Required:    true,
			},
			want: []string{"-n, --name", "<string>", "Name of the item", "(required)"},
		},
		{
			name: "boolean flag",
			flag: FlagHelp{
				Short:       "v",
				Long:        "verbose",
				Description: "Enable verbose output",
				Type:        "bool",
			},
			want: []string{"-v, --verbose", "Enable verbose output"},
		},
		{
			name: "flag with default and choices",
			flag: FlagHelp{
				Long:        "format",
				Description: "Output format",
				Type:        "string",
				Default:     "json",
				Choices:     []string{"json", "yaml", "text"},
			},
			want: []string{"--format", "(default: json)", "(choices: json, yaml, text)"},
		},
	}
	
	generator := NewGenerator(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.FormatFlag(tt.flag)
			
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("FormatFlag() output missing '%s'\nFull output: %s", want, result)
				}
			}
		})
	}
}

func TestBuildUsage(t *testing.T) {
	// This test requires access to buildUsage which is private
	// We'll test it indirectly through GenerateCommandHelp
	generator := NewGenerator(&HelpConfig{
		ProgramName:   "test-app",
		UsageTemplate: "{{.Usage}}",
	})
	
	cmdInfo := CommandInfo{
		Name:       "deploy",
		ConfigType: reflect.TypeOf(TestConfig{}),
	}
	
	result, err := generator.GenerateCommandHelp("deploy", cmdInfo)
	if err != nil {
		t.Fatalf("GenerateCommandHelp() error = %v", err)
	}
	
	// Should contain program name, command name, options, and positional args
	expected := []string{
		"test-app",
		"deploy",
		"[options]",
		"<FILE>", // Required positional
		"[OPTIONAL]", // Optional positional
	}
	
	for _, want := range expected {
		if !strings.Contains(result, want) {
			t.Errorf("Usage missing '%s'\nFull output: %s", want, result)
		}
	}
}

func TestHelpConfig_Validation(t *testing.T) {
	config := &HelpConfig{
		ProgramName:  "test-app",
		Version:      "1.0.0",
		Description:  "Test application",
		ColorEnabled: true,
		MaxWidth:     120,
	}
	
	if config.ProgramName != "test-app" {
		t.Errorf("Expected program name 'test-app', got '%s'", config.ProgramName)
	}
	
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}
	
	if config.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got '%s'", config.Description)
	}
	
	if !config.ColorEnabled {
		t.Error("Expected ColorEnabled to be true")
	}
	
	if config.MaxWidth != 120 {
		t.Errorf("Expected MaxWidth 120, got %d", config.MaxWidth)
	}
}

func TestDefaultUsageTemplateSections(t *testing.T) {
	if DefaultUsageTemplate == "" {
		t.Error("DefaultUsageTemplate should not be empty")
	}
	
	// Check that template contains expected sections
	expectedSections := []string{
		"{{.Description}}",
		"{{.Usage}}",
		"{{.Flags}}",
		"{{.Positional}}",
		"{{.Examples}}",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(DefaultUsageTemplate, section) {
			t.Errorf("DefaultUsageTemplate missing section: %s", section)
		}
	}
}