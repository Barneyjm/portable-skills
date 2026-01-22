package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage creates a simple test image
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}
	return img
}

// createTestPNGFile creates a test PNG file and returns its path
func createTestPNGFile(t *testing.T, width, height int) string {
	t.Helper()
	img := createTestImage(width, height)

	tmpFile, err := os.CreateTemp("", "test-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = tmpFile.Close() }()

	if err := png.Encode(tmpFile, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	return tmpFile.Name()
}

func TestResize(t *testing.T) {
	tests := []struct {
		name        string
		srcWidth    int
		srcHeight   int
		opts        ResizeOptions
		wantWidth   int
		wantHeight  int
		expectError bool
	}{
		{
			name:       "resize width only (contain)",
			srcWidth:   100,
			srcHeight:  50,
			opts:       ResizeOptions{Width: 50, Height: 0, Fit: "contain"},
			wantWidth:  50,
			wantHeight: 25,
		},
		{
			name:       "resize height only (contain)",
			srcWidth:   100,
			srcHeight:  50,
			opts:       ResizeOptions{Width: 0, Height: 25, Fit: "contain"},
			wantWidth:  50,
			wantHeight: 25,
		},
		{
			name:       "resize both (contain)",
			srcWidth:   100,
			srcHeight:  50,
			opts:       ResizeOptions{Width: 40, Height: 40, Fit: "contain"},
			wantWidth:  40,
			wantHeight: 20,
		},
		{
			name:       "resize cover mode",
			srcWidth:   100,
			srcHeight:  50,
			opts:       ResizeOptions{Width: 30, Height: 30, Fit: "cover"},
			wantWidth:  30,
			wantHeight: 30,
		},
		{
			name:       "resize fill mode",
			srcWidth:   100,
			srcHeight:  50,
			opts:       ResizeOptions{Width: 30, Height: 30, Fit: "fill"},
			wantWidth:  30,
			wantHeight: 30,
		},
		{
			name:        "no dimensions specified",
			srcWidth:    100,
			srcHeight:   50,
			opts:        ResizeOptions{Width: 0, Height: 0},
			expectError: true,
		},
		{
			name:        "cover mode without height",
			srcWidth:    100,
			srcHeight:   50,
			opts:        ResizeOptions{Width: 50, Height: 0, Fit: "cover"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := createTestImage(tt.srcWidth, tt.srcHeight)
			result, err := Resize(src, tt.opts)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			bounds := result.Bounds()
			if bounds.Dx() != tt.wantWidth || bounds.Dy() != tt.wantHeight {
				t.Errorf("Got dimensions %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestResizeFile(t *testing.T) {
	testFile := createTestPNGFile(t, 100, 50)
	defer func() { _ = os.Remove(testFile) }()

	result, err := ResizeFile(testFile, ResizeOptions{Width: 50, Height: 0})
	if err != nil {
		t.Fatalf("ResizeFile failed: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 25 {
		t.Errorf("Got dimensions %dx%d, want 50x25", bounds.Dx(), bounds.Dy())
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"png", FormatPNG},
		{"PNG", FormatPNG},
		{".png", FormatPNG},
		{"jpg", FormatJPEG},
		{"jpeg", FormatJPEG},
		{"JPEG", FormatJPEG},
		{".jpg", FormatJPEG},
		{"gif", FormatGIF},
		{"bmp", FormatBMP},
		{"tiff", FormatTIFF},
		{"webp", FormatWEBP},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeFormat(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	supported := []Format{FormatPNG, FormatJPEG, FormatJPG, FormatGIF, FormatBMP, FormatTIFF, FormatWEBP}
	for _, f := range supported {
		if !IsSupported(f) {
			t.Errorf("Expected %q to be supported", f)
		}
	}

	if IsSupported("invalid") {
		t.Error("Expected 'invalid' to not be supported")
	}
}

func TestFormatFromExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected Format
	}{
		{"image.png", FormatPNG},
		{"photo.jpg", FormatJPEG},
		{"photo.jpeg", FormatJPEG},
		{"animation.gif", FormatGIF},
		{"image.bmp", FormatBMP},
		{"scan.tiff", FormatTIFF},
		{"modern.webp", FormatWEBP},
		{"/path/to/IMAGE.PNG", FormatPNG},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := FormatFromExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("FormatFromExtension(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create test image
	testFile := createTestPNGFile(t, 100, 100)
	defer func() { _ = os.Remove(testFile) }()

	// Load it
	img, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Loaded image has wrong dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Save to different formats
	formats := []struct {
		format Format
		ext    string
	}{
		{FormatPNG, ".png"},
		{FormatJPEG, ".jpg"},
		{FormatGIF, ".gif"},
		{FormatBMP, ".bmp"},
	}

	for _, f := range formats {
		t.Run(string(f.format), func(t *testing.T) {
			outputPath := filepath.Join(os.TempDir(), "test-output"+f.ext)
			defer func() { _ = os.Remove(outputPath) }()

			err := Save(img, outputPath, SaveOptions{Format: f.format, Quality: 85})
			if err != nil {
				t.Fatalf("Save failed for %s: %v", f.format, err)
			}

			// Verify file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created for %s", f.format)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	img := createTestImage(10, 10)

	tests := []struct {
		format      Format
		expectError bool
	}{
		{FormatPNG, false},
		{FormatJPEG, false},
		{FormatGIF, false},
		{FormatBMP, false},
		{FormatTIFF, false},
		{FormatWEBP, true}, // WebP encoding not supported
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, img, SaveOptions{Format: tt.format, Quality: 85})

			if tt.expectError && err == nil {
				t.Errorf("Expected error for format %s, got nil", tt.format)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for format %s: %v", tt.format, err)
			}
			if !tt.expectError && buf.Len() == 0 {
				t.Errorf("Expected non-empty output for format %s", tt.format)
			}
		})
	}
}

func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		input    string
		format   Format
		suffix   string
		expected string
	}{
		{"input.png", FormatJPEG, "", "input.jpg"},
		{"input.png", FormatPNG, "resized", "input_resized.png"},
		{"/path/to/image.bmp", FormatPNG, "", "/path/to/image.png"},
		{"photo.jpg", FormatPNG, "thumb", "photo_thumb.png"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateOutputPath(tt.input, tt.format, tt.suffix)
			if result != tt.expected {
				t.Errorf("GenerateOutputPath(%q, %q, %q) = %q, want %q",
					tt.input, tt.format, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	testFile := createTestPNGFile(t, 100, 50)
	defer func() { _ = os.Remove(testFile) }()

	meta, err := Info(testFile)
	if err != nil {
		t.Fatalf("Info failed: %v", err)
	}

	if meta.Width != 100 {
		t.Errorf("Width = %d, want 100", meta.Width)
	}
	if meta.Height != 50 {
		t.Errorf("Height = %d, want 50", meta.Height)
	}
	if meta.Format != "png" {
		t.Errorf("Format = %q, want %q", meta.Format, "png")
	}
	if meta.SizeBytes <= 0 {
		t.Errorf("SizeBytes = %d, want > 0", meta.SizeBytes)
	}
}

func TestConvert(t *testing.T) {
	img := createTestImage(10, 10)

	// Valid conversion
	result, err := Convert(img, ConvertOptions{Format: FormatJPEG})
	if err != nil {
		t.Errorf("Convert failed: %v", err)
	}
	if result == nil {
		t.Error("Convert returned nil image")
	}

	// Invalid format
	_, err = Convert(img, ConvertOptions{Format: "invalid"})
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestConvertFile(t *testing.T) {
	testFile := createTestPNGFile(t, 50, 50)
	defer func() { _ = os.Remove(testFile) }()

	result, err := ConvertFile(testFile, ConvertOptions{Format: FormatJPEG})
	if err != nil {
		t.Errorf("ConvertFile failed: %v", err)
	}
	if result == nil {
		t.Error("ConvertFile returned nil image")
	}
}
