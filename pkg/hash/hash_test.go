package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		input    string
		algo     Algorithm
		expected string
	}{
		{"hello", MD5, "5d41402abc4b2a76b9719d911017c592"},
		{"hello", SHA1, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"hello", SHA256, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"", MD5, "d41d8cd98f00b204e9800998ecf8427e"},
		{"", SHA256, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		t.Run(string(tt.algo)+":"+tt.input, func(t *testing.T) {
			result, err := String(tt.input, tt.algo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestFile(t *testing.T) {
	// Create a temp file with known content
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := File(path, SHA256)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if result.Hash != expected {
		t.Errorf("got %s, want %s", result.Hash, expected)
	}

	if result.Size != int64(len(content)) {
		t.Errorf("got size %d, want %d", result.Size, len(content))
	}
}

func TestVerify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test matching hash
	match, _, err := Verify(path, SHA256, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected hash to match")
	}

	// Test non-matching hash
	match, _, err = Verify(path, SHA256, "0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Error("expected hash to not match")
	}
}

func TestIsSupported(t *testing.T) {
	if !IsSupported("md5") {
		t.Error("md5 should be supported")
	}
	if !IsSupported("sha256") {
		t.Error("sha256 should be supported")
	}
	if IsSupported("invalid") {
		t.Error("invalid should not be supported")
	}
}
