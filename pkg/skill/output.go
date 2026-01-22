package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Output handles writing output to file path or stdout
type Output struct {
	Path   string
	Writer io.Writer
}

// NewOutput creates an Output for a file path or stdout
func NewOutput(pathOrDash string) (*Output, error) {
	if pathOrDash == "-" || pathOrDash == "" {
		return &Output{
			Path:   "",
			Writer: os.Stdout,
		}, nil
	}

	f, err := os.Create(pathOrDash)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return &Output{
		Path:   pathOrDash,
		Writer: f,
	}, nil
}

// Close closes the output if it's a file
func (o *Output) Close() error {
	if closer, ok := o.Writer.(io.Closer); ok && o.Path != "" {
		return closer.Close()
	}
	return nil
}

// WriteJSON writes data as JSON to the output
func WriteJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// WriteJSONCompact writes data as compact JSON to the output
func WriteJSONCompact(v interface{}) error {
	return json.NewEncoder(os.Stdout).Encode(v)
}

// PrintSuccess prints a success message to stderr (not stdout, to not interfere with data output)
func PrintSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintError prints an error message to stderr
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// ExitError prints an error and exits with code 1
func ExitError(format string, args ...interface{}) {
	PrintError(format, args...)
	os.Exit(1)
}

// Result represents a structured result for JSON output
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Output  string      `json:"output,omitempty"`
}

// WriteResult writes a structured result
func WriteResult(success bool, data interface{}, err error, outputPath string) error {
	result := Result{
		Success: success,
		Data:    data,
		Output:  outputPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return WriteJSON(result)
}
