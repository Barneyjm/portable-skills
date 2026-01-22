package image

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"
)

// Format represents a supported image format
type Format string

const (
	FormatPNG  Format = "png"
	FormatJPEG Format = "jpeg"
	FormatJPG  Format = "jpg"
	FormatGIF  Format = "gif"
	FormatBMP  Format = "bmp"
	FormatTIFF Format = "tiff"
	FormatWEBP Format = "webp"
)

// SupportedFormats lists all supported image formats
var SupportedFormats = []Format{
	FormatPNG,
	FormatJPEG,
	FormatJPG,
	FormatGIF,
	FormatBMP,
	FormatTIFF,
	FormatWEBP,
}

// ConvertOptions configures image conversion
type ConvertOptions struct {
	Format  Format // Target format
	Quality int    // Output quality (1-100, for JPEG)
}

// NormalizeFormat normalizes format strings (e.g., "jpg" -> "jpeg")
func NormalizeFormat(format string) Format {
	f := Format(strings.ToLower(strings.TrimPrefix(format, ".")))
	if f == FormatJPG {
		return FormatJPEG
	}
	return f
}

// IsSupported checks if a format is supported
func IsSupported(format Format) bool {
	normalized := NormalizeFormat(string(format))
	for _, f := range SupportedFormats {
		if normalized == f || (normalized == FormatJPG && f == FormatJPEG) {
			return true
		}
	}
	return false
}

// FormatFromExtension extracts format from filename
func FormatFromExtension(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))
	return NormalizeFormat(ext)
}

// Convert converts an image to a different format
// The actual format conversion happens during Save()
func Convert(src image.Image, opts ConvertOptions) (image.Image, error) {
	if !IsSupported(opts.Format) {
		return nil, fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// The image itself doesn't change - format conversion happens on save
	// This function exists for API consistency and validation
	return src, nil
}

// ConvertFile loads an image and prepares it for conversion
func ConvertFile(inputPath string, opts ConvertOptions) (image.Image, error) {
	src, err := Load(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	return Convert(src, opts)
}
