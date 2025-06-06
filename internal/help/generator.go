package help

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"claude-code-test/internal/bind"
)

// HelpConfig configures help generation
type HelpConfig struct {
	ProgramName    string
	Version        string
	Description    string
	UsageTemplate  string
	ExamplesText   string
	Footer         string
	ColorEnabled   bool
	MaxWidth       int
}

// DefaultHelpConfig returns a default help configuration
func DefaultHelpConfig(programName string) *HelpConfig {
	return &HelpConfig{
		ProgramName:   programName,
		Version:       "1.0.0",
		Description:   "",
		UsageTemplate: DefaultUsageTemplate,
		ColorEnabled:  true,
		MaxWidth:      80,
	}
}

// Generator generates help text for commands
type Generator struct {
	config   *HelpConfig
	analyzer *bind.Analyzer
}

// NewGenerator creates a new help generator
func NewGenerator(config *HelpConfig) *Generator {
	if config == nil {
		config = DefaultHelpConfig("cli")
	}
	return &Generator{
		config:   config,
		analyzer: bind.NewAnalyzer("posix"),
	}
}

// GenerateMainHelp generates help for the main CLI
func (g *Generator) GenerateMainHelp(commands map[string]CommandInfo) string {
	var sb strings.Builder
	
	// Header
	if g.config.Description != "" {
		sb.WriteString(g.config.Description)
		sb.WriteString("\n\n")
	}
	
	// Usage
	sb.WriteString("Usage:\n")
	sb.WriteString(fmt.Sprintf("  %s <command> [options]\n\n", g.config.ProgramName))
	
	// Commands
	if len(commands) > 0 {
		sb.WriteString("Commands:\n")
		
		// Calculate max command name length for alignment
		maxLen := 0
		for name := range commands {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}
		
		// Sort commands alphabetically
		names := make([]string, 0, len(commands))
		for name := range commands {
			names = append(names, name)
		}
		sort.Strings(names)
		
		// Format commands
		for _, name := range names {
			cmd := commands[name]
			padding := strings.Repeat(" ", maxLen-len(name)+2)
			sb.WriteString(fmt.Sprintf("  %s%s%s\n", name, padding, cmd.Description))
		}
		sb.WriteString("\n")
	}
	
	// Global options (if any)
	sb.WriteString("Global Options:\n")
	sb.WriteString("  -h, --help     Show help\n")
	sb.WriteString("  -v, --version  Show version\n\n")
	
	// Footer
	if g.config.Footer != "" {
		sb.WriteString(g.config.Footer)
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// GenerateCommandHelp generates help for a specific command
func (g *Generator) GenerateCommandHelp(name string, info CommandInfo) (string, error) {
	// Analyze the command's config struct
	metadata, err := g.analyzer.Analyze(info.ConfigType)
	if err != nil {
		return "", fmt.Errorf("failed to analyze command config: %w", err)
	}
	
	// Prepare template data
	data := CommandHelpData{
		ProgramName: g.config.ProgramName,
		CommandName: name,
		Description: info.Description,
		Usage:       g.buildUsage(name, metadata),
		Flags:       g.buildFlagsHelp(metadata),
		Positional:  g.buildPositionalHelp(metadata),
		Examples:    info.Examples,
		MaxWidth:    g.config.MaxWidth,
	}
	
	// Execute template
	tmpl, err := template.New("command").Parse(g.config.UsageTemplate)
	if err != nil {
		return "", err
	}
	
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	
	return sb.String(), nil
}

// buildUsage builds the usage line
func (g *Generator) buildUsage(commandName string, metadata *bind.StructMetadata) string {
	var parts []string
	parts = append(parts, g.config.ProgramName, commandName)
	
	// Add flags placeholder
	if len(metadata.FieldMap) > 0 {
		parts = append(parts, "[options]")
	}
	
	// Add positional arguments
	for _, field := range metadata.Positional {
		if field.Type.Kind() == reflect.Slice {
			parts = append(parts, fmt.Sprintf("[%s...]", strings.ToUpper(field.Name)))
		} else {
			if field.Required {
				parts = append(parts, fmt.Sprintf("<%s>", strings.ToUpper(field.Name)))
			} else {
				parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(field.Name)))
			}
		}
	}
	
	return strings.Join(parts, " ")
}

// buildFlagsHelp builds the flags help section
func (g *Generator) buildFlagsHelp(metadata *bind.StructMetadata) []FlagHelp {
	var flags []FlagHelp
	
	// Collect all flags
	for _, field := range metadata.Fields {
		if field.Positional || field.Hidden {
			continue
		}
		
		flag := FlagHelp{
			Short:       field.Short,
			Long:        field.Long,
			Description: field.Description,
			Type:        g.getTypeString(field.Type),
			Required:    field.Required,
			Default:     field.Default,
			Choices:     field.Choices,
		}
		
		flags = append(flags, flag)
	}
	
	// Sort flags alphabetically by long name
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Long < flags[j].Long
	})
	
	return flags
}

// buildPositionalHelp builds the positional arguments help section
func (g *Generator) buildPositionalHelp(metadata *bind.StructMetadata) []PositionalHelp {
	var positional []PositionalHelp
	
	for _, field := range metadata.Positional {
		pos := PositionalHelp{
			Name:        field.Name,
			Description: field.Description,
			Type:        g.getTypeString(field.Type),
			Required:    field.Required,
		}
		
		positional = append(positional, pos)
	}
	
	return positional
}

// getTypeString returns a human-readable type string
func (g *Generator) getTypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Slice:
		elemType := g.getTypeString(t.Elem())
		return fmt.Sprintf("[]%s", elemType)
	default:
		return t.String()
	}
}

// FormatFlag formats a single flag for display
func (g *Generator) FormatFlag(flag FlagHelp) string {
	var parts []string
	
	// Build flag part
	if flag.Short != "" && flag.Long != "" {
		parts = append(parts, fmt.Sprintf("-%s, --%s", flag.Short, flag.Long))
	} else if flag.Short != "" {
		parts = append(parts, fmt.Sprintf("-%s", flag.Short))
	} else {
		parts = append(parts, fmt.Sprintf("--%s", flag.Long))
	}
	
	// Add type if not boolean
	if flag.Type != "bool" {
		parts = append(parts, fmt.Sprintf("<%s>", flag.Type))
	}
	
	flagStr := strings.Join(parts, " ")
	
	// Build description part
	var descParts []string
	if flag.Description != "" {
		descParts = append(descParts, flag.Description)
	}
	
	if flag.Required {
		descParts = append(descParts, "(required)")
	}
	
	if flag.Default != "" {
		descParts = append(descParts, fmt.Sprintf("(default: %s)", flag.Default))
	}
	
	if len(flag.Choices) > 0 {
		descParts = append(descParts, fmt.Sprintf("(choices: %s)", strings.Join(flag.Choices, ", ")))
	}
	
	description := strings.Join(descParts, " ")
	
	// Calculate padding
	const maxFlagWidth = 30
	flagWidth := len(flagStr)
	if flagWidth > maxFlagWidth {
		return fmt.Sprintf("  %s\n%s%s", flagStr, strings.Repeat(" ", maxFlagWidth+2), description)
	}
	
	padding := strings.Repeat(" ", maxFlagWidth-flagWidth)
	return fmt.Sprintf("  %s%s  %s", flagStr, padding, description)
}

// Data structures for template rendering

// CommandInfo contains metadata about a command
type CommandInfo struct {
	Name        string
	Description string
	ConfigType  reflect.Type
	Examples    []string
}

// CommandHelpData contains data for command help template
type CommandHelpData struct {
	ProgramName string
	CommandName string
	Description string
	Usage       string
	Flags       []FlagHelp
	Positional  []PositionalHelp
	Examples    []string
	MaxWidth    int
}

// FlagHelp contains flag help information
type FlagHelp struct {
	Short       string
	Long        string
	Description string
	Type        string
	Required    bool
	Default     string
	Choices     []string
}

// PositionalHelp contains positional argument help information
type PositionalHelp struct {
	Name        string
	Description string
	Type        string
	Required    bool
}

// DefaultUsageTemplate is the default template for command help
const DefaultUsageTemplate = `{{.Description}}

Usage:
  {{.Usage}}

{{- if .Flags}}

Options:
{{- range .Flags}}
  {{- if .Short}}
  -{{.Short}}, --{{.Long}}{{if ne .Type "bool"}} <{{.Type}}>{{end}}
  {{- else}}
      --{{.Long}}{{if ne .Type "bool"}} <{{.Type}}>{{end}}
  {{- end}}
    {{- if .Description}}
        {{.Description}}
    {{- end}}
    {{- if .Required}} (required){{end}}
    {{- if .Default}} (default: {{.Default}}){{end}}
    {{- if .Choices}} (choices: {{range $i, $c := .Choices}}{{if $i}}, {{end}}{{$c}}{{end}}){{end}}
{{- end}}
{{- end}}

{{- if .Positional}}

Arguments:
{{- range .Positional}}
  {{.Name}}{{if ne .Type "string"}} <{{.Type}}>{{end}}
    {{- if .Description}}
        {{.Description}}
    {{- end}}
    {{- if .Required}} (required){{end}}
{{- end}}
{{- end}}

{{- if .Examples}}

Examples:
{{- range .Examples}}
  {{.}}
{{- end}}
{{- end}}
`