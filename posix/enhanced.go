package posix

import (
	"fmt"
	"strings"
)

// ParseResult represents the result of parsing command line arguments
type ParseResult struct {
	Flags      map[string]any
	Positional []string
	Remaining  []string
}

// POSIXParser implements a standalone POSIX parser
type POSIXParser struct{}

// NewPOSIXParser creates a new POSIX parser
func NewPOSIXParser() *POSIXParser {
	return &POSIXParser{}
}

// Parse parses command line arguments in POSIX style
func (p *POSIXParser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags:      make(map[string]any),
		Positional: make([]string, 0),
		Remaining:  make([]string, 0),
	}
	
	i := 0
	endOfFlags := false
	
	for i < len(args) {
		arg := args[i]
		
		// Handle -- (end of flags)
		if arg == "--" {
			endOfFlags = true
			i++
			continue
		}
		
		// Handle positional arguments after --
		if endOfFlags || !strings.HasPrefix(arg, "-") {
			result.Positional = append(result.Positional, arg)
			i++
			continue
		}
		
		// Handle single - (stdin/stdout indicator)
		if arg == "-" {
			result.Positional = append(result.Positional, arg)
			i++
			continue
		}
		
		// Handle long flags --flag or --flag=value
		if strings.HasPrefix(arg, "--") {
			if err := p.parseLongFlag(args, &i, result); err != nil {
				return nil, err
			}
			continue
		}
		
		// Handle short flags -f or -abc (bundled)
		if strings.HasPrefix(arg, "-") {
			if err := p.parseShortFlags(args, &i, result); err != nil {
				return nil, err
			}
			continue
		}
		
		i++
	}
	
	return result, nil
}

// parseLongFlag handles --flag and --flag=value forms
func (p *POSIXParser) parseLongFlag(args []string, i *int, result *ParseResult) error {
	arg := args[*i]
	flagName := arg[2:] // Remove --
	
	var value any = true // Default for boolean flags
	hasValue := false
	
	// Check for --flag=value form
	if eqIndex := strings.Index(flagName, "="); eqIndex != -1 {
		value = flagName[eqIndex+1:]
		flagName = flagName[:eqIndex]
		hasValue = true
	}
	
	// If no explicit value and next arg doesn't start with -, use it as value
	if !hasValue && *i+1 < len(args) && !strings.HasPrefix(args[*i+1], "-") {
		// Try to determine if next argument is a value or positional
		// For now, assume it's a value if it exists
		*i++
		value = args[*i]
		hasValue = true
	}
	
	result.Flags[flagName] = value
	*i++
	return nil
}

// parseShortFlags handles -f and -abc (bundled) forms
func (p *POSIXParser) parseShortFlags(args []string, i *int, result *ParseResult) error {
	arg := args[*i]
	flags := arg[1:] // Remove -
	
	// Handle each character as a separate flag
	for j, r := range flags {
		flagName := string(r)
		
		// For bundled flags, all are boolean except possibly the last one
		if j < len(flags)-1 {
			// Middle flags in bundle are always boolean
			result.Flags[flagName] = true
		} else {
			// Last flag might take a value
			var value any = true // Default boolean
			
			// Check if next argument could be a value
			if *i+1 < len(args) && !strings.HasPrefix(args[*i+1], "-") {
				*i++
				value = args[*i]
			}
			
			result.Flags[flagName] = value
		}
	}
	
	*i++
	return nil
}

// FlagInfo represents metadata about a flag
type FlagInfo struct {
	Name        string
	Short       string
	Long        string
	Description string
	Type        string
	Required    bool
	Default     any
	Choices     []string
}

// ParserConfig configures the POSIX parser behavior
type ParserConfig struct {
	KnownFlags    map[string]*FlagInfo
	StrictMode    bool // Fail on unknown flags
	BooleanFlags  map[string]bool
	StringFlags   map[string]bool
	IntegerFlags  map[string]bool
}

// ConfigurableParser provides configurable POSIX parsing
type ConfigurableParser struct {
	config *ParserConfig
}

// NewConfigurableParser creates a parser with configuration
func NewConfigurableParser(config *ParserConfig) *ConfigurableParser {
	if config == nil {
		config = &ParserConfig{
			KnownFlags:   make(map[string]*FlagInfo),
			BooleanFlags: make(map[string]bool),
			StringFlags:  make(map[string]bool),
			IntegerFlags: make(map[string]bool),
		}
	}
	return &ConfigurableParser{config: config}
}

// AddFlag adds a flag definition
func (cp *ConfigurableParser) AddFlag(info *FlagInfo) {
	cp.config.KnownFlags[info.Long] = info
	if info.Short != "" {
		cp.config.KnownFlags[info.Short] = info
	}
	
	switch info.Type {
	case "bool":
		cp.config.BooleanFlags[info.Long] = true
		if info.Short != "" {
			cp.config.BooleanFlags[info.Short] = true
		}
	case "string":
		cp.config.StringFlags[info.Long] = true
		if info.Short != "" {
			cp.config.StringFlags[info.Short] = true
		}
	case "int":
		cp.config.IntegerFlags[info.Long] = true
		if info.Short != "" {
			cp.config.IntegerFlags[info.Short] = true
		}
	}
}

// Parse parses arguments with configuration
func (cp *ConfigurableParser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags:      make(map[string]any),
		Positional: make([]string, 0),
		Remaining:  make([]string, 0),
	}
	
	i := 0
	endOfFlags := false
	
	for i < len(args) {
		arg := args[i]
		
		if arg == "--" {
			endOfFlags = true
			i++
			continue
		}
		
		if endOfFlags || !strings.HasPrefix(arg, "-") {
			result.Positional = append(result.Positional, arg)
			i++
			continue
		}
		
		if arg == "-" {
			result.Positional = append(result.Positional, arg)
			i++
			continue
		}
		
		if strings.HasPrefix(arg, "--") {
			if err := cp.parseConfiguredLongFlag(args, &i, result); err != nil {
				return nil, err
			}
			continue
		}
		
		if strings.HasPrefix(arg, "-") {
			if err := cp.parseConfiguredShortFlags(args, &i, result); err != nil {
				return nil, err
			}
			continue
		}
		
		i++
	}
	
	return result, nil
}

// parseConfiguredLongFlag handles long flags with configuration
func (cp *ConfigurableParser) parseConfiguredLongFlag(args []string, i *int, result *ParseResult) error {
	arg := args[*i]
	flagName := arg[2:]
	
	var value string
	hasValue := false
	
	if eqIndex := strings.Index(flagName, "="); eqIndex != -1 {
		value = flagName[eqIndex+1:]
		flagName = flagName[:eqIndex]
		hasValue = true
	}
	
	// Check if flag is known
	flagInfo, known := cp.config.KnownFlags[flagName]
	if cp.config.StrictMode && !known {
		return fmt.Errorf("unknown flag: --%s", flagName)
	}
	
	// Determine flag type - default to string if unknown
	isBool := false
	if known {
		isBool = cp.config.BooleanFlags[flagName]
		if flagInfo != nil {
			isBool = flagInfo.Type == "bool"
		}
	} else {
		// For unknown flags, check if next argument looks like a value
		if !hasValue && *i+1 < len(args) && !strings.HasPrefix(args[*i+1], "-") {
			// Treat as string flag
			isBool = false
		} else {
			// Treat as boolean flag
			isBool = true
		}
	}
	
	if isBool {
		result.Flags[flagName] = true
		*i++
		return nil
	}
	
	// Non-boolean flag needs a value
	if !hasValue {
		*i++
		if *i >= len(args) {
			return fmt.Errorf("flag --%s requires a value", flagName)
		}
		value = args[*i]
	}
	
	// Convert value based on type
	convertedValue, err := cp.convertValue(value, flagInfo)
	if err != nil {
		return fmt.Errorf("invalid value for flag --%s: %w", flagName, err)
	}
	
	result.Flags[flagName] = convertedValue
	*i++
	return nil
}

// parseConfiguredShortFlags handles short flags with configuration
func (cp *ConfigurableParser) parseConfiguredShortFlags(args []string, i *int, result *ParseResult) error {
	arg := args[*i]
	flags := arg[1:]
	
	for j, r := range flags {
		flagName := string(r)
		
		flagInfo, known := cp.config.KnownFlags[flagName]
		if cp.config.StrictMode && !known {
			return fmt.Errorf("unknown flag: -%s", flagName)
		}
		
		// Determine if boolean - default to true for unknown single-char flags
		isBool := true
		if known {
			isBool = cp.config.BooleanFlags[flagName]
			if flagInfo != nil {
				isBool = flagInfo.Type == "bool"
			}
		} else {
			// For unknown flags, assume boolean unless it's the last flag and has a potential value
			if j == len(flags)-1 && *i+1 < len(args) && !strings.HasPrefix(args[*i+1], "-") {
				isBool = false
			}
		}
		
		if isBool {
			result.Flags[flagName] = true
			continue
		}
		
		// Non-boolean flag must be last in bundle
		if j < len(flags)-1 {
			return fmt.Errorf("non-boolean flag -%s must be last in bundle", flagName)
		}
		
		// Get value
		*i++
		if *i >= len(args) {
			return fmt.Errorf("flag -%s requires a value", flagName)
		}
		
		value := args[*i]
		convertedValue, err := cp.convertValue(value, flagInfo)
		if err != nil {
			return fmt.Errorf("invalid value for flag -%s: %w", flagName, err)
		}
		
		result.Flags[flagName] = convertedValue
	}
	
	*i++
	return nil
}

// convertValue converts string values based on flag configuration
func (cp *ConfigurableParser) convertValue(value string, flagInfo *FlagInfo) (any, error) {
	if flagInfo == nil {
		return value, nil // Default to string
	}
	
	switch flagInfo.Type {
	case "string":
		return value, nil
	case "int":
		return parseInt(value)
	case "bool":
		return parseBool(value)
	default:
		return value, nil
	}
}

// applyDefaults applies default values to unset flags
func (cp *ConfigurableParser) applyDefaults(result *ParseResult) error {
	for _, flagInfo := range cp.config.KnownFlags {
		if flagInfo.Default == nil {
			continue
		}
		
		// Only apply default if flag wasn't set
		if _, exists := result.Flags[flagInfo.Long]; !exists {
			result.Flags[flagInfo.Long] = flagInfo.Default
		}
	}
	return nil
}

// validateRequired validates that required flags are present
func (cp *ConfigurableParser) validateRequired(result *ParseResult) error {
	for _, flagInfo := range cp.config.KnownFlags {
		if !flagInfo.Required {
			continue
		}
		
		if _, exists := result.Flags[flagInfo.Long]; !exists {
			return fmt.Errorf("required flag --%s is missing", flagInfo.Long)
		}
	}
	return nil
}

// Helper functions
func parseInt(s string) (int, error) {
	// Simple integer parsing - could be enhanced
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "1":
		return true, nil
	case "false", "f", "no", "n", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}