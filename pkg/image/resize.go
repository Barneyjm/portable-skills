package image

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// ResizeOptions configures image resizing behavior
type ResizeOptions struct {
	Width   int    // Target width in pixels (0 to auto-calculate from aspect ratio)
	Height  int    // Target height in pixels (0 to auto-calculate from aspect ratio)
	Fit     string // Fit mode: "contain", "cover", "fill"
	Quality int    // Output quality (1-100, for JPEG/WebP)
}

// Resize resizes an image according to the specified options
func Resize(src image.Image, opts ResizeOptions) (image.Image, error) {
	if opts.Width <= 0 && opts.Height <= 0 {
		return nil, fmt.Errorf("at least one of width or height must be specified")
	}

	// Default quality
	if opts.Quality <= 0 {
		opts.Quality = 85
	}

	var resized image.Image
	switch opts.Fit {
	case "cover":
		// Fill the dimensions, cropping excess
		if opts.Width <= 0 || opts.Height <= 0 {
			return nil, fmt.Errorf("cover mode requires both width and height")
		}
		resized = imaging.Fill(src, opts.Width, opts.Height, imaging.Center, imaging.Lanczos)
	case "fill":
		// Stretch to exact dimensions (ignores aspect ratio)
		if opts.Width <= 0 || opts.Height <= 0 {
			return nil, fmt.Errorf("fill mode requires both width and height")
		}
		resized = imaging.Resize(src, opts.Width, opts.Height, imaging.Lanczos)
	default: // "contain" - fit within dimensions, preserving aspect ratio
		if opts.Width > 0 && opts.Height > 0 {
			// Both dimensions specified - use Fit to maintain aspect ratio within bounds
			resized = imaging.Fit(src, opts.Width, opts.Height, imaging.Lanczos)
		} else {
			// Only one dimension - use Resize which auto-calculates the other
			resized = imaging.Resize(src, opts.Width, opts.Height, imaging.Lanczos)
		}
	}

	return resized, nil
}

// ResizeFile loads, resizes, and returns the processed image
func ResizeFile(inputPath string, opts ResizeOptions) (image.Image, error) {
	src, err := Load(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	return Resize(src, opts)
}
