package skill

import (
	"fmt"
	"io"
	"os"
)

// Input handles reading input from file path or stdin
type Input struct {
	Path   string
	Reader io.Reader
	IsTemp bool
}

// NewInput creates an Input from a file path or stdin
func NewInput(pathOrDash string) (*Input, error) {
	if pathOrDash == "-" || pathOrDash == "" {
		// Read from stdin
		return &Input{
			Path:   "",
			Reader: os.Stdin,
		}, nil
	}

	// Check if file exists
	if _, err := os.Stat(pathOrDash); err != nil {
		return nil, fmt.Errorf("input file not found: %s", pathOrDash)
	}

	f, err := os.Open(pathOrDash)
	if err != nil {
		return nil, fmt.Errorf("failed to open input: %w", err)
	}

	return &Input{
		Path:   pathOrDash,
		Reader: f,
	}, nil
}

// NewInputFromStdin creates an Input that saves stdin to a temp file
// This is useful when the underlying library needs a file path
func NewInputFromStdin(extension string) (*Input, error) {
	tmpFile, err := os.CreateTemp("", "ps-input-*"+extension)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	return &Input{
		Path:   tmpFile.Name(),
		IsTemp: true,
	}, nil
}

// Close cleans up any temporary resources
func (i *Input) Close() error {
	if i.IsTemp && i.Path != "" {
		return os.Remove(i.Path)
	}
	if closer, ok := i.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// GetPath returns the file path, creating a temp file from stdin if necessary
func (i *Input) GetPath(extension string) (string, error) {
	if i.Path != "" {
		return i.Path, nil
	}

	// Need to materialize stdin to a temp file
	tmpFile, err := os.CreateTemp("", "ps-input-*"+extension)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, i.Reader); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	i.Path = tmpFile.Name()
	i.IsTemp = true
	return i.Path, nil
}
