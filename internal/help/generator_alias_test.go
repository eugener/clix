package help

import (
	"testing"
)

func TestGenerator_FormatCommandDisplayName(t *testing.T) {
	generator := NewGenerator(DefaultHelpConfig("test"))

	tests := []struct {
		name     string
		cmdName  string
		aliases  []string
		expected string
	}{
		{
			name:     "command without aliases",
			cmdName:  "deploy",
			aliases:  nil,
			expected: "deploy",
		},
		{
			name:     "command with empty aliases",
			cmdName:  "deploy",
			aliases:  []string{},
			expected: "deploy",
		},
		{
			name:     "command with single alias",
			cmdName:  "deploy",
			aliases:  []string{"d"},
			expected: "deploy, d",
		},
		{
			name:     "command with multiple aliases",
			cmdName:  "deploy",
			aliases:  []string{"d", "dep"},
			expected: "deploy, d, dep",
		},
		{
			name:     "command with many aliases",
			cmdName:  "kubernetes-deploy",
			aliases:  []string{"k8s-deploy", "k8s", "deploy"},
			expected: "kubernetes-deploy, k8s-deploy, k8s, deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.formatCommandDisplayName(tt.cmdName, tt.aliases)
			if result != tt.expected {
				t.Errorf("formatCommandDisplayName(%q, %v) = %q, want %q",
					tt.cmdName, tt.aliases, result, tt.expected)
			}
		})
	}
}
