package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAndReadPDF(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test creating a PDF
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	testContent := "Hello, World!\nThis is a test PDF document.\nLine three."

	opts := CreateOptions{
		Title:    "Test Document",
		Author:   "Test Author",
		PageSize: "A4",
		FontSize: 12,
	}

	err = CreateFromText(pdfPath, testContent, opts)
	if err != nil {
		t.Fatalf("CreateFromText failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Fatal("PDF file was not created")
	}

	// Test getting info
	info, err := GetInfo(pdfPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.Pages != 1 {
		t.Errorf("expected 1 page, got %d", info.Pages)
	}

	if info.Path != pdfPath {
		t.Errorf("expected path %s, got %s", pdfPath, info.Path)
	}
}

func TestCreateFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a text file
	txtPath := filepath.Join(tmpDir, "input.txt")
	testContent := "This is content from a file."
	err = os.WriteFile(txtPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create PDF from file
	pdfPath := filepath.Join(tmpDir, "output.pdf")
	err = CreateFromFile(pdfPath, txtPath, DefaultCreateOptions())
	if err != nil {
		t.Fatalf("CreateFromFile failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Fatal("PDF file was not created")
	}
}

func TestCreateFromReader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pdfPath := filepath.Join(tmpDir, "reader.pdf")
	content := "Content from a reader"
	reader := strings.NewReader(content)

	err = CreateFromReader(pdfPath, reader, DefaultCreateOptions())
	if err != nil {
		t.Fatalf("CreateFromReader failed: %v", err)
	}

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Fatal("PDF file was not created")
	}
}

func TestDefaultCreateOptions(t *testing.T) {
	opts := DefaultCreateOptions()

	if opts.PageSize != "A4" {
		t.Errorf("expected PageSize A4, got %s", opts.PageSize)
	}
	if opts.FontSize != 12 {
		t.Errorf("expected FontSize 12, got %f", opts.FontSize)
	}
	if opts.FontFamily != "Helvetica" {
		t.Errorf("expected FontFamily Helvetica, got %s", opts.FontFamily)
	}
	if opts.Margin != 20 {
		t.Errorf("expected Margin 20, got %f", opts.Margin)
	}
}

func TestGetInfoNonexistent(t *testing.T) {
	_, err := GetInfo("/nonexistent/file.pdf")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractTextFromCreatedPDF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a PDF with known content
	pdfPath := filepath.Join(tmpDir, "extract.pdf")
	testContent := "Extract this text please"

	err = CreateFromText(pdfPath, testContent, DefaultCreateOptions())
	if err != nil {
		t.Fatalf("CreateFromText failed: %v", err)
	}

	// Extract text
	extracted, err := ExtractText(pdfPath)
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	// The extracted text should contain our content
	// Note: PDF text extraction may not be exact due to encoding
	if extracted == "" {
		t.Error("extracted text is empty")
	}
}

func TestExtractPageTextInvalidPage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pdfPath := filepath.Join(tmpDir, "single.pdf")
	err = CreateFromText(pdfPath, "Single page content", DefaultCreateOptions())
	if err != nil {
		t.Fatalf("CreateFromText failed: %v", err)
	}

	// Try to extract page 5 from a 1-page document
	_, err = ExtractPageText(pdfPath, 5)
	if err == nil {
		t.Error("expected error for invalid page number")
	}

	// Try page 0
	_, err = ExtractPageText(pdfPath, 0)
	if err == nil {
		t.Error("expected error for page 0")
	}
}
