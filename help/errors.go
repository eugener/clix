package help

import (
	"fmt"
	"sort"
	"strings"
)

// ErrorFormatter provides enhanced error messages with suggestions
type ErrorFormatter struct {
	programName string
	useColor    bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(programName string, colorize bool) *ErrorFormatter {
	return &ErrorFormatter{
		programName: programName,
		useColor:    colorize,
	}
}

// FormatError formats an error with context and suggestions
func (ef *ErrorFormatter) FormatError(err error, context *ErrorContext) string {
	if context == nil {
		return ef.formatBasicError(err)
	}

	switch context.Type {
	case ErrorTypeUnknownCommand:
		return ef.formatUnknownCommand(err, context)
	case ErrorTypeUnknownFlag:
		return ef.formatUnknownFlag(err, context)
	case ErrorTypeMissingRequired:
		return ef.formatMissingRequired(err, context)
	case ErrorTypeInvalidValue:
		return ef.formatInvalidValue(err, context)
	case ErrorTypeValidation:
		return ef.formatValidationError(err, context)
	default:
		return ef.formatBasicError(err)
	}
}

// ErrorType represents different types of CLI errors
type ErrorType int

const (
	ErrorTypeGeneric ErrorType = iota
	ErrorTypeUnknownCommand
	ErrorTypeUnknownFlag
	ErrorTypeMissingRequired
	ErrorTypeInvalidValue
	ErrorTypeValidation
)

// ErrorContext provides context for error formatting
type ErrorContext struct {
	Type            ErrorType
	Command         string
	Flag            string
	Value           string
	Suggestions     []string
	AvailableItems  []string
	RequiredFlags   []string
	AllFlags        []string
	AllCommands     []string
	Examples        []string
	HelpCommand     string
}

// formatUnknownCommand formats errors for unknown commands
func (ef *ErrorFormatter) formatUnknownCommand(err error, context *ErrorContext) string {
	var msg strings.Builder
	
	// Main error
	msg.WriteString(ef.colorize(ColorRed, "❌ Unknown command: "))
	msg.WriteString(ef.colorize(ColorBold, fmt.Sprintf("'%s'", context.Command)))
	msg.WriteString("\n\n")
	
	// Suggestions
	if len(context.Suggestions) > 0 {
		msg.WriteString("💡 Did you mean:\n")
		for _, suggestion := range context.Suggestions {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "→"), 
				ef.colorize(ColorCyan, suggestion)))
		}
		msg.WriteString("\n")
	}
	
	// Available commands
	if len(context.AllCommands) > 0 {
		msg.WriteString("📋 Available commands:\n")
		for _, cmd := range context.AllCommands {
			msg.WriteString(fmt.Sprintf("   %s\n", ef.colorize(ColorCyan, cmd)))
		}
		msg.WriteString("\n")
	}
	
	// Help hint
	msg.WriteString(fmt.Sprintf("💬 Use %s for more information\n", 
		ef.colorize(ColorYellow, fmt.Sprintf("'%s help'", ef.programName))))
	
	return msg.String()
}

// formatUnknownFlag formats errors for unknown flags
func (ef *ErrorFormatter) formatUnknownFlag(err error, context *ErrorContext) string {
	var msg strings.Builder
	
	// Main error
	msg.WriteString(ef.colorize(ColorRed, "❌ Unknown flag: "))
	msg.WriteString(ef.colorize(ColorBold, context.Flag))
	msg.WriteString("\n\n")
	
	// Suggestions
	if len(context.Suggestions) > 0 {
		msg.WriteString("💡 Did you mean:\n")
		for _, suggestion := range context.Suggestions {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "→"), 
				ef.colorize(ColorCyan, suggestion)))
		}
		msg.WriteString("\n")
	}
	
	// Available flags
	if len(context.AllFlags) > 0 {
		msg.WriteString("🏳️  Available flags:\n")
		sort.Strings(context.AllFlags)
		for _, flag := range context.AllFlags {
			msg.WriteString(fmt.Sprintf("   %s\n", ef.colorize(ColorCyan, flag)))
		}
		msg.WriteString("\n")
	}
	
	// Help hint
	if context.Command != "" {
		msg.WriteString(fmt.Sprintf("💬 Use %s for detailed usage\n", 
			ef.colorize(ColorYellow, fmt.Sprintf("'%s help %s'", ef.programName, context.Command))))
	}
	
	return msg.String()
}

// formatMissingRequired formats errors for missing required fields
func (ef *ErrorFormatter) formatMissingRequired(err error, context *ErrorContext) string {
	var msg strings.Builder
	
	// Main error
	msg.WriteString(ef.colorize(ColorRed, "❌ Missing required flag: "))
	msg.WriteString(ef.colorize(ColorBold, context.Flag))
	msg.WriteString("\n\n")
	
	// Show what's required
	if len(context.RequiredFlags) > 0 {
		msg.WriteString("🔴 Required flags:\n")
		for _, flag := range context.RequiredFlags {
			status := ef.colorize(ColorRed, "✗")
			if flag == context.Flag {
				status = ef.colorize(ColorRed, "✗ (missing)")
			}
			msg.WriteString(fmt.Sprintf("   %s %s\n", status, ef.colorize(ColorCyan, flag)))
		}
		msg.WriteString("\n")
	}
	
	// Examples
	if len(context.Examples) > 0 {
		msg.WriteString("📝 Examples:\n")
		for _, example := range context.Examples {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "$"), 
				ef.colorize(ColorWhite, example)))
		}
		msg.WriteString("\n")
	}
	
	// Help hint
	if context.Command != "" {
		msg.WriteString(fmt.Sprintf("💬 Use %s for detailed usage\n", 
			ef.colorize(ColorYellow, fmt.Sprintf("'%s help %s'", ef.programName, context.Command))))
	}
	
	return msg.String()
}

// formatInvalidValue formats errors for invalid flag values
func (ef *ErrorFormatter) formatInvalidValue(err error, context *ErrorContext) string {
	var msg strings.Builder
	
	// Main error
	msg.WriteString(ef.colorize(ColorRed, "❌ Invalid value for flag "))
	msg.WriteString(ef.colorize(ColorBold, context.Flag))
	msg.WriteString(": ")
	msg.WriteString(ef.colorize(ColorBold, fmt.Sprintf("'%s'", context.Value)))
	msg.WriteString("\n\n")
	
	// Valid options
	if len(context.AvailableItems) > 0 {
		msg.WriteString("✅ Valid options:\n")
		for _, item := range context.AvailableItems {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "→"), 
				ef.colorize(ColorCyan, item)))
		}
		msg.WriteString("\n")
	}
	
	// Suggestions
	if len(context.Suggestions) > 0 {
		msg.WriteString("💡 Did you mean:\n")
		for _, suggestion := range context.Suggestions {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "→"), 
				ef.colorize(ColorCyan, suggestion)))
		}
		msg.WriteString("\n")
	}
	
	return msg.String()
}

// formatValidationError formats validation errors
func (ef *ErrorFormatter) formatValidationError(err error, context *ErrorContext) string {
	var msg strings.Builder
	
	// Main error
	msg.WriteString(ef.colorize(ColorRed, "❌ Validation failed: "))
	msg.WriteString(err.Error())
	msg.WriteString("\n\n")
	
	// Examples if available
	if len(context.Examples) > 0 {
		msg.WriteString("📝 Examples:\n")
		for _, example := range context.Examples {
			msg.WriteString(fmt.Sprintf("   %s %s\n", 
				ef.colorize(ColorGreen, "$"), 
				ef.colorize(ColorWhite, example)))
		}
		msg.WriteString("\n")
	}
	
	return msg.String()
}

// formatBasicError formats basic errors without context
func (ef *ErrorFormatter) formatBasicError(err error) string {
	return fmt.Sprintf("%s %s\n", 
		ef.colorize(ColorRed, "❌"), 
		err.Error())
}

// Color constants
type Color string

const (
	ColorReset  Color = "\033[0m"
	ColorRed    Color = "\033[31m"
	ColorGreen  Color = "\033[32m"
	ColorYellow Color = "\033[33m"
	ColorBlue   Color = "\033[34m"
	ColorPurple Color = "\033[35m"
	ColorCyan   Color = "\033[36m"
	ColorWhite  Color = "\033[37m"
	ColorBold   Color = "\033[1m"
)

// colorize applies color if colorization is enabled
func (ef *ErrorFormatter) colorize(color Color, text string) string {
	if !ef.useColor {
		return text
	}
	return string(color) + text + string(ColorReset)
}

// SuggestionEngine provides intelligent suggestions for typos
type SuggestionEngine struct{}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() *SuggestionEngine {
	return &SuggestionEngine{}
}

// SuggestCommands suggests similar commands based on input
func (se *SuggestionEngine) SuggestCommands(input string, available []string) []string {
	return se.suggest(input, available, 2) // Allow 2 character difference
}

// SuggestFlags suggests similar flags based on input
func (se *SuggestionEngine) SuggestFlags(input string, available []string) []string {
	return se.suggest(input, available, 2)
}

// SuggestValues suggests similar values based on input
func (se *SuggestionEngine) SuggestValues(input string, available []string) []string {
	return se.suggest(input, available, 1) // Be stricter for values
}

// suggest provides suggestions using Levenshtein distance
func (se *SuggestionEngine) suggest(input string, available []string, maxDistance int) []string {
	var suggestions []string
	
	input = strings.ToLower(input)
	
	for _, item := range available {
		distance := se.levenshteinDistance(input, strings.ToLower(item))
		if distance <= maxDistance && distance > 0 {
			suggestions = append(suggestions, item)
		}
	}
	
	// Sort by distance (closest first)
	sort.Slice(suggestions, func(i, j int) bool {
		distI := se.levenshteinDistance(input, strings.ToLower(suggestions[i]))
		distJ := se.levenshteinDistance(input, strings.ToLower(suggestions[j]))
		return distI < distJ
	})
	
	// Limit to 3 suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	
	return suggestions
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (se *SuggestionEngine) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// ErrorContextBuilder helps build error contexts
type ErrorContextBuilder struct {
	context *ErrorContext
}

// NewErrorContext creates a new error context builder
func NewErrorContext() *ErrorContextBuilder {
	return &ErrorContextBuilder{
		context: &ErrorContext{},
	}
}

// Type sets the error type
func (ecb *ErrorContextBuilder) Type(t ErrorType) *ErrorContextBuilder {
	ecb.context.Type = t
	return ecb
}

// Command sets the command name
func (ecb *ErrorContextBuilder) Command(cmd string) *ErrorContextBuilder {
	ecb.context.Command = cmd
	return ecb
}

// Flag sets the flag name
func (ecb *ErrorContextBuilder) Flag(flag string) *ErrorContextBuilder {
	ecb.context.Flag = flag
	return ecb
}

// Value sets the invalid value
func (ecb *ErrorContextBuilder) Value(value string) *ErrorContextBuilder {
	ecb.context.Value = value
	return ecb
}

// Suggestions sets suggestions
func (ecb *ErrorContextBuilder) Suggestions(suggestions []string) *ErrorContextBuilder {
	ecb.context.Suggestions = suggestions
	return ecb
}

// AvailableItems sets available items
func (ecb *ErrorContextBuilder) AvailableItems(items []string) *ErrorContextBuilder {
	ecb.context.AvailableItems = items
	return ecb
}

// AllCommands sets all available commands
func (ecb *ErrorContextBuilder) AllCommands(commands []string) *ErrorContextBuilder {
	ecb.context.AllCommands = commands
	return ecb
}

// AllFlags sets all available flags
func (ecb *ErrorContextBuilder) AllFlags(flags []string) *ErrorContextBuilder {
	ecb.context.AllFlags = flags
	return ecb
}

// RequiredFlags sets required flags
func (ecb *ErrorContextBuilder) RequiredFlags(flags []string) *ErrorContextBuilder {
	ecb.context.RequiredFlags = flags
	return ecb
}

// Examples sets usage examples
func (ecb *ErrorContextBuilder) Examples(examples []string) *ErrorContextBuilder {
	ecb.context.Examples = examples
	return ecb
}

// Build returns the built context
func (ecb *ErrorContextBuilder) Build() *ErrorContext {
	return ecb.context
}