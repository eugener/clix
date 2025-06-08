package config

import (
	"fmt"
	"reflect"

	"github.com/eugener/clix/internal/bind"
)

// MergeConfigs merges base configuration into target configuration
// Values in target (from CLI args) take precedence over base (from config file)
func MergeConfigs(target, base any) error {
	targetValue := reflect.ValueOf(target)
	baseValue := reflect.ValueOf(base)

	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	if baseValue.Kind() == reflect.Ptr {
		baseValue = baseValue.Elem()
	}

	if baseValue.Kind() != reflect.Struct {
		return fmt.Errorf("base must be a struct or pointer to struct")
	}

	targetStruct := targetValue.Elem()
	baseStruct := baseValue

	// Check that both structs have the same type
	if targetStruct.Type() != baseStruct.Type() {
		return fmt.Errorf("target and base configurations must have the same type")
	}

	// Copy non-zero values from base to target where target field is zero
	for i := 0; i < targetStruct.NumField(); i++ {
		targetField := targetStruct.Field(i)
		baseField := baseStruct.Field(i)

		// Skip unexported fields
		if !targetField.CanSet() {
			continue
		}

		// If target field is zero and base field is not zero, copy from base
		if targetField.IsZero() && !baseField.IsZero() {
			if targetField.Type() == baseField.Type() {
				targetField.Set(baseField)
			}
		}
	}

	return nil
}

// ValidateConfig validates the parsed configuration
func ValidateConfig(config any) error {
	// Use the binder's analyzer for validation
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	analyzer := bind.NewAnalyzer("posix")
	metadata, err := analyzer.Analyze(configValue.Type())
	if err != nil {
		return err
	}

	// Check required fields
	for _, fieldInfo := range metadata.Fields {
		if !fieldInfo.Required {
			continue
		}

		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			return fmt.Errorf("required field %s is missing", fieldInfo.Name)
		}
	}

	// Check choices validation
	for _, fieldInfo := range metadata.Fields {
		if len(fieldInfo.Choices) == 0 {
			continue
		}

		field := configValue.FieldByName(fieldInfo.Name)
		if !field.IsValid() || field.IsZero() {
			continue
		}

		value := fmt.Sprintf("%v", field.Interface())
		valid := false
		for _, choice := range fieldInfo.Choices {
			if value == choice {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("field %s must be one of: %v", fieldInfo.Name, fieldInfo.Choices)
		}
	}

	return nil
}