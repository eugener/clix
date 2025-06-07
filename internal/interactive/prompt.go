package interactive

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/eugener/clix/internal/bind"
)

// Scanner interface for input scanning (allows mocking)
type Scanner interface {
	Scan() bool
	Text() string
	Err() error
}

// Prompter handles interactive prompting for missing values
type Prompter struct {
	scanner  Scanner
	analyzer *bind.Analyzer
}

// NewPrompter creates a new interactive prompter
func NewPrompter() *Prompter {
	return &Prompter{
		scanner:  bufio.NewScanner(os.Stdin),
		analyzer: bind.NewAnalyzer("posix"),
	}
}

// PromptForMissing prompts user for missing required fields
func (p *Prompter) PromptForMissing(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	targetType := targetStruct.Type()

	// Analyze struct to find required fields
	metadata, err := p.analyzer.Analyze(targetType)
	if err != nil {
		return err
	}

	// Check each required field
	for _, fieldInfo := range metadata.Fields {
		if !fieldInfo.Required || fieldInfo.Positional {
			continue
		}

		field := targetStruct.FieldByName(fieldInfo.Name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// Check if field is already set (non-zero value)
		if !field.IsZero() {
			continue
		}

		// Prompt for this field
		value, err := p.promptForField(fieldInfo)
		if err != nil {
			return err
		}

		// Set the value
		if err := p.setFieldValue(field, fieldInfo.Type, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldInfo.Name, err)
		}
	}

	return nil
}

// promptForField prompts for a specific field
func (p *Prompter) promptForField(fieldInfo bind.FieldInfo) (string, error) {
	// Build prompt message
	prompt := p.buildPrompt(fieldInfo)

	for {
		fmt.Print(prompt)

		if !p.scanner.Scan() {
			return "", fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(p.scanner.Text())

		// Allow empty input for optional fields
		if input == "" && !fieldInfo.Required {
			return input, nil
		}

		// Validate input
		if err := p.validateInput(input, fieldInfo); err != nil {
			fmt.Printf("❌ %s Please try again.\n", err.Error())
			continue
		}

		return input, nil
	}
}

// buildPrompt builds a user-friendly prompt message
func (p *Prompter) buildPrompt(fieldInfo bind.FieldInfo) string {
	var parts []string

	// Main prompt
	if fieldInfo.Description != "" {
		parts = append(parts, fieldInfo.Description)
	} else {
		parts = append(parts, fieldInfo.Name)
	}

	// Add type hint
	typeHint := p.getTypeHint(fieldInfo)
	if typeHint != "" {
		parts = append(parts, fmt.Sprintf("(%s)", typeHint))
	}

	// Add choices if available
	if len(fieldInfo.Choices) > 0 {
		choices := strings.Join(fieldInfo.Choices, ", ")
		parts = append(parts, fmt.Sprintf("[%s]", choices))
	}

	// Add default if available
	if fieldInfo.Default != "" {
		parts = append(parts, fmt.Sprintf("(default: %s)", fieldInfo.Default))
	}

	// Add required indicator
	if fieldInfo.Required {
		parts = append(parts, "*")
	}

	prompt := strings.Join(parts, " ") + ": "

	// Add color if supported
	if p.supportsColor() {
		if fieldInfo.Required {
			prompt = "\033[1m" + prompt + "\033[0m" // Bold for required
		} else {
			prompt = "\033[2m" + prompt + "\033[0m" // Dim for optional
		}
	}

	return prompt
}

// getTypeHint returns a user-friendly type hint
func (p *Prompter) getTypeHint(fieldInfo bind.FieldInfo) string {
	switch fieldInfo.Type.Kind() {
	case reflect.Bool:
		return "true/false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "decimal"
	case reflect.String:
		if len(fieldInfo.Choices) > 0 {
			return "choice"
		}
		return "text"
	case reflect.Slice:
		return "comma-separated"
	default:
		return ""
	}
}

// validateInput validates user input against field constraints
func (p *Prompter) validateInput(input string, fieldInfo bind.FieldInfo) error {
	// Check if empty when required
	if fieldInfo.Required && input == "" {
		return fmt.Errorf("this field is required")
	}

	// Check choices
	if len(fieldInfo.Choices) > 0 && input != "" {
		valid := false
		for _, choice := range fieldInfo.Choices {
			if strings.EqualFold(input, choice) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("must be one of: %s", strings.Join(fieldInfo.Choices, ", "))
		}
	}

	// Type-specific validation
	switch fieldInfo.Type.Kind() {
	case reflect.Bool:
		if input != "" {
			_, err := p.parseBool(input)
			if err != nil {
				return fmt.Errorf("must be true or false")
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if input != "" {
			_, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				return fmt.Errorf("must be a valid number")
			}
		}
	case reflect.Float32, reflect.Float64:
		if input != "" {
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return fmt.Errorf("must be a valid decimal number")
			}
		}
	}

	return nil
}

// setFieldValue sets a field value with type conversion
func (p *Prompter) setFieldValue(field reflect.Value, fieldType reflect.Type, value string) error {
	if value == "" {
		// Keep zero value
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		boolVal, err := p.parseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)

	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			// Split comma-separated values
			values := strings.Split(value, ",")
			slice := reflect.MakeSlice(fieldType, len(values), len(values))
			for i, v := range values {
				slice.Index(i).SetString(strings.TrimSpace(v))
			}
			field.Set(slice)
		}

	default:
		return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}

	return nil
}

// parseBool parses boolean values flexibly
func (p *Prompter) parseBool(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "true", "t", "yes", "y", "1", "on":
		return true, nil
	case "false", "f", "no", "n", "0", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// supportsColor checks if the terminal supports color output
func (p *Prompter) supportsColor() bool {
	// Simple check - in a real implementation you'd check TERM, COLORTERM, etc.
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

// ConfirmPrompter provides confirmation prompts
type ConfirmPrompter struct {
	scanner Scanner
}

// NewConfirmPrompter creates a new confirmation prompter
func NewConfirmPrompter() *ConfirmPrompter {
	return &ConfirmPrompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Confirm prompts for a yes/no confirmation
func (cp *ConfirmPrompter) Confirm(message string, defaultValue bool) (bool, error) {
	defaultText := "y/N"
	if defaultValue {
		defaultText = "Y/n"
	}

	prompt := fmt.Sprintf("%s [%s]: ", message, defaultText)

	for {
		fmt.Print(prompt)

		if !cp.scanner.Scan() {
			return false, fmt.Errorf("failed to read input")
		}

		input := strings.ToLower(strings.TrimSpace(cp.scanner.Text()))

		switch input {
		case "":
			return defaultValue, nil
		case "y", "yes", "true", "1":
			return true, nil
		case "n", "no", "false", "0":
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// SelectPrompter provides selection prompts
type SelectPrompter struct {
	scanner Scanner
}

// NewSelectPrompter creates a new selection prompter
func NewSelectPrompter() *SelectPrompter {
	return &SelectPrompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Select prompts user to select from a list of options
func (sp *SelectPrompter) Select(message string, options []string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}

	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	for {
		fmt.Print("Select option (1-", len(options), "): ")

		if !sp.scanner.Scan() {
			return "", fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(sp.scanner.Text())

		// Try to parse as number
		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= len(options) {
				return options[num-1], nil
			}
		}

		// Try to match by name
		for _, option := range options {
			if strings.EqualFold(input, option) {
				return option, nil
			}
		}

		fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(options))
	}
}

// MultilinePrompter handles multiline input
type MultilinePrompter struct {
	scanner Scanner
}

// NewMultilinePrompter creates a new multiline prompter
func NewMultilinePrompter() *MultilinePrompter {
	return &MultilinePrompter{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// PromptMultiline prompts for multiline input
func (mp *MultilinePrompter) PromptMultiline(message string, endMarker string) (string, error) {
	if endMarker == "" {
		endMarker = "EOF"
	}

	fmt.Printf("%s (end with '%s' on a new line):\n", message, endMarker)

	var lines []string

	for mp.scanner.Scan() {
		line := mp.scanner.Text()
		if line == endMarker {
			break
		}
		lines = append(lines, line)
	}

	if err := mp.scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// SmartPrompter combines all prompter types for intelligent prompting
type SmartPrompter struct {
	basic     *Prompter
	confirm   *ConfirmPrompter
	select_   *SelectPrompter
	multiline *MultilinePrompter
}

// NewSmartPrompter creates a new smart prompter
func NewSmartPrompter() *SmartPrompter {
	return &SmartPrompter{
		basic:     NewPrompter(),
		confirm:   NewConfirmPrompter(),
		select_:   NewSelectPrompter(),
		multiline: NewMultilinePrompter(),
	}
}

// PromptMissing prompts for missing fields using the most appropriate method
func (sp *SmartPrompter) PromptMissing(target any) error {
	return sp.basic.PromptForMissing(target)
}

// Confirm prompts for confirmation
func (sp *SmartPrompter) Confirm(message string, defaultValue bool) (bool, error) {
	return sp.confirm.Confirm(message, defaultValue)
}

// Select prompts for selection
func (sp *SmartPrompter) Select(message string, options []string) (string, error) {
	return sp.select_.Select(message, options)
}

// PromptMultiline prompts for multiline input
func (sp *SmartPrompter) PromptMultiline(message string, endMarker string) (string, error) {
	return sp.multiline.PromptMultiline(message, endMarker)
}
