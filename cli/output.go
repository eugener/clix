package cli

import (
	"io"
	"os"

	"github.com/eugener/clix/internal/output"
)

// Format represents different output formats - re-exported from internal package
type Format = output.Format

// Format constants for public use
const (
	FormatJSON  = output.FormatJSON
	FormatYAML  = output.FormatYAML
	FormatTable = output.FormatTable
	FormatText  = output.FormatText
)

// OutputConfig provides configuration for structured output
type OutputConfig struct {
	Format Format `posix:"f,format,Output format (json|yaml|table|text)"`
}

// NewFormatter creates a new output formatter with the specified format
func NewFormatter(format output.Format) *output.Formatter {
	return output.NewFormatter(os.Stdout, format)
}

// NewFormatterWithWriter creates a new output formatter with a custom writer
func NewFormatterWithWriter(writer io.Writer, format output.Format) *output.Formatter {
	return output.NewFormatter(writer, format)
}

// FormatAndOutput formats data using the specified format and writes to stdout
func FormatAndOutput(data any, format output.Format) error {
	formatter := NewFormatter(format)
	return formatter.Format(data)
}

// FormatToWriter formats data using the specified format and writes to the given writer
func FormatToWriter(writer io.Writer, data any, format output.Format) error {
	formatter := NewFormatterWithWriter(writer, format)
	return formatter.Format(data)
}

// OutputCmd creates a command with built-in output formatting support
func OutputCmd[T any](name, description string, handler func(config T, outputFormat output.Format) (any, error)) any {
	return func(config T) error {
		// Extract output format if the config implements OutputConfig
		outputFormat := output.FormatText
		if oc, ok := any(config).(interface{ GetFormat() output.Format }); ok {
			outputFormat = oc.GetFormat()
		}

		// Execute the handler
		result, err := handler(config, outputFormat)
		if err != nil {
			return err
		}

		// Format and output the result
		return FormatAndOutput(result, outputFormat)
	}
}

// SupportedFormats returns all supported output formats
func SupportedFormats() []output.Format {
	return output.SupportedFormats()
}

// ValidFormat checks if a format string is valid
func ValidFormat(format string) bool {
	return output.ValidFormat(format)
}
