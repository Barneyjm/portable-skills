package pptx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAndReadPPTX(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a presentation
	pptxPath := filepath.Join(tmpDir, "test.pptx")
	slides := []SlideContent{
		{Title: "Introduction", Body: "Welcome to the presentation"},
		{Title: "Main Points", Body: "Point 1\nPoint 2\nPoint 3"},
		{Title: "Conclusion", Body: "Thank you!"},
	}

	opts := CreateOptions{
		Title:   "Test Presentation",
		Creator: "Test Author",
	}

	err = Create(pptxPath, slides, opts)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(pptxPath); os.IsNotExist(err) {
		t.Fatal("PPTX file was not created")
	}

	// Test GetInfo
	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 3 {
		t.Errorf("expected 3 slides, got %d", info.SlideCount)
	}

	if info.Title != "Test Presentation" {
		t.Errorf("expected title 'Test Presentation', got '%s'", info.Title)
	}

	if info.Creator != "Test Author" {
		t.Errorf("expected creator 'Test Author', got '%s'", info.Creator)
	}
}

func TestExtractText(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "text.pptx")
	slides := []SlideContent{
		{Title: "Slide One", Body: "Content for slide one"},
		{Title: "Slide Two", Body: "Content for slide two"},
	}

	err = Create(pptxPath, slides, CreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Extract all text
	text, err := ExtractText(pptxPath)
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if !strings.Contains(text, "Slide One") {
		t.Error("extracted text should contain 'Slide One'")
	}

	if !strings.Contains(text, "Content for slide one") {
		t.Error("extracted text should contain body content")
	}
}

func TestExtractSlideText(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "slide.pptx")
	slides := []SlideContent{
		{Title: "First", Body: "First content"},
		{Title: "Second", Body: "Second content"},
	}

	err = Create(pptxPath, slides, CreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Extract specific slide
	text, err := ExtractSlideText(pptxPath, 2)
	if err != nil {
		t.Fatalf("ExtractSlideText failed: %v", err)
	}

	if !strings.Contains(text, "Second") {
		t.Errorf("slide 2 should contain 'Second', got: %s", text)
	}

	// Test invalid slide number
	_, err = ExtractSlideText(pptxPath, 99)
	if err == nil {
		t.Error("expected error for invalid slide number")
	}
}

func TestListSlides(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "list.pptx")
	slides := []SlideContent{
		{Title: "Alpha", Body: "A"},
		{Title: "Beta", Body: "B"},
		{Title: "Gamma", Body: "C"},
	}

	err = Create(pptxPath, slides, CreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	slideList, err := ListSlides(pptxPath)
	if err != nil {
		t.Fatalf("ListSlides failed: %v", err)
	}

	if len(slideList) != 3 {
		t.Errorf("expected 3 slides, got %d", len(slideList))
	}

	// Check order
	if slideList[0].Number != 1 || slideList[0].Title != "Alpha" {
		t.Errorf("slide 1 incorrect: %+v", slideList[0])
	}
}

func TestCreateFromText(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "fromtext.pptx")
	text := `First Slide Title
This is the body of the first slide.

Second Slide Title
Body of slide two.
More content here.

Third Slide
Final content.`

	err = CreateFromText(pptxPath, text, CreateOptions{Title: "From Text"})
	if err != nil {
		t.Fatalf("CreateFromText failed: %v", err)
	}

	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 3 {
		t.Errorf("expected 3 slides, got %d", info.SlideCount)
	}
}

func TestValidatePPTX(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create valid PPTX
	pptxPath := filepath.Join(tmpDir, "valid.pptx")
	err = Create(pptxPath, []SlideContent{{Title: "Test"}}, CreateOptions{})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = ValidatePPTX(pptxPath)
	if err != nil {
		t.Errorf("ValidatePPTX should pass for valid file: %v", err)
	}

	// Test invalid file
	invalidPath := filepath.Join(tmpDir, "invalid.txt")
	if err := os.WriteFile(invalidPath, []byte("not a pptx"), 0644); err != nil {
		t.Fatal(err)
	}

	err = ValidatePPTX(invalidPath)
	if err == nil {
		t.Error("ValidatePPTX should fail for invalid file")
	}
}

func TestIsPPTX(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"presentation.pptx", true},
		{"DECK.PPTX", true},
		{"document.docx", false},
		{"file.txt", false},
		{"no-extension", false},
	}

	for _, tt := range tests {
		result := IsPPTX(tt.path)
		if result != tt.expected {
			t.Errorf("IsPPTX(%s) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetInfoNonexistent(t *testing.T) {
	_, err := GetInfo("/nonexistent/file.pptx")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
