package skill

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Input handles reading input from file path or stdin
type Input struct {
	Path           string
	Reader         io.Reader
	IsTemp         bool
	DetectedFormat string // Format detected from magic bytes (e.g., "png", "jpeg")
}

// Image format magic bytes
var magicBytes = map[string][]byte{
	"png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"jpeg": {0xFF, 0xD8, 0xFF},
	"gif":  {0x47, 0x49, 0x46, 0x38},
	"bmp":  {0x42, 0x4D},
	"webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header (need to also check for WEBP)
	"tiff": {0x49, 0x49, 0x2A, 0x00}, // Little-endian TIFF
}

var magicBytesTiffBE = []byte{0x4D, 0x4D, 0x00, 0x2A} // Big-endian TIFF

// detectFormat detects image format from the first bytes of data
func detectFormat(header []byte) string {
	for format, magic := range magicBytes {
		if len(header) >= len(magic) && bytes.HasPrefix(header, magic) {
			// Special handling for WEBP - need to verify WEBP marker at offset 8
			if format == "webp" {
				if len(header) >= 12 && string(header[8:12]) == "WEBP" {
					return "webp"
				}
				continue
			}
			return format
		}
	}
	// Check big-endian TIFF
	if len(header) >= len(magicBytesTiffBE) && bytes.HasPrefix(header, magicBytesTiffBE) {
		return "tiff"
	}
	return ""
}

// formatToExtension maps detected format to file extension
func formatToExtension(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	case "bmp":
		return ".bmp"
	case "webp":
		return ".webp"
	case "tiff":
		return ".tiff"
	default:
		return ".img" // Generic extension for unknown formats
	}
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
// with auto-detected format extension based on magic bytes
func NewInputFromStdin() (*Input, error) {
	// Read all stdin into memory first to detect format
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("stdin is empty")
	}

	// Detect format from magic bytes
	detectedFormat := detectFormat(data)
	extension := formatToExtension(detectedFormat)

	// Create temp file with appropriate extension
	tmpFile, err := os.CreateTemp("", "ps-input-*"+extension)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	return &Input{
		Path:           tmpFile.Name(),
		IsTemp:         true,
		DetectedFormat: detectedFormat,
	}, nil
}

// NewInputFromStdinWithExt creates an Input that saves stdin to a temp file
// with a specified extension (for backwards compatibility or when format is known)
func NewInputFromStdinWithExt(extension string) (*Input, error) {
	tmpFile, err := os.CreateTemp("", "ps-input-*"+extension)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
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
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	i.Path = tmpFile.Name()
	i.IsTemp = true
	return i.Path, nil
}
