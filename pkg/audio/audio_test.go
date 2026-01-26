package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"song.mp3", "MP3"},
		{"song.m4a", "M4A/MP4"},
		{"song.flac", "FLAC"},
		{"song.ogg", "OGG"},
		{"song.wav", "WAV"},
		{"song.unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectFileType(tt.path)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInfoNonExistent(t *testing.T) {
	_, err := Info("/nonexistent/file.mp3")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestInfoInvalidFile(t *testing.T) {
	// Create a non-audio file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.mp3")
	if err := os.WriteFile(path, []byte("not an audio file"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should return basic info even for invalid audio
	meta, err := Info(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.File != "test.mp3" {
		t.Errorf("expected file 'test.mp3', got %q", meta.File)
	}

	if meta.FileType != "MP3" {
		t.Errorf("expected type 'MP3', got %q", meta.FileType)
	}
}

func TestExtractPictureNoArt(t *testing.T) {
	// Create a non-audio file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.mp3")
	if err := os.WriteFile(path, []byte("not an audio file"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	output := filepath.Join(dir, "art.jpg")
	err := ExtractPicture(path, output)
	if err == nil {
		t.Error("expected error when extracting from invalid file")
	}
}
