package complete

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/eugener/clix/core"
	"github.com/eugener/clix/internal/bind"
)

// CompletionType represents the type of completion
type CompletionType int

const (
	CompletionCommands CompletionType = iota
	CompletionFlags
	CompletionValues
	CompletionFiles
	CompletionDirectories
)

// CompletionItem represents a single completion suggestion
type CompletionItem struct {
	Value       string
	Description string
	Type        CompletionType
}

// Generator generates shell completions
type Generator struct {
	registry *core.Registry
	analyzer *bind.Analyzer
}

// NewGenerator creates a new completion generator
func NewGenerator(registry *core.Registry) *Generator {
	return &Generator{
		registry: registry,
		analyzer: bind.NewAnalyzer("posix"),
	}
}

// Complete generates completions for the given command line
func (g *Generator) Complete(args []string, cursorPos int) ([]CompletionItem, error) {
	if len(args) == 0 {
		return g.completeCommands(""), nil
	}
	
	// If we're completing the first argument, complete commands
	if len(args) == 1 {
		return g.completeCommands(args[0]), nil
	}
	
	// Otherwise, complete for the specific command
	commandName := args[0]
	commandArgs := args[1:]
	
	return g.completeForCommand(commandName, commandArgs, cursorPos)
}

// completeCommands returns command completions
func (g *Generator) completeCommands(prefix string) []CompletionItem {
	var items []CompletionItem
	
	for name, desc := range g.registry.ListCommands() {
		if strings.HasPrefix(name, prefix) {
			items = append(items, CompletionItem{
				Value:       name,
				Description: desc.GetDescription(),
				Type:        CompletionCommands,
			})
		}
	}
	
	return items
}

// completeForCommand generates completions for a specific command
func (g *Generator) completeForCommand(commandName string, args []string, cursorPos int) ([]CompletionItem, error) {
	descriptor, exists := g.registry.GetCommand(commandName)
	if !exists {
		return nil, fmt.Errorf("unknown command: %s", commandName)
	}
	
	// Analyze the command's config struct
	metadata, err := g.analyzer.Analyze(descriptor.GetConfigType())
	if err != nil {
		return nil, err
	}
	
	// Parse existing arguments to understand context
	context, err := g.parseArgsContext(args, metadata)
	if err != nil {
		return nil, err
	}
	
	// Determine what we're completing
	return g.determineCompletions(context, metadata)
}

// ArgContext represents the context of parsed arguments
type ArgContext struct {
	ParsedFlags   map[string]bool
	LastArg       string
	IsFlag        bool
	NeedsValue    *bind.FieldInfo
	PositionalPos int
}

// parseArgsContext analyzes the current argument context
func (g *Generator) parseArgsContext(args []string, metadata *bind.StructMetadata) (*ArgContext, error) {
	context := &ArgContext{
		ParsedFlags:   make(map[string]bool),
		PositionalPos: 0,
	}
	
	if len(args) == 0 {
		return context, nil
	}
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		// Check if this is the last argument (what we're completing)
		if i == len(args)-1 {
			context.LastArg = arg
			
			if strings.HasPrefix(arg, "-") {
				context.IsFlag = true
				
				// Check if this flag needs a value
				flagName := g.extractFlagName(arg)
				if fieldInfo := g.findFlagField(flagName, metadata); fieldInfo != nil {
					if fieldInfo.Type.Kind() != reflect.Bool {
						context.NeedsValue = fieldInfo
					}
				}
			}
			break
		}
		
		// Handle flags
		if strings.HasPrefix(arg, "-") {
			flagName := g.extractFlagName(arg)
			context.ParsedFlags[flagName] = true
			
			// Check if this flag has a value
			if fieldInfo := g.findFlagField(flagName, metadata); fieldInfo != nil {
				if fieldInfo.Type.Kind() != reflect.Bool {
					i++ // Skip the value
				}
			}
		} else {
			// Positional argument
			context.PositionalPos++
		}
		
		i++
	}
	
	return context, nil
}

// determineCompletions determines what completions to offer
func (g *Generator) determineCompletions(context *ArgContext, metadata *bind.StructMetadata) ([]CompletionItem, error) {
	var items []CompletionItem
	
	// If we need a value for a flag, complete that
	if context.NeedsValue != nil {
		return g.completeForFieldType(context.NeedsValue), nil
	}
	
	// If the last arg is a partial flag, complete flags
	if context.IsFlag {
		return g.completeFlags(context.LastArg, context.ParsedFlags, metadata), nil
	}
	
	// Otherwise, offer both flags and positional completions
	
	// Add available flags
	flagItems := g.completeFlags("", context.ParsedFlags, metadata)
	items = append(items, flagItems...)
	
	// Add positional completions if we have positional fields
	if context.PositionalPos < len(metadata.Positional) {
		fieldInfo := metadata.Positional[context.PositionalPos]
		positionalItems := g.completeForFieldType(fieldInfo)
		items = append(items, positionalItems...)
	}
	
	return items, nil
}

// completeFlags generates flag completions
func (g *Generator) completeFlags(prefix string, usedFlags map[string]bool, metadata *bind.StructMetadata) []CompletionItem {
	var items []CompletionItem
	
	for _, fieldInfo := range metadata.Fields {
		if fieldInfo.Positional || fieldInfo.Hidden {
			continue
		}
		
		// Skip already used flags
		if usedFlags[fieldInfo.Long] || (fieldInfo.Short != "" && usedFlags[fieldInfo.Short]) {
			continue
		}
		
		// Long flag
		longFlag := "--" + fieldInfo.Long
		if strings.HasPrefix(longFlag, prefix) {
			desc := fieldInfo.Description
			if fieldInfo.Required {
				desc += " (required)"
			}
			items = append(items, CompletionItem{
				Value:       longFlag,
				Description: desc,
				Type:        CompletionFlags,
			})
		}
		
		// Short flag
		if fieldInfo.Short != "" {
			shortFlag := "-" + fieldInfo.Short
			if strings.HasPrefix(shortFlag, prefix) {
				items = append(items, CompletionItem{
					Value:       shortFlag,
					Description: fieldInfo.Description,
					Type:        CompletionFlags,
				})
			}
		}
	}
	
	return items
}

// completeForFieldType generates completions based on field type
func (g *Generator) completeForFieldType(fieldInfo *bind.FieldInfo) []CompletionItem {
	var items []CompletionItem
	
	// If field has choices, use them
	if len(fieldInfo.Choices) > 0 {
		for _, choice := range fieldInfo.Choices {
			items = append(items, CompletionItem{
				Value:       choice,
				Description: fmt.Sprintf("Choice for %s", fieldInfo.Name),
				Type:        CompletionValues,
			})
		}
		return items
	}
	
	// Type-specific completions
	switch fieldInfo.Type.Kind() {
	case reflect.Bool:
		items = append(items, 
			CompletionItem{Value: "true", Type: CompletionValues},
			CompletionItem{Value: "false", Type: CompletionValues},
		)
	case reflect.String:
		// For strings, offer file completion by default
		items = append(items, CompletionItem{
			Value: "<filename>",
			Type:  CompletionFiles,
		})
	default:
		// For other types, provide a generic placeholder
		items = append(items, CompletionItem{
			Value: fmt.Sprintf("<%s>", fieldInfo.Type.String()),
			Type:  CompletionValues,
		})
	}
	
	return items
}

// Helper functions

// extractFlagName extracts the flag name from a flag argument
func (g *Generator) extractFlagName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		// Long flag
		name := arg[2:]
		if eqIndex := strings.Index(name, "="); eqIndex != -1 {
			name = name[:eqIndex]
		}
		return name
	} else if strings.HasPrefix(arg, "-") {
		// Short flag (take last character for bundled flags)
		flags := arg[1:]
		if len(flags) > 0 {
			return string(flags[len(flags)-1])
		}
	}
	return ""
}

// findFlagField finds a field by flag name (short or long)
func (g *Generator) findFlagField(flagName string, metadata *bind.StructMetadata) *bind.FieldInfo {
	// Try long name first
	if fieldInfo, exists := metadata.FieldMap[flagName]; exists {
		return fieldInfo
	}
	
	// Try short name
	if fieldInfo, exists := metadata.ShortMap[flagName]; exists {
		return fieldInfo
	}
	
	return nil
}

// Shell-specific generators

// GenerateBashCompletion generates bash completion script
func (g *Generator) GenerateBashCompletion(programName string) string {
	return fmt.Sprintf(`# Bash completion for %s
_%s_completions() {
    local cur prev words cword
    _init_completion || return

    # Get completions from the program
    local completions
    completions=$(%s __complete "${COMP_WORDS[@]:1}" 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        COMPREPLY=($(compgen -W "$completions" -- "$cur"))
    fi
}

complete -F _%s_completions %s`, 
		programName, programName, programName, programName, programName)
}

// GenerateZshCompletion generates zsh completion script
func (g *Generator) GenerateZshCompletion(programName string) string {
	return fmt.Sprintf(`#compdef %s

_%s() {
    local context state line
    
    # Get completions from the program
    local completions
    completions=($(%s __complete "${words[@]:1}" 2>/dev/null))
    
    if [[ $? -eq 0 ]]; then
        _describe 'completions' completions
    fi
}

_%s "$@"`, programName, programName, programName, programName)
}

// GenerateFishCompletion generates fish completion script
func (g *Generator) GenerateFishCompletion(programName string) string {
	return fmt.Sprintf(`# Fish completion for %s
function __%s_complete
    set -l tokens (commandline -opc)
    set -l current (commandline -ct)
    
    # Get completions from the program
    %s __complete $tokens 2>/dev/null
end

complete -c %s -f -a "(__%s_complete)"`, 
		programName, programName, programName, programName, programName)
}

// CompletionHandler handles the __complete command
type CompletionHandler struct {
	generator *Generator
}

// NewCompletionHandler creates a new completion handler
func NewCompletionHandler(registry *core.Registry) *CompletionHandler {
	return &CompletionHandler{
		generator: NewGenerator(registry),
	}
}

// Handle processes completion requests
func (ch *CompletionHandler) Handle(args []string) {
	items, err := ch.generator.Complete(args, len(strings.Join(args, " ")))
	if err != nil {
		return // Silent failure for completions
	}
	
	// Output completions in a simple format
	for _, item := range items {
		fmt.Println(item.Value)
	}
}