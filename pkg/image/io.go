package image

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"  // BMP support
	_ "golang.org/x/image/tiff" // TIFF support
	_ "golang.org/x/image/webp" // WebP decode support
)

// Load reads an image from a file path
func Load(path string) (image.Image, error) {
	return imaging.Open(path, imaging.AutoOrientation(true))
}

// LoadFromReader reads an image from an io.Reader
func LoadFromReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

// SaveOptions configures image saving
type SaveOptions struct {
	Format  Format // Target format (defaults to extension-based detection)
	Quality int    // Quality for lossy formats (1-100, default 85)
}

// Save writes an image to a file
func Save(img image.Image, path string, opts ...SaveOptions) error {
	var opt SaveOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Determine format from extension if not specified
	if opt.Format == "" {
		opt.Format = FormatFromExtension(path)
	}

	// Default quality
	if opt.Quality <= 0 {
		opt.Quality = 85
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create output file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return Encode(f, img, opt)
}

// Encode writes an image to an io.Writer
func Encode(w io.Writer, img image.Image, opts SaveOptions) error {
	format := NormalizeFormat(string(opts.Format))

	switch format {
	case FormatPNG:
		return png.Encode(w, img)

	case FormatJPEG, FormatJPG:
		quality := opts.Quality
		if quality <= 0 {
			quality = 85
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})

	case FormatGIF:
		return gif.Encode(w, img, nil)

	case FormatBMP:
		return imaging.Encode(w, img, imaging.BMP)

	case FormatTIFF:
		return imaging.Encode(w, img, imaging.TIFF)

	case FormatWEBP:
		// WebP encoding not supported in pure Go
		// Fall back to PNG for now
		return fmt.Errorf("webp encoding not supported (decode only); use png or jpeg instead")

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// GenerateOutputPath creates an output path based on input and format
func GenerateOutputPath(inputPath string, format Format, suffix string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	if suffix != "" {
		name = name + "_" + suffix
	}

	newExt := "." + string(format)
	if format == FormatJPEG {
		newExt = ".jpg"
	}

	return filepath.Join(dir, name+newExt)
}
