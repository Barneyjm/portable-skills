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

// Enhanced feature tests

func TestCreateEnhancedWithTheme(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "themed.pptx")
	slides := []EnhancedSlide{
		{Title: "Welcome", Body: "Introduction to our topic"},
		{Title: "Main Content", Body: "Key points here"},
	}

	opts := EnhancedCreateOptions{
		Title: "Themed Presentation",
		Theme: "teal",
	}

	err = CreateEnhanced(pptxPath, slides, opts)
	if err != nil {
		t.Fatalf("CreateEnhanced failed: %v", err)
	}

	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 2 {
		t.Errorf("expected 2 slides, got %d", info.SlideCount)
	}
}

func TestCreateEnhancedWithBackground(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "background.pptx")
	slides := []EnhancedSlide{
		{
			Title:      "Dark Slide",
			Body:       "With custom background",
			Background: &Background{Color: "1E1E1E"},
		},
		{
			Title: "Default Slide",
			Body:  "Uses global background",
		},
	}

	opts := EnhancedCreateOptions{
		Title:      "Background Test",
		Background: &Background{Color: "023047"},
	}

	err = CreateEnhanced(pptxPath, slides, opts)
	if err != nil {
		t.Fatalf("CreateEnhanced failed: %v", err)
	}

	// Verify file is valid
	err = ValidatePPTX(pptxPath)
	if err != nil {
		t.Errorf("created file is not valid: %v", err)
	}
}

func TestCreateEnhancedWithElements(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "elements.pptx")
	slides := []EnhancedSlide{
		{
			Elements: []Element{
				{
					Type: "text",
					Text: &TextElement{
						Content:  "Custom positioned title",
						Position: Position{X: 0.5, Y: 1.0, Width: 9.0, Height: 1.5},
						Style: TextStyle{
							FontSize: 48,
							Bold:     true,
							Color:    "FFFFFF",
							Align:    "center",
						},
					},
				},
				{
					Type: "shape",
					Shape: &ShapeElement{
						Type:      "rectangle",
						Position:  Position{X: 0.25, Y: 0.25, Width: 9.5, Height: 0.1},
						FillColor: "FFB703",
					},
				},
			},
			Background: &Background{Color: "023047"},
		},
	}

	opts := EnhancedCreateOptions{
		Title: "Elements Test",
		Theme: "teal",
	}

	err = CreateEnhanced(pptxPath, slides, opts)
	if err != nil {
		t.Fatalf("CreateEnhanced failed: %v", err)
	}

	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 1 {
		t.Errorf("expected 1 slide, got %d", info.SlideCount)
	}
}

func TestCreateFromJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "json.pptx")
	jsonInput := `{
		"options": {
			"title": "JSON Created",
			"theme": "dark"
		},
		"slides": [
			{
				"title": "From JSON",
				"body": "This slide was created from JSON input"
			},
			{
				"elements": [
					{
						"type": "text",
						"text": {
							"content": "Centered text",
							"position": {"x": 1, "y": 3, "width": 8, "height": 1},
							"style": {"fontSize": 36, "align": "center"}
						}
					}
				]
			}
		]
	}`

	err = CreateFromJSON(pptxPath, []byte(jsonInput))
	if err != nil {
		t.Fatalf("CreateFromJSON failed: %v", err)
	}

	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 2 {
		t.Errorf("expected 2 slides, got %d", info.SlideCount)
	}

	if info.Title != "JSON Created" {
		t.Errorf("expected title 'JSON Created', got '%s'", info.Title)
	}
}

func TestParseJSONInput(t *testing.T) {
	// Valid JSON
	valid := `{"slides": [{"title": "Test"}]}`
	input, err := ParseJSONInput([]byte(valid))
	if err != nil {
		t.Fatalf("ParseJSONInput failed: %v", err)
	}
	if len(input.Slides) != 1 {
		t.Errorf("expected 1 slide, got %d", len(input.Slides))
	}

	// Invalid JSON
	invalid := `{not valid json}`
	_, err = ParseJSONInput([]byte(invalid))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetThemePresets(t *testing.T) {
	presets := GetThemePresets()

	// Should have at least the core presets
	expected := []string{"default", "dark", "light", "teal", "coral"}
	for _, name := range expected {
		found := false
		for _, p := range presets {
			if p == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected preset '%s' not found in %v", name, presets)
		}
	}
}

func TestPositionToEMU(t *testing.T) {
	pos := Position{X: 1.0, Y: 2.0, Width: 3.0, Height: 4.0}
	x, y, cx, cy := pos.ToEMU()

	expectedX := int64(914400)  // 1 inch
	expectedY := int64(1828800) // 2 inches
	expectedCX := int64(2743200) // 3 inches
	expectedCY := int64(3657600) // 4 inches

	if x != expectedX {
		t.Errorf("expected x=%d, got %d", expectedX, x)
	}
	if y != expectedY {
		t.Errorf("expected y=%d, got %d", expectedY, y)
	}
	if cx != expectedCX {
		t.Errorf("expected cx=%d, got %d", expectedCX, cx)
	}
	if cy != expectedCY {
		t.Errorf("expected cy=%d, got %d", expectedCY, cy)
	}
}

func TestCreateFromTextEnhanced(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "enhanced.pptx")
	text := `Introduction
Welcome to our presentation.

Main Content
Key points to discuss.

Conclusion
Thank you!`

	opts := EnhancedCreateOptions{
		Title:      "Enhanced Text",
		Theme:      "coral",
		Background: &Background{Color: "FFF5F5"},
	}

	err = CreateFromTextEnhanced(pptxPath, text, opts)
	if err != nil {
		t.Fatalf("CreateFromTextEnhanced failed: %v", err)
	}

	info, err := GetInfo(pptxPath)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.SlideCount != 3 {
		t.Errorf("expected 3 slides, got %d", info.SlideCount)
	}
}

func TestShapeTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pptx-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pptxPath := filepath.Join(tmpDir, "shapes.pptx")
	slides := []EnhancedSlide{
		{
			Elements: []Element{
				{Type: "shape", Shape: &ShapeElement{Type: "rectangle", Position: Position{X: 1, Y: 1, Width: 2, Height: 1}, FillColor: "FF0000"}},
				{Type: "shape", Shape: &ShapeElement{Type: "oval", Position: Position{X: 4, Y: 1, Width: 2, Height: 1}, FillColor: "00FF00"}},
				{Type: "shape", Shape: &ShapeElement{Type: "line", Position: Position{X: 1, Y: 3, Width: 6, Height: 0}, LineColor: "0000FF", LineWidth: 2}},
				{Type: "shape", Shape: &ShapeElement{Type: "roundRect", Position: Position{X: 1, Y: 4, Width: 2, Height: 1}, FillColor: "FFFF00"}},
				{Type: "shape", Shape: &ShapeElement{Type: "triangle", Position: Position{X: 4, Y: 4, Width: 2, Height: 2}, FillColor: "FF00FF"}},
			},
		},
	}

	err = CreateEnhanced(pptxPath, slides, EnhancedCreateOptions{Title: "Shapes"})
	if err != nil {
		t.Fatalf("CreateEnhanced with shapes failed: %v", err)
	}

	err = ValidatePPTX(pptxPath)
	if err != nil {
		t.Errorf("shapes presentation is invalid: %v", err)
	}
}
