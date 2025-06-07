package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents different output formats
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
	FormatText  Format = "text"
)

// SupportedFormats returns all supported output formats
func SupportedFormats() []Format {
	return []Format{FormatJSON, FormatYAML, FormatTable, FormatText}
}

// ValidFormat checks if a format string is valid
func ValidFormat(format string) bool {
	for _, f := range SupportedFormats() {
		if string(f) == format {
			return true
		}
	}
	return false
}

// Formatter handles structured output formatting
type Formatter struct {
	writer io.Writer
	format Format
}

// NewFormatter creates a new output formatter
func NewFormatter(writer io.Writer, format Format) *Formatter {
	return &Formatter{
		writer: writer,
		format: format,
	}
}

// Format formats and writes data to the output
func (f *Formatter) Format(data any) error {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	case FormatTable:
		return f.formatTable(data)
	case FormatText:
		return f.formatText(data)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

// formatJSON formats data as JSON
func (f *Formatter) formatJSON(data any) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatYAML formats data as YAML
func (f *Formatter) formatYAML(data any) error {
	encoder := yaml.NewEncoder(f.writer)
	defer func() {
		_ = encoder.Close()
	}()
	return encoder.Encode(data)
}

// formatTable formats data as a table
func (f *Formatter) formatTable(data any) error {
	return f.renderTable(data)
}

// formatText formats data as plain text
func (f *Formatter) formatText(data any) error {
	switch v := data.(type) {
	case string:
		_, err := fmt.Fprintln(f.writer, v)
		return err
	case fmt.Stringer:
		_, err := fmt.Fprintln(f.writer, v.String())
		return err
	default:
		_, err := fmt.Fprintln(f.writer, v)
		return err
	}
}

// renderTable renders data as a formatted table
func (f *Formatter) renderTable(data any) error {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Slice:
		return f.renderSliceAsTable(value)
	case reflect.Struct:
		return f.renderStructAsTable(value)
	case reflect.Map:
		return f.renderMapAsTable(value)
	default:
		// Fallback to simple text format
		return f.formatText(data)
	}
}

// renderSliceAsTable renders a slice as a table
func (f *Formatter) renderSliceAsTable(value reflect.Value) error {
	if value.Len() == 0 {
		_, err := fmt.Fprintln(f.writer, "No data")
		return err
	}

	// Get the first element to determine structure
	firstElem := value.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	switch firstElem.Kind() {
	case reflect.Struct:
		return f.renderStructSliceAsTable(value)
	case reflect.Map:
		return f.renderMapSliceAsTable(value)
	default:
		return f.renderSimpleSliceAsTable(value)
	}
}

// renderStructSliceAsTable renders a slice of structs as a table
func (f *Formatter) renderStructSliceAsTable(value reflect.Value) error {
	if value.Len() == 0 {
		return nil
	}

	// Get headers from the first struct
	firstElem := value.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	headers := f.getStructHeaders(firstElem.Type())
	if len(headers) == 0 {
		return f.formatText(value.Interface())
	}

	// Calculate column widths
	widths := make(map[string]int)
	for _, header := range headers {
		widths[header] = len(header)
	}

	// Collect all rows and calculate max widths
	rows := make([]map[string]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		row := f.structToRow(elem, headers)
		rows[i] = row

		for header, content := range row {
			if len(content) > widths[header] {
				widths[header] = len(content)
			}
		}
	}

	// Render table
	return f.renderTableWithData(headers, rows, widths)
}

// renderMapSliceAsTable renders a slice of maps as a table
func (f *Formatter) renderMapSliceAsTable(value reflect.Value) error {
	if value.Len() == 0 {
		return nil
	}

	// Collect all unique keys
	keySet := make(map[string]bool)
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		for _, key := range elem.MapKeys() {
			keySet[fmt.Sprintf("%v", key.Interface())] = true
		}
	}

	// Sort headers
	headers := make([]string, 0, len(keySet))
	for key := range keySet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	// Calculate column widths
	widths := make(map[string]int)
	for _, header := range headers {
		widths[header] = len(header)
	}

	// Collect rows
	rows := make([]map[string]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		row := make(map[string]string)

		for _, header := range headers {
			key := reflect.ValueOf(header)
			val := elem.MapIndex(key)
			content := ""
			if val.IsValid() {
				content = fmt.Sprintf("%v", val.Interface())
			}
			row[header] = content

			if len(content) > widths[header] {
				widths[header] = len(content)
			}
		}
		rows[i] = row
	}

	return f.renderTableWithData(headers, rows, widths)
}

// renderSimpleSliceAsTable renders a slice of simple values as a table
func (f *Formatter) renderSimpleSliceAsTable(value reflect.Value) error {
	headers := []string{"Value"}
	widths := map[string]int{"Value": 5}

	rows := make([]map[string]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		content := fmt.Sprintf("%v", value.Index(i).Interface())
		rows[i] = map[string]string{"Value": content}
		if len(content) > widths["Value"] {
			widths["Value"] = len(content)
		}
	}

	return f.renderTableWithData(headers, rows, widths)
}

// renderStructAsTable renders a single struct as a table
func (f *Formatter) renderStructAsTable(value reflect.Value) error {
	structType := value.Type()

	headers := []string{"Field", "Value"}
	widths := map[string]int{"Field": 5, "Value": 5}

	var rows []map[string]string
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		fieldValue := fmt.Sprintf("%v", value.Field(i).Interface())

		row := map[string]string{
			"Field": fieldName,
			"Value": fieldValue,
		}
		rows = append(rows, row)

		if len(fieldName) > widths["Field"] {
			widths["Field"] = len(fieldName)
		}
		if len(fieldValue) > widths["Value"] {
			widths["Value"] = len(fieldValue)
		}
	}

	return f.renderTableWithData(headers, rows, widths)
}

// renderMapAsTable renders a map as a table
func (f *Formatter) renderMapAsTable(value reflect.Value) error {
	headers := []string{"Key", "Value"}
	widths := map[string]int{"Key": 3, "Value": 5}

	var rows []map[string]string
	for _, key := range value.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		valueStr := fmt.Sprintf("%v", value.MapIndex(key).Interface())

		row := map[string]string{
			"Key":   keyStr,
			"Value": valueStr,
		}
		rows = append(rows, row)

		if len(keyStr) > widths["Key"] {
			widths["Key"] = len(keyStr)
		}
		if len(valueStr) > widths["Value"] {
			widths["Value"] = len(valueStr)
		}
	}

	return f.renderTableWithData(headers, rows, widths)
}

// getStructHeaders gets field names from a struct type
func (f *Formatter) getStructHeaders(structType reflect.Type) []string {
	var headers []string
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.IsExported() {
			headers = append(headers, field.Name)
		}
	}
	return headers
}

// structToRow converts a struct to a row map
func (f *Formatter) structToRow(value reflect.Value, headers []string) map[string]string {
	row := make(map[string]string)
	structType := value.Type()

	for _, header := range headers {
		field, found := structType.FieldByName(header)
		if found && field.IsExported() {
			fieldValue := value.FieldByName(header)
			row[header] = fmt.Sprintf("%v", fieldValue.Interface())
		} else {
			row[header] = ""
		}
	}

	return row
}

// renderTableWithData renders the actual table with headers, rows, and calculated widths
func (f *Formatter) renderTableWithData(headers []string, rows []map[string]string, widths map[string]int) error {
	// Unicode box-drawing characters for better appearance
	const (
		topLeft     = "┌"
		topRight    = "┐"
		bottomLeft  = "└"
		bottomRight = "┘"
		horizontal  = "─"
		vertical    = "│"
		cross       = "┼"
		topTee      = "┬"
		bottomTee   = "┴"
		leftTee     = "├"
		rightTee    = "┤"
	)

	// Build separators with Unicode box-drawing characters
	// Add 2 to each width to account for padding spaces around content
	var topSeparatorParts []string
	var midSeparatorParts []string
	var bottomSeparatorParts []string

	for _, header := range headers {
		line := strings.Repeat(horizontal, widths[header]+2) // +2 for padding spaces
		topSeparatorParts = append(topSeparatorParts, line)
		midSeparatorParts = append(midSeparatorParts, line)
		bottomSeparatorParts = append(bottomSeparatorParts, line)
	}

	topSeparator := topLeft + strings.Join(topSeparatorParts, topTee) + topRight
	midSeparator := leftTee + strings.Join(midSeparatorParts, cross) + rightTee
	bottomSeparator := bottomLeft + strings.Join(bottomSeparatorParts, bottomTee) + bottomRight

	// Render top border
	if _, err := fmt.Fprintln(f.writer, topSeparator); err != nil {
		return err
	}

	// Render header
	var headerParts []string
	for _, header := range headers {
		headerParts = append(headerParts, fmt.Sprintf("%-*s", widths[header], header))
	}
	if _, err := fmt.Fprintf(f.writer, "%s %s %s\n", vertical, strings.Join(headerParts, " "+vertical+" "), vertical); err != nil {
		return err
	}

	// Render middle separator
	if _, err := fmt.Fprintln(f.writer, midSeparator); err != nil {
		return err
	}

	// Render rows
	for _, row := range rows {
		var rowParts []string
		for _, header := range headers {
			content := row[header]
			rowParts = append(rowParts, fmt.Sprintf("%-*s", widths[header], content))
		}
		if _, err := fmt.Fprintf(f.writer, "%s %s %s\n", vertical, strings.Join(rowParts, " "+vertical+" "), vertical); err != nil {
			return err
		}
	}

	// Render bottom border
	_, err := fmt.Fprintln(f.writer, bottomSeparator)
	return err
}
