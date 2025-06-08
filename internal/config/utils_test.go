package config

import (
	"testing"
)

type TestConfig struct {
	Name     string `posix:"n,name,Name,required"`
	Count    int    `posix:"c,count,Count,default=1"`
	Enabled  bool   `posix:"e,enabled,Enabled"`
	Category string `posix:"cat,category,Category,choices=dev;staging;prod"`
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name     string
		target   *TestConfig
		base     TestConfig
		expected TestConfig
		wantErr  bool
	}{
		{
			name:   "merge non-zero base values to zero target fields",
			target: &TestConfig{Name: "target-name"},
			base:   TestConfig{Name: "base-name", Count: 5, Enabled: true},
			expected: TestConfig{Name: "target-name", Count: 5, Enabled: true}, // target name preserved, base values copied
		},
		{
			name:   "target values take precedence",
			target: &TestConfig{Name: "target-name", Count: 10},
			base:   TestConfig{Name: "base-name", Count: 5, Enabled: true},
			expected: TestConfig{Name: "target-name", Count: 10, Enabled: true}, // target values preserved
		},
		{
			name:   "empty target gets all base values",
			target: &TestConfig{},
			base:   TestConfig{Name: "base-name", Count: 5, Enabled: true},
			expected: TestConfig{Name: "base-name", Count: 5, Enabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeConfigs(tt.target, tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if tt.target.Name != tt.expected.Name {
					t.Errorf("Name: got %v, want %v", tt.target.Name, tt.expected.Name)
				}
				if tt.target.Count != tt.expected.Count {
					t.Errorf("Count: got %v, want %v", tt.target.Count, tt.expected.Count)
				}
				if tt.target.Enabled != tt.expected.Enabled {
					t.Errorf("Enabled: got %v, want %v", tt.target.Enabled, tt.expected.Enabled)
				}
			}
		})
	}
}

func TestMergeConfigs_Errors(t *testing.T) {
	tests := []struct {
		name   string
		target interface{}
		base   interface{}
	}{
		{
			name:   "non-pointer target",
			target: TestConfig{},
			base:   TestConfig{},
		},
		{
			name:   "non-struct target",
			target: new(string),
			base:   TestConfig{},
		},
		{
			name:   "different types",
			target: &TestConfig{},
			base:   struct{ Name string }{Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeConfigs(tt.target, tt.base)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid config with required field",
			config: TestConfig{Name: "test", Category: "dev"},
		},
		{
			name:    "missing required field",
			config:  TestConfig{Count: 5},
			wantErr: true,
			errMsg:  "required field Name is missing",
		},
		{
			name:    "invalid choice value",
			config:  TestConfig{Name: "test", Category: "invalid"},
			wantErr: true,
			errMsg:  "field Category must be one of",
		},
		{
			name:   "valid choice value",
			config: TestConfig{Name: "test", Category: "staging"},
		},
		{
			name:   "empty choice field (allowed)",
			config: TestConfig{Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg {
					// For choice validation, just check if the error contains the expected text
					if tt.errMsg == "field Category must be one of" {
						if err == nil || !contains(err.Error(), tt.errMsg) {
							t.Errorf("ValidateConfig() error = %v, want error containing %v", err, tt.errMsg)
						}
					} else {
						t.Errorf("ValidateConfig() error = %v, want %v", err, tt.errMsg)
					}
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}