package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZipAndUnzip(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	file1 := filepath.Join(dir, "test1.txt")
	file2 := filepath.Join(dir, "test2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create archive
	archivePath := filepath.Join(dir, "test.zip")
	err := Zip(archivePath, []string{file1, file2}, DefaultZipOptions())
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}

	// Extract to new directory
	extractDir := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("failed to create extract dir: %v", err)
	}

	extracted, err := Unzip(archivePath, extractDir, UnzipOptions{})
	if err != nil {
		t.Fatalf("Unzip failed: %v", err)
	}

	if len(extracted) != 2 {
		t.Errorf("expected 2 extracted files, got %d", len(extracted))
	}

	// Verify content
	content, err := os.ReadFile(filepath.Join(extractDir, "test1.txt"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "content1" {
		t.Errorf("content mismatch: got %q", string(content))
	}
}

func TestZipDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create a directory with files
	subDir := filepath.Join(dir, "mydir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Zip the directory
	archivePath := filepath.Join(dir, "dir.zip")
	err := Zip(archivePath, []string{subDir}, DefaultZipOptions())
	if err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// List contents
	entries, err := List(archivePath)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should have directory entry and file
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entries))
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	// Create test file
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create archive
	archivePath := filepath.Join(dir, "test.zip")
	if err := Zip(archivePath, []string{file}, DefaultZipOptions()); err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// List contents
	entries, err := List(archivePath)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", entries[0].Name)
	}

	if entries[0].Size != 11 {
		t.Errorf("expected size 11, got %d", entries[0].Size)
	}
}

func TestUnzipOverwrite(t *testing.T) {
	dir := t.TempDir()

	// Create test file and archive
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	archivePath := filepath.Join(dir, "test.zip")
	if err := Zip(archivePath, []string{file}, DefaultZipOptions()); err != nil {
		t.Fatalf("Zip failed: %v", err)
	}

	// Create existing file in extract location
	extractDir := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("failed to create extract dir: %v", err)
	}
	existingFile := filepath.Join(extractDir, "test.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	// Try to extract without overwrite - should fail
	_, err := Unzip(archivePath, extractDir, UnzipOptions{Overwrite: false})
	if err == nil {
		t.Error("expected error when file exists and overwrite=false")
	}

	// Extract with overwrite
	_, err = Unzip(archivePath, extractDir, UnzipOptions{Overwrite: true})
	if err != nil {
		t.Fatalf("Unzip with overwrite failed: %v", err)
	}

	// Verify content was overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "original" {
		t.Errorf("expected 'original', got %q", string(content))
	}
}
