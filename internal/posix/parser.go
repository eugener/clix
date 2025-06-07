package posix

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Parser implements POSIX-compliant argument parsing
type Parser struct{}

// NewParser creates a new POSIX parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses command line arguments into the target struct
func (p *Parser) Parse(args []string, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	targetType := targetStruct.Type()

	// Build field map from struct tags
	fieldMap, positionalFields := p.buildFieldMap(targetType)

	// Parse arguments
	result, err := p.parseArgs(args, fieldMap)
	if err != nil {
		return err
	}

	// Set flag values
	if err := p.setFieldValues(targetStruct, targetType, result.Flags, fieldMap); err != nil {
		return err
	}

	// Set positional arguments
	if err := p.setPositionalValues(targetStruct, targetType, result.Positional, positionalFields); err != nil {
		return err
	}

	return nil
}

// ValidateStruct validates struct field constraints
func (p *Parser) ValidateStruct(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := targetType.Field(i)

		tag := fieldType.Tag.Get("posix")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		if len(parts) >= 4 && strings.Contains(parts[3], "required") {
			if p.isZeroValue(field) {
				return fmt.Errorf("required field %s is missing", fieldType.Name)
			}
		}
	}

	return nil
}

// parseArgs implements the POSIX argument parsing state machine
func (p *Parser) parseArgs(args []string, fieldMap map[string]fieldInfo) (*parseResult, error) {
	result := &parseResult{
		Flags:      make(map[string]any),
		Positional: []string{},
		Remaining:  []string{},
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
			if err := p.parseLongFlag(args, &i, result, fieldMap); err != nil {
				return nil, err
			}
			continue
		}

		// Handle short flags -f or -abc (bundled)
		if strings.HasPrefix(arg, "-") {
			if err := p.parseShortFlag(args, &i, result, fieldMap); err != nil {
				return nil, err
			}
			continue
		}

		i++
	}

	return result, nil
}

// parseLongFlag handles --flag and --flag=value forms
func (p *Parser) parseLongFlag(args []string, i *int, result *parseResult, fieldMap map[string]fieldInfo) error {
	arg := args[*i]
	flagName := arg[2:] // Remove --

	var value string
	hasValue := false

	// Check for --flag=value form
	if eqIndex := strings.Index(flagName, "="); eqIndex != -1 {
		value = flagName[eqIndex+1:]
		flagName = flagName[:eqIndex]
		hasValue = true
	}

	field, exists := fieldMap[flagName]
	if !exists {
		return fmt.Errorf("unknown flag: --%s", flagName)
	}

	// Handle boolean flags
	if field.Type.Kind() == reflect.Bool {
		result.Flags[flagName] = true
		*i++
		return nil
	}

	// Get value for non-boolean flags
	if !hasValue {
		*i++
		if *i >= len(args) {
			return fmt.Errorf("flag --%s requires a value", flagName)
		}
		value = args[*i]
	}

	parsedValue, err := p.parseValue(value, field.Type)
	if err != nil {
		return fmt.Errorf("invalid value for flag --%s: %w", flagName, err)
	}

	result.Flags[flagName] = parsedValue
	*i++
	return nil
}

// parseShortFlag handles -f and -abc (bundled) forms
func (p *Parser) parseShortFlag(args []string, i *int, result *parseResult, fieldMap map[string]fieldInfo) error {
	arg := args[*i]
	flags := arg[1:] // Remove -

	for j, r := range flags {
		flagName := string(r)

		// Find field by short name
		var field fieldInfo
		var exists bool
		for _, f := range fieldMap {
			if f.Short == flagName {
				field = f
				exists = true
				break
			}
		}

		if !exists {
			return fmt.Errorf("unknown flag: -%s", flagName)
		}

		// Handle boolean flags
		if field.Type.Kind() == reflect.Bool {
			result.Flags[field.Long] = true
			continue
		}

		// Non-boolean flag must be the last in a bundle
		if j < len(flags)-1 {
			return fmt.Errorf("non-boolean flag -%s must be last in bundle", flagName)
		}

		// Get value
		*i++
		if *i >= len(args) {
			return fmt.Errorf("flag -%s requires a value", flagName)
		}

		value := args[*i]
		parsedValue, err := p.parseValue(value, field.Type)
		if err != nil {
			return fmt.Errorf("invalid value for flag -%s: %w", flagName, err)
		}

		result.Flags[field.Long] = parsedValue
	}

	*i++
	return nil
}

// Helper types and functions
type parseResult struct {
	Flags      map[string]any
	Positional []string
	Remaining  []string
}

type fieldInfo struct {
	Short    string
	Long     string
	Type     reflect.Type
	Required bool
	Desc     string
}

func (p *Parser) buildFieldMap(structType reflect.Type) (map[string]fieldInfo, []reflect.StructField) {
	fieldMap := make(map[string]fieldInfo)
	var positionalFields []reflect.StructField

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("posix")

		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		if len(parts) >= 4 && strings.Contains(parts[3], "positional") {
			positionalFields = append(positionalFields, field)
			continue
		}

		info := fieldInfo{
			Type: field.Type,
		}

		if len(parts) > 0 && parts[0] != "" {
			info.Short = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			info.Long = parts[1]
		} else {
			info.Long = strings.ToLower(field.Name)
		}
		if len(parts) > 3 {
			info.Required = strings.Contains(parts[3], "required")
			info.Desc = strings.TrimSpace(parts[2])
		}

		fieldMap[info.Long] = info
	}

	return fieldMap, positionalFields
}

func (p *Parser) parseValue(value string, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType.Kind())
	}
}

func (p *Parser) setFieldValues(targetStruct reflect.Value, targetType reflect.Type, flags map[string]any, fieldMap map[string]fieldInfo) error {
	for flagName, value := range flags {
		field, exists := fieldMap[flagName]
		if !exists {
			continue
		}

		// Find struct field
		var structField reflect.Value
		for i := 0; i < targetStruct.NumField(); i++ {
			if targetType.Field(i).Name == field.Long ||
				strings.ToLower(targetType.Field(i).Name) == field.Long {
				structField = targetStruct.Field(i)
				break
			}
		}

		if !structField.IsValid() || !structField.CanSet() {
			continue
		}

		// Convert and set value
		convertedValue := reflect.ValueOf(value).Convert(field.Type)
		structField.Set(convertedValue)
	}

	return nil
}

func (p *Parser) setPositionalValues(targetStruct reflect.Value, targetType reflect.Type, positional []string, positionalFields []reflect.StructField) error {
	for i, field := range positionalFields {
		if i >= len(positional) {
			break
		}

		structField := targetStruct.FieldByName(field.Name)
		if !structField.IsValid() || !structField.CanSet() {
			continue
		}

		// Handle slice types for remaining positional args
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
			remaining := positional[i:]
			slice := reflect.MakeSlice(field.Type, len(remaining), len(remaining))
			for j, arg := range remaining {
				slice.Index(j).SetString(arg)
			}
			structField.Set(slice)
			break
		}

		// Handle single values
		if field.Type.Kind() == reflect.String {
			structField.SetString(positional[i])
		}
	}

	return nil
}

func (p *Parser) isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
