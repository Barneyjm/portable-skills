// Package archive provides zip/unzip utilities.
package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipOptions configures zip creation.
type ZipOptions struct {
	CompressionLevel int // 0-9, -1 for default
}

// DefaultZipOptions returns sensible defaults.
func DefaultZipOptions() ZipOptions {
	return ZipOptions{
		CompressionLevel: -1, // default compression
	}
}

// Zip creates a zip archive from files/directories.
func Zip(outputPath string, paths []string, opts ZipOptions) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	defer func() { _ = w.Close() }()

	for _, path := range paths {
		err := addToZip(w, path, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func addToZip(w *zip.Writer, path, basePath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	// Calculate the name in the archive
	name := filepath.Base(path)
	if basePath != "" {
		name = filepath.Join(basePath, name)
	}

	if info.IsDir() {
		// Add directory entry
		_, err := w.Create(name + "/")
		if err != nil {
			return fmt.Errorf("failed to create directory entry: %w", err)
		}

		// Recursively add contents
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if err := addToZip(w, entryPath, name); err != nil {
				return err
			}
		}
	} else {
		// Add file
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create header: %w", err)
		}
		header.Name = name
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create entry: %w", err)
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = f.Close() }()

		if _, err := io.Copy(writer, f); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

// UnzipOptions configures extraction.
type UnzipOptions struct {
	Overwrite bool // Overwrite existing files
}

// Unzip extracts a zip archive.
func Unzip(archivePath, outputDir string, opts UnzipOptions) ([]string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	var extracted []string

	for _, f := range r.File {
		destPath := filepath.Join(outputDir, f.Name)

		// Security check: prevent zip slip
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(destPath); err == nil && !opts.Overwrite {
			return nil, fmt.Errorf("file exists: %s (use --overwrite to replace)", destPath)
		}

		if err := extractFile(f, destPath); err != nil {
			return nil, err
		}

		extracted = append(extracted, destPath)
	}

	return extracted, nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open archive entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}

// ListEntry represents an entry in an archive.
type ListEntry struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Compressed int64  `json:"compressed"`
	IsDir      bool   `json:"is_dir"`
	Modified   string `json:"modified"`
}

// List returns the contents of a zip archive.
func List(archivePath string) ([]ListEntry, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	var entries []ListEntry
	for _, f := range r.File {
		entries = append(entries, ListEntry{
			Name:       f.Name,
			Size:       int64(f.UncompressedSize64),
			Compressed: int64(f.CompressedSize64),
			IsDir:      f.FileInfo().IsDir(),
			Modified:   f.Modified.Format("2006-01-02 15:04:05"),
		})
	}

	return entries, nil
}
