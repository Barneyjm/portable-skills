package qr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "test.png")

	err := Generate("https://example.com", output, DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created and has content
	stat, err := os.Stat(output)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if stat.Size() == 0 {
		t.Error("output file is empty")
	}

	// Verify it's a PNG (check magic bytes)
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) < 8 || string(data[:4]) != "\x89PNG" {
		t.Error("output is not a valid PNG")
	}
}

func TestGenerateASCII(t *testing.T) {
	result, err := GenerateASCII("hello", DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ASCII output should contain QR code characters
	if result == "" {
		t.Error("empty ASCII output")
	}

	// Should have multiple lines
	lines := strings.Split(result, "\n")
	if len(lines) < 5 {
		t.Error("ASCII output too short")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		hasError bool
	}{
		{"L", LevelL, false},
		{"l", LevelL, false},
		{"low", LevelL, false},
		{"M", LevelM, false},
		{"Q", LevelQ, false},
		{"H", LevelH, false},
		{"invalid", LevelM, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLevel(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if level != tt.expected {
					t.Errorf("got %v, want %v", level, tt.expected)
				}
			}
		})
	}
}

func TestGenerateWithOptions(t *testing.T) {
	dir := t.TempDir()

	// Test different sizes
	for _, size := range []int{128, 256, 512} {
		output := filepath.Join(dir, "test.png")
		opts := Options{Size: size, Level: LevelM}

		err := Generate("test", output, opts)
		if err != nil {
			t.Fatalf("error with size %d: %v", size, err)
		}

		// Just verify it was created
		if _, err := os.Stat(output); err != nil {
			t.Errorf("output not created for size %d", size)
		}
		_ = os.Remove(output)
	}

	// Test with invert
	output := filepath.Join(dir, "inverted.png")
	opts := Options{Size: 256, Level: LevelM, Invert: true}
	err := Generate("test", output, opts)
	if err != nil {
		t.Fatalf("error with invert: %v", err)
	}
}
