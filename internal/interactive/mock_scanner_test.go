package interactive

import (
	"bufio"
	"strings"
	"testing"
)

// MockScanner simulates user input for testing
type MockScanner struct {
	inputs []string
	index  int
	err    error
}

func NewMockScanner(inputs ...string) *MockScanner {
	return &MockScanner{
		inputs: inputs,
		index:  0,
	}
}

func (ms *MockScanner) Scan() bool {
	if ms.index >= len(ms.inputs) {
		return false
	}
	ms.index++
	return true
}

func (ms *MockScanner) Text() string {
	if ms.index <= 0 || ms.index > len(ms.inputs) {
		return ""
	}
	return ms.inputs[ms.index-1]
}

func (ms *MockScanner) Err() error {
	return ms.err
}

func (ms *MockScanner) SetError(err error) {
	ms.err = err
}

// Helper function to create a prompter with mocked input
func createPrompterWithMockInput(inputs ...string) *Prompter {
	scanner := NewMockScanner(inputs...)
	prompter := NewPrompter()
	prompter.scanner = scanner
	return prompter
}

func TestConfirmPrompter_WithMockInput(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		defaultValue bool
		inputs       []string
		want         bool
		wantErr      bool
	}{
		{
			name:         "yes input",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{"yes"},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "no input",
			message:      "Continue?",
			defaultValue: true,
			inputs:       []string{"no"},
			want:         false,
			wantErr:      false,
		},
		{
			name:         "y input",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{"y"},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "n input",
			message:      "Continue?",
			defaultValue: true,
			inputs:       []string{"n"},
			want:         false,
			wantErr:      false,
		},
		{
			name:         "empty input uses default true",
			message:      "Continue?",
			defaultValue: true,
			inputs:       []string{""},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "empty input uses default false",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{""},
			want:         false,
			wantErr:      false,
		},
		{
			name:         "invalid then valid input",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{"maybe", "yes"},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "case insensitive YES",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{"YES"},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "numeric 1 for true",
			message:      "Continue?",
			defaultValue: false,
			inputs:       []string{"1"},
			want:         true,
			wantErr:      false,
		},
		{
			name:         "numeric 0 for false",
			message:      "Continue?",
			defaultValue: true,
			inputs:       []string{"0"},
			want:         false,
			wantErr:      false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewMockScanner(tt.inputs...)
			prompter := &ConfirmPrompter{scanner: scanner}
			
			got, err := prompter.Confirm(tt.message, tt.defaultValue)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectPrompter_WithMockInput(t *testing.T) {
	tests := []struct {
		name    string
		message string
		options []string
		inputs  []string
		want    string
		wantErr bool
	}{
		{
			name:    "select by number",
			message: "Choose option",
			options: []string{"apple", "banana", "cherry"},
			inputs:  []string{"2"},
			want:    "banana",
			wantErr: false,
		},
		{
			name:    "select by name",
			message: "Choose option",
			options: []string{"apple", "banana", "cherry"},
			inputs:  []string{"cherry"},
			want:    "cherry",
			wantErr: false,
		},
		{
			name:    "case insensitive name selection",
			message: "Choose option",
			options: []string{"Apple", "Banana", "Cherry"},
			inputs:  []string{"apple"},
			want:    "Apple",
			wantErr: false,
		},
		{
			name:    "invalid then valid selection",
			message: "Choose option",
			options: []string{"apple", "banana", "cherry"},
			inputs:  []string{"4", "invalid", "1"},
			want:    "apple",
			wantErr: false,
		},
		{
			name:    "first option",
			message: "Choose option",
			options: []string{"apple", "banana", "cherry"},
			inputs:  []string{"1"},
			want:    "apple",
			wantErr: false,
		},
		{
			name:    "last option",
			message: "Choose option",
			options: []string{"apple", "banana", "cherry"},
			inputs:  []string{"3"},
			want:    "cherry",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewMockScanner(tt.inputs...)
			prompter := &SelectPrompter{scanner: scanner}
			
			got, err := prompter.Select(tt.message, tt.options)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("Select() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultilinePrompter_WithMockInput(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		endMarker string
		inputs    []string
		want      string
		wantErr   bool
	}{
		{
			name:      "simple multiline",
			message:   "Enter text",
			endMarker: "EOF",
			inputs:    []string{"line 1", "line 2", "EOF"},
			want:      "line 1\nline 2",
			wantErr:   false,
		},
		{
			name:      "custom end marker",
			message:   "Enter text",
			endMarker: "END",
			inputs:    []string{"hello", "world", "END"},
			want:      "hello\nworld",
			wantErr:   false,
		},
		{
			name:      "empty input",
			message:   "Enter text",
			endMarker: "EOF",
			inputs:    []string{"EOF"},
			want:      "",
			wantErr:   false,
		},
		{
			name:      "single line",
			message:   "Enter text",
			endMarker: "EOF",
			inputs:    []string{"only line", "EOF"},
			want:      "only line",
			wantErr:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewMockScanner(tt.inputs...)
			prompter := &MultilinePrompter{scanner: scanner}
			
			got, err := prompter.PromptMultiline(tt.message, tt.endMarker)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptMultiline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got != tt.want {
				t.Errorf("PromptMultiline() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSmartPrompter_WithMockInput(t *testing.T) {
	t.Run("PromptMissing delegates to basic prompter", func(t *testing.T) {
		smartPrompter := NewSmartPrompter()
		
		// Test with invalid target (should return error)
		err := smartPrompter.PromptMissing("not a pointer")
		if err == nil {
			t.Error("Expected error for invalid target")
		}
	})
	
	t.Run("Confirm delegates to confirm prompter", func(t *testing.T) {
		scanner := NewMockScanner("yes")
		smartPrompter := NewSmartPrompter()
		smartPrompter.confirm.scanner = scanner
		
		got, err := smartPrompter.Confirm("Continue?", false)
		if err != nil {
			t.Errorf("Confirm() error = %v", err)
		}
		if !got {
			t.Error("Expected true for 'yes' input")
		}
	})
	
	t.Run("Select delegates to select prompter", func(t *testing.T) {
		scanner := NewMockScanner("1")
		smartPrompter := NewSmartPrompter()
		smartPrompter.select_.scanner = scanner
		
		got, err := smartPrompter.Select("Choose", []string{"option1", "option2"})
		if err != nil {
			t.Errorf("Select() error = %v", err)
		}
		if got != "option1" {
			t.Errorf("Expected 'option1', got %v", got)
		}
	})
	
	t.Run("PromptMultiline delegates to multiline prompter", func(t *testing.T) {
		scanner := NewMockScanner("line1", "EOF")
		smartPrompter := NewSmartPrompter()
		smartPrompter.multiline.scanner = scanner
		
		got, err := smartPrompter.PromptMultiline("Enter text", "EOF")
		if err != nil {
			t.Errorf("PromptMultiline() error = %v", err)
		}
		if got != "line1" {
			t.Errorf("Expected 'line1', got %v", got)
		}
	})
}

// Test error scenarios
func TestPrompter_ErrorScenarios(t *testing.T) {
	t.Run("ConfirmPrompter scan failure", func(t *testing.T) {
		scanner := NewMockScanner()
		prompter := &ConfirmPrompter{scanner: scanner}
		
		_, err := prompter.Confirm("Continue?", false)
		if err == nil {
			t.Error("Expected error when scanner fails")
		}
		
		if !strings.Contains(err.Error(), "failed to read input") {
			t.Errorf("Expected 'failed to read input' error, got: %v", err)
		}
	})
	
	t.Run("SelectPrompter scan failure", func(t *testing.T) {
		scanner := NewMockScanner()
		prompter := &SelectPrompter{scanner: scanner}
		
		_, err := prompter.Select("Choose", []string{"option1"})
		if err == nil {
			t.Error("Expected error when scanner fails")
		}
		
		if !strings.Contains(err.Error(), "failed to read input") {
			t.Errorf("Expected 'failed to read input' error, got: %v", err)
		}
	})
}

// Test with real bufio.Scanner using string reader
func TestPrompter_WithStringReader(t *testing.T) {
	input := "test input\n"
	reader := strings.NewReader(input)
	scanner := bufio.NewScanner(reader)
	
	prompter := &Prompter{
		scanner:  scanner,
		analyzer: NewPrompter().analyzer,
	}
	
	// Test that we can create a prompter with a real scanner
	if prompter.scanner == nil {
		t.Error("Scanner should not be nil")
	}
	
	if prompter.analyzer == nil {
		t.Error("Analyzer should not be nil")
	}
}