// Package pptx provides PowerPoint (.pptx) file reading and creation functionality.
package pptx

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// EMU conversion constants (English Metric Units)
// PowerPoint uses EMUs for positioning: 914400 EMUs = 1 inch
const (
	EMUPerInch = 914400
	EMUPerPt   = 12700 // Points to EMU (1 point = 1/72 inch)
)

// Position represents element positioning in inches
type Position struct {
	X      float64 `json:"x"`      // X position in inches from left
	Y      float64 `json:"y"`      // Y position in inches from top
	Width  float64 `json:"width"`  // Width in inches (0 = auto)
	Height float64 `json:"height"` // Height in inches (0 = auto)
}

// ToEMU converts position to EMU values
func (p Position) ToEMU() (x, y, cx, cy int64) {
	return int64(p.X * EMUPerInch),
		int64(p.Y * EMUPerInch),
		int64(p.Width * EMUPerInch),
		int64(p.Height * EMUPerInch)
}

// TextStyle defines text formatting options
type TextStyle struct {
	FontFace  string `json:"fontFace,omitempty"`  // Font name (e.g., "Arial", "Calibri")
	FontSize  int    `json:"fontSize,omitempty"`  // Font size in points
	Color     string `json:"color,omitempty"`     // Hex color without # (e.g., "FF0000")
	Bold      bool   `json:"bold,omitempty"`      // Bold text
	Italic    bool   `json:"italic,omitempty"`    // Italic text
	Underline bool   `json:"underline,omitempty"` // Underlined text
	Align     string `json:"align,omitempty"`     // "left", "center", "right"
}

// Background defines slide background
type Background struct {
	Color    string `json:"color,omitempty"`    // Solid color (hex without #)
	Gradient *struct {
		Colors []string `json:"colors"` // Gradient colors
		Angle  int      `json:"angle"`  // Gradient angle in degrees
	} `json:"gradient,omitempty"`
	Image string `json:"image,omitempty"` // Background image path
}

// ImageElement represents an image to embed
type ImageElement struct {
	Path     string   `json:"path"`               // File path or base64 data URI
	Position Position `json:"position"`           // Position and size
	AltText  string   `json:"altText,omitempty"`  // Accessibility text
}

// ShapeElement represents a shape (rectangle, line, etc.)
type ShapeElement struct {
	Type      string   `json:"type"`                // "rectangle", "line", "oval"
	Position  Position `json:"position"`            // Position and size
	FillColor string   `json:"fillColor,omitempty"` // Fill color (hex)
	LineColor string   `json:"lineColor,omitempty"` // Border/line color (hex)
	LineWidth float64  `json:"lineWidth,omitempty"` // Line width in points
}

// TextElement represents a positioned text box
type TextElement struct {
	Content  string    `json:"content"`            // Text content
	Position Position  `json:"position"`           // Position and size
	Style    TextStyle `json:"style,omitempty"`    // Text formatting
}

// Element is a union type for slide elements
type Element struct {
	Type  string        `json:"type"` // "text", "image", "shape"
	Text  *TextElement  `json:"text,omitempty"`
	Image *ImageElement `json:"image,omitempty"`
	Shape *ShapeElement `json:"shape,omitempty"`
}

// EnhancedSlide represents a slide with full control over elements
type EnhancedSlide struct {
	Title      string      `json:"title,omitempty"`      // Optional title (legacy support)
	Body       string      `json:"body,omitempty"`       // Optional body text (legacy support)
	Background *Background `json:"background,omitempty"` // Slide background
	Elements   []Element   `json:"elements,omitempty"`   // Positioned elements
	Notes      string      `json:"notes,omitempty"`      // Speaker notes
	Layout     string      `json:"layout,omitempty"`     // Layout preset name
}

// ThemeColors defines a color palette
type ThemeColors struct {
	Name       string `json:"name"`
	Dark1      string `json:"dk1"`      // Primary dark (usually text)
	Light1     string `json:"lt1"`      // Primary light (usually background)
	Dark2      string `json:"dk2"`      // Secondary dark
	Light2     string `json:"lt2"`      // Secondary light
	Accent1    string `json:"accent1"`  // Primary accent
	Accent2    string `json:"accent2"`  // Secondary accent
	Accent3    string `json:"accent3"`
	Accent4    string `json:"accent4"`
	Accent5    string `json:"accent5"`
	Accent6    string `json:"accent6"`
	Hyperlink  string `json:"hlink"`
	FollowLink string `json:"folHlink"`
}

// Theme defines presentation theming
type Theme struct {
	Colors    ThemeColors `json:"colors"`
	TitleFont string      `json:"titleFont,omitempty"` // Font for titles
	BodyFont  string      `json:"bodyFont,omitempty"`  // Font for body text
}

// Predefined themes
var ThemePresets = map[string]Theme{
	"default": {
		Colors: ThemeColors{
			Name:       "Office",
			Dark1:      "000000",
			Light1:     "FFFFFF",
			Dark2:      "44546A",
			Light2:     "E7E6E6",
			Accent1:    "4472C4",
			Accent2:    "ED7D31",
			Accent3:    "A5A5A5",
			Accent4:    "FFC000",
			Accent5:    "5B9BD5",
			Accent6:    "70AD47",
			Hyperlink:  "0563C1",
			FollowLink: "954F72",
		},
		TitleFont: "Calibri Light",
		BodyFont:  "Calibri",
	},
	"dark": {
		Colors: ThemeColors{
			Name:       "Dark",
			Dark1:      "FFFFFF",
			Light1:     "1E1E1E",
			Dark2:      "E0E0E0",
			Light2:     "2D2D2D",
			Accent1:    "61AFEF",
			Accent2:    "E06C75",
			Accent3:    "98C379",
			Accent4:    "E5C07B",
			Accent5:    "C678DD",
			Accent6:    "56B6C2",
			Hyperlink:  "61AFEF",
			FollowLink: "C678DD",
		},
		TitleFont: "Segoe UI",
		BodyFont:  "Segoe UI",
	},
	"light": {
		Colors: ThemeColors{
			Name:       "Light",
			Dark1:      "333333",
			Light1:     "FAFAFA",
			Dark2:      "555555",
			Light2:     "F0F0F0",
			Accent1:    "2196F3",
			Accent2:    "FF5722",
			Accent3:    "9E9E9E",
			Accent4:    "FFC107",
			Accent5:    "03A9F4",
			Accent6:    "4CAF50",
			Hyperlink:  "1976D2",
			FollowLink: "7B1FA2",
		},
		TitleFont: "Arial",
		BodyFont:  "Arial",
	},
	"teal": {
		Colors: ThemeColors{
			Name:       "Teal",
			Dark1:      "FFFFFF",
			Light1:     "023047",
			Dark2:      "E0E0E0",
			Light2:     "034764",
			Accent1:    "219EBC",
			Accent2:    "FFB703",
			Accent3:    "8ECAE6",
			Accent4:    "FB8500",
			Accent5:    "126782",
			Accent6:    "F4D35E",
			Hyperlink:  "8ECAE6",
			FollowLink: "FFB703",
		},
		TitleFont: "Arial Black",
		BodyFont:  "Arial",
	},
	"coral": {
		Colors: ThemeColors{
			Name:       "Coral",
			Dark1:      "2D3436",
			Light1:     "FFF5F5",
			Dark2:      "636E72",
			Light2:     "FFE8E8",
			Accent1:    "E17055",
			Accent2:    "00B894",
			Accent3:    "FDCB6E",
			Accent4:    "6C5CE7",
			Accent5:    "74B9FF",
			Accent6:    "A29BFE",
			Hyperlink:  "0984E3",
			FollowLink: "6C5CE7",
		},
		TitleFont: "Georgia",
		BodyFont:  "Georgia",
	},
}

// EnhancedCreateOptions extends CreateOptions with theming
type EnhancedCreateOptions struct {
	Title      string      `json:"title,omitempty"`
	Creator    string      `json:"creator,omitempty"`
	Subject    string      `json:"subject,omitempty"`
	Theme      string      `json:"theme,omitempty"`      // Theme preset name
	CustomTheme *Theme     `json:"customTheme,omitempty"` // Custom theme definition
	Background *Background `json:"background,omitempty"` // Default background for all slides
}

// PresentationInput is the JSON input format for full control
type PresentationInput struct {
	Options EnhancedCreateOptions `json:"options,omitempty"`
	Slides  []EnhancedSlide       `json:"slides"`
}

// ParseJSONInput parses JSON input for presentation creation
func ParseJSONInput(data []byte) (*PresentationInput, error) {
	var input PresentationInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return &input, nil
}

// Info contains metadata about a PowerPoint presentation.
type Info struct {
	Path       string `json:"path"`
	SlideCount int    `json:"slideCount"`
	Title      string `json:"title,omitempty"`
	Creator    string `json:"creator,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Created    string `json:"created,omitempty"`
	Modified   string `json:"modified,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Slide represents a single slide in the presentation.
type Slide struct {
	Number int    `json:"number"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text"`
}

// coreProperties represents docProps/core.xml
type coreProperties struct {
	Title    string `xml:"title"`
	Creator  string `xml:"creator"`
	Subject  string `xml:"subject"`
	Created  string `xml:"created"`
	Modified string `xml:"modified"`
}

// GetInfo returns metadata about a PowerPoint file.
func GetInfo(path string) (*Info, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	info := &Info{
		Path: path,
	}

	// Count slides
	slidePattern := regexp.MustCompile(`^ppt/slides/slide\d+\.xml$`)
	for _, f := range r.File {
		if slidePattern.MatchString(f.Name) {
			info.SlideCount++
		}
	}

	// Read core properties
	for _, f := range r.File {
		if f.Name == "docProps/core.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			props, err := parseCoreProperties(rc)
			_ = rc.Close()
			if err == nil {
				info.Title = props.Title
				info.Creator = props.Creator
				info.Subject = props.Subject
				info.Created = props.Created
				info.Modified = props.Modified
			}
			break
		}
	}

	return info, nil
}

func parseCoreProperties(r io.Reader) (*coreProperties, error) {
	var props coreProperties
	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if se, ok := token.(xml.StartElement); ok {
			switch se.Name.Local {
			case "title":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Title = val
				}
			case "creator":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Creator = val
				}
			case "subject":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Subject = val
				}
			case "created":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Created = val
				}
			case "modified":
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					props.Modified = val
				}
			}
		}
	}
	return &props, nil
}

// ExtractText extracts all text content from a PowerPoint file.
func ExtractText(path string) (string, error) {
	slides, err := GetSlides(path)
	if err != nil {
		return "", err
	}

	var parts []string
	for _, slide := range slides {
		if slide.Text != "" {
			header := fmt.Sprintf("--- Slide %d ---", slide.Number)
			if slide.Title != "" {
				header = fmt.Sprintf("--- Slide %d: %s ---", slide.Number, slide.Title)
			}
			parts = append(parts, header)
			parts = append(parts, slide.Text)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// ExtractSlideText extracts text from a specific slide (1-indexed).
func ExtractSlideText(path string, slideNum int) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideNum)
	for _, f := range r.File {
		if f.Name == slidePath {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open slide: %w", err)
			}
			text, err := extractTextFromXML(rc)
			_ = rc.Close()
			return text, err
		}
	}

	return "", fmt.Errorf("slide %d not found", slideNum)
}

// GetSlides returns all slides with their content.
func GetSlides(path string) ([]Slide, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Find all slide files
	slidePattern := regexp.MustCompile(`^ppt/slides/slide(\d+)\.xml$`)
	var slideFiles []*zip.File
	slideNumbers := make(map[*zip.File]int)

	for _, f := range r.File {
		matches := slidePattern.FindStringSubmatch(f.Name)
		if matches != nil {
			var num int
			if _, err := fmt.Sscanf(matches[1], "%d", &num); err == nil {
				slideFiles = append(slideFiles, f)
				slideNumbers[f] = num
			}
		}
	}

	// Sort by slide number
	sort.Slice(slideFiles, func(i, j int) bool {
		return slideNumbers[slideFiles[i]] < slideNumbers[slideFiles[j]]
	})

	var slides []Slide
	for _, f := range slideFiles {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		text, _ := extractTextFromXML(rc)
		_ = rc.Close()

		slide := Slide{
			Number: slideNumbers[f],
			Text:   text,
		}

		// Try to extract title (first text block is often the title)
		lines := strings.Split(text, "\n")
		if len(lines) > 0 && lines[0] != "" {
			slide.Title = lines[0]
		}

		slides = append(slides, slide)
	}

	return slides, nil
}

// extractTextFromXML extracts text from PowerPoint XML content.
// Text is contained in <a:t> tags.
func extractTextFromXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var texts []string
	var currentParagraph []string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// <a:t> contains text
			if t.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &t); err == nil && text != "" {
					currentParagraph = append(currentParagraph, text)
				}
			}
		case xml.EndElement:
			// End of paragraph <a:p>
			if t.Name.Local == "p" && len(currentParagraph) > 0 {
				texts = append(texts, strings.Join(currentParagraph, ""))
				currentParagraph = nil
			}
		}
	}

	// Add any remaining text
	if len(currentParagraph) > 0 {
		texts = append(texts, strings.Join(currentParagraph, ""))
	}

	return strings.Join(texts, "\n"), nil
}

// ListSlides returns a summary of all slides.
func ListSlides(path string) ([]Slide, error) {
	return GetSlides(path)
}

// ValidatePPTX checks if a file is a valid PPTX file.
func ValidatePPTX(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("not a valid ZIP file: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Check for required PPTX structure
	hasContentTypes := false
	hasPresentation := false

	for _, f := range r.File {
		switch f.Name {
		case "[Content_Types].xml":
			hasContentTypes = true
		case "ppt/presentation.xml":
			hasPresentation = true
		}
	}

	if !hasContentTypes {
		return fmt.Errorf("missing [Content_Types].xml - not a valid Office document")
	}
	if !hasPresentation {
		return fmt.Errorf("missing ppt/presentation.xml - not a valid PowerPoint file")
	}

	return nil
}

// IsPPTX checks if a file path appears to be a PowerPoint file.
func IsPPTX(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pptx"
}

// CreateOptions configures PPTX creation.
type CreateOptions struct {
	Title   string
	Creator string
	Subject string
}

// SlideContent represents content for a slide to create.
type SlideContent struct {
	Title string
	Body  string
}

// Create creates a new PowerPoint file with the given slides.
func Create(outputPath string, slides []SlideContent, opts CreateOptions) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	defer func() { _ = w.Close() }()

	// Write required PPTX structure
	if err := writeContentTypes(w, len(slides)); err != nil {
		return err
	}
	if err := writeRels(w); err != nil {
		return err
	}
	if err := writeCoreProps(w, opts); err != nil {
		return err
	}
	if err := writeAppProps(w); err != nil {
		return err
	}
	if err := writePresentation(w, len(slides)); err != nil {
		return err
	}
	if err := writePresentationRels(w, len(slides)); err != nil {
		return err
	}

	// Write each slide
	for i, slide := range slides {
		if err := writeSlide(w, i+1, slide); err != nil {
			return err
		}
		if err := writeSlideRels(w, i+1); err != nil {
			return err
		}
	}

	// Write slide layouts and masters (minimal)
	if err := writeSlideLayout(w); err != nil {
		return err
	}
	if err := writeSlideMaster(w); err != nil {
		return err
	}
	if err := writeTheme(w); err != nil {
		return err
	}

	return w.Close()
}

// CreateFromText creates a simple presentation from text.
// Each paragraph becomes a slide with the first line as title.
func CreateFromText(outputPath string, text string, opts CreateOptions) error {
	paragraphs := strings.Split(strings.TrimSpace(text), "\n\n")
	var slides []SlideContent

	for _, para := range paragraphs {
		lines := strings.SplitN(strings.TrimSpace(para), "\n", 2)
		slide := SlideContent{
			Title: lines[0],
		}
		if len(lines) > 1 {
			slide.Body = lines[1]
		}
		slides = append(slides, slide)
	}

	return Create(outputPath, slides, opts)
}

// imageData holds image bytes and metadata for embedding
type imageData struct {
	Data      []byte
	Extension string
	ContentType string
}

// CreateEnhanced creates a presentation with full feature support
func CreateEnhanced(outputPath string, slides []EnhancedSlide, opts EnhancedCreateOptions) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	defer func() { _ = w.Close() }()

	// Resolve theme
	theme := ThemePresets["default"]
	if opts.Theme != "" {
		if t, ok := ThemePresets[opts.Theme]; ok {
			theme = t
		}
	}
	if opts.CustomTheme != nil {
		theme = *opts.CustomTheme
	}

	// Collect all images from slides
	imageFiles := make(map[string]imageData)
	slideImageRefs := make([][]string, len(slides)) // per-slide image references

	for i, slide := range slides {
		slideImageRefs[i] = []string{}
		for _, elem := range slide.Elements {
			if elem.Type == "image" && elem.Image != nil {
				imgPath := elem.Image.Path
				if _, exists := imageFiles[imgPath]; !exists {
					imgData, err := loadImageData(imgPath)
					if err != nil {
						return fmt.Errorf("failed to load image %s: %w", imgPath, err)
					}
					imageFiles[imgPath] = imgData
				}
				slideImageRefs[i] = append(slideImageRefs[i], imgPath)
			}
		}
	}

	// Write required PPTX structure with image support
	if err := writeContentTypesEnhanced(w, len(slides), imageFiles); err != nil {
		return err
	}
	if err := writeRels(w); err != nil {
		return err
	}
	if err := writeCorePropsEnhanced(w, opts); err != nil {
		return err
	}
	if err := writeAppProps(w); err != nil {
		return err
	}
	if err := writePresentation(w, len(slides)); err != nil {
		return err
	}
	if err := writePresentationRels(w, len(slides)); err != nil {
		return err
	}

	// Write images to media folder
	imageIdxMap := make(map[string]int) // path -> media index
	mediaIdx := 1
	for imgPath, imgData := range imageFiles {
		mediaFileName := fmt.Sprintf("ppt/media/image%d.%s", mediaIdx, imgData.Extension)
		if err := writeBinaryFile(w, mediaFileName, imgData.Data); err != nil {
			return fmt.Errorf("failed to write image: %w", err)
		}
		imageIdxMap[imgPath] = mediaIdx
		mediaIdx++
	}

	// Write each slide with enhanced features
	for i, slide := range slides {
		bg := slide.Background
		if bg == nil && opts.Background != nil {
			bg = opts.Background
		}

		if err := writeEnhancedSlide(w, i+1, slide, bg, theme, slideImageRefs[i], imageIdxMap); err != nil {
			return err
		}
		if err := writeEnhancedSlideRels(w, i+1, slideImageRefs[i], imageIdxMap); err != nil {
			return err
		}
	}

	// Write slide layouts and masters
	if err := writeSlideLayoutWithTheme(w, theme); err != nil {
		return err
	}
	if err := writeSlideMasterWithTheme(w, theme); err != nil {
		return err
	}
	if err := writeThemeWithColors(w, theme); err != nil {
		return err
	}

	return w.Close()
}

// CreateFromJSON creates a presentation from JSON input
func CreateFromJSON(outputPath string, jsonData []byte) error {
	input, err := ParseJSONInput(jsonData)
	if err != nil {
		return err
	}
	return CreateEnhanced(outputPath, input.Slides, input.Options)
}

// CreateFromTextEnhanced creates a presentation from text with theme and background options
func CreateFromTextEnhanced(outputPath string, text string, opts EnhancedCreateOptions) error {
	paragraphs := strings.Split(strings.TrimSpace(text), "\n\n")
	var slides []EnhancedSlide

	for _, para := range paragraphs {
		lines := strings.SplitN(strings.TrimSpace(para), "\n", 2)
		slide := EnhancedSlide{
			Title: lines[0],
		}
		if len(lines) > 1 {
			slide.Body = lines[1]
		}
		slides = append(slides, slide)
	}

	return CreateEnhanced(outputPath, slides, opts)
}

// loadImageData loads image from file path or base64 data URI
func loadImageData(path string) (imageData, error) {
	var data imageData

	// Check for base64 data URI
	if strings.HasPrefix(path, "data:image/") {
		return parseDataURI(path)
	}

	// Read from file
	content, err := os.ReadFile(path)
	if err != nil {
		return data, err
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	data.Data = content
	data.Extension = ext
	data.ContentType = getImageContentType(ext)

	return data, nil
}

// parseDataURI parses a data:image/xxx;base64,... URI
func parseDataURI(uri string) (imageData, error) {
	var data imageData

	// Format: data:image/png;base64,iVBOR...
	if !strings.HasPrefix(uri, "data:image/") {
		return data, fmt.Errorf("invalid data URI format")
	}

	parts := strings.SplitN(uri[11:], ";", 2)
	if len(parts) < 2 {
		return data, fmt.Errorf("invalid data URI format")
	}

	ext := parts[0]
	if !strings.HasPrefix(parts[1], "base64,") {
		return data, fmt.Errorf("only base64 encoding supported")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1][7:])
	if err != nil {
		return data, fmt.Errorf("failed to decode base64: %w", err)
	}

	data.Data = decoded
	data.Extension = ext
	data.ContentType = getImageContentType(ext)

	return data, nil
}

func getImageContentType(ext string) string {
	switch ext {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func writeContentTypesEnhanced(w *zip.Writer, slideCount int, images map[string]imageData) error {
	// Collect unique image extensions
	extensions := make(map[string]string)
	for _, img := range images {
		extensions[img.Extension] = img.ContentType
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>`

	// Add image extensions
	for ext, contentType := range extensions {
		content += fmt.Sprintf(`
  <Default Extension="%s" ContentType="%s"/>`, ext, contentType)
	}

	content += `
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
  <Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
  <Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`

	for i := 1; i <= slideCount; i++ {
		content += fmt.Sprintf(`
  <Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i)
	}

	content += `
</Types>`

	return writeFile(w, "[Content_Types].xml", content)
}

func writeCorePropsEnhanced(w *zip.Writer, opts EnhancedCreateOptions) error {
	title := opts.Title
	if title == "" {
		title = "Presentation"
	}
	creator := opts.Creator
	if creator == "" {
		creator = "ps-pptx"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:creator>%s</dc:creator>
  <dc:subject>%s</dc:subject>
</cp:coreProperties>`, escapeXML(title), escapeXML(creator), escapeXML(opts.Subject))
	return writeFile(w, "docProps/core.xml", content)
}

func writeBinaryFile(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

func writeEnhancedSlide(w *zip.Writer, num int, slide EnhancedSlide, bg *Background, theme Theme, imageRefs []string, imageIdxMap map[string]int) error {
	var content strings.Builder

	content.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>`)

	// Write background if specified
	if bg != nil && bg.Color != "" {
		content.WriteString(fmt.Sprintf(`
    <p:bg>
      <p:bgPr>
        <a:solidFill>
          <a:srgbClr val="%s"/>
        </a:solidFill>
        <a:effectLst/>
      </p:bgPr>
    </p:bg>`, bg.Color))
	}

	content.WriteString(`
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr/>`)

	shapeID := 2 // Start at 2 (1 is group)
	rIdCounter := 2 // rId1 is slideLayout

	// Check if we should use legacy title/body or elements
	hasElements := len(slide.Elements) > 0

	if !hasElements && (slide.Title != "" || slide.Body != "") {
		// Legacy mode: title + body
		content.WriteString(buildTitleShape(shapeID, slide.Title, theme))
		shapeID++

		if slide.Body != "" {
			content.WriteString(buildBodyShape(shapeID, slide.Body, theme))
			shapeID++
		}
	}

	// Process elements
	imageRefIdx := 0
	for _, elem := range slide.Elements {
		switch elem.Type {
		case "text":
			if elem.Text != nil {
				content.WriteString(buildTextElement(shapeID, elem.Text, theme))
				shapeID++
			}
		case "image":
			if elem.Image != nil && imageRefIdx < len(imageRefs) {
				rId := fmt.Sprintf("rId%d", rIdCounter)
				content.WriteString(buildImageElement(shapeID, elem.Image, rId))
				shapeID++
				rIdCounter++
				imageRefIdx++
			}
		case "shape":
			if elem.Shape != nil {
				content.WriteString(buildShapeElement(shapeID, elem.Shape))
				shapeID++
			}
		}
	}

	content.WriteString(`
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>`)

	return writeFile(w, fmt.Sprintf("ppt/slides/slide%d.xml", num), content.String())
}

func buildTitleShape(id int, title string, theme Theme) string {
	titleFont := theme.TitleFont
	if titleFont == "" {
		titleFont = "Calibri Light"
	}

	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="%d" name="Title"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph type="title"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="274638"/>
            <a:ext cx="8229600" cy="1143000"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p>
            <a:r>
              <a:rPr lang="en-US" sz="4400" b="1">
                <a:latin typeface="%s"/>
              </a:rPr>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>`, id, escapeXML(titleFont), escapeXML(title))
}

func buildBodyShape(id int, body string, theme Theme) string {
	bodyFont := theme.BodyFont
	if bodyFont == "" {
		bodyFont = "Calibri"
	}

	// Build body paragraphs
	var bodyContent strings.Builder
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		bodyContent.WriteString(fmt.Sprintf(`
              <a:p>
                <a:r>
                  <a:rPr lang="en-US" sz="2400">
                    <a:latin typeface="%s"/>
                  </a:rPr>
                  <a:t>%s</a:t>
                </a:r>
              </a:p>`, escapeXML(bodyFont), escapeXML(line)))
	}

	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="%d" name="Content"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph idx="1"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="1600200"/>
            <a:ext cx="8229600" cy="4525963"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>%s
        </p:txBody>
      </p:sp>`, id, bodyContent.String())
}

func buildTextElement(id int, text *TextElement, theme Theme) string {
	x, y, cx, cy := text.Position.ToEMU()

	// Default size if not specified
	if cx == 0 {
		cx = 8229600 // ~9 inches
	}
	if cy == 0 {
		cy = 914400 // ~1 inch
	}

	// Text styling
	fontSize := text.Style.FontSize
	if fontSize == 0 {
		fontSize = 24
	}
	fontSizePt := fontSize * 100 // PowerPoint uses hundredths of a point

	fontFace := text.Style.FontFace
	if fontFace == "" {
		fontFace = theme.BodyFont
		if fontFace == "" {
			fontFace = "Calibri"
		}
	}

	// Build run properties
	var rPr strings.Builder
	rPr.WriteString(fmt.Sprintf(`<a:rPr lang="en-US" sz="%d"`, fontSizePt))
	if text.Style.Bold {
		rPr.WriteString(` b="1"`)
	}
	if text.Style.Italic {
		rPr.WriteString(` i="1"`)
	}
	if text.Style.Underline {
		rPr.WriteString(` u="sng"`)
	}
	rPr.WriteString(">")

	if text.Style.Color != "" {
		rPr.WriteString(fmt.Sprintf(`<a:solidFill><a:srgbClr val="%s"/></a:solidFill>`, text.Style.Color))
	}
	rPr.WriteString(fmt.Sprintf(`<a:latin typeface="%s"/>`, escapeXML(fontFace)))
	rPr.WriteString("</a:rPr>")

	// Text alignment
	align := ""
	switch text.Style.Align {
	case "center":
		align = ` algn="ctr"`
	case "right":
		align = ` algn="r"`
	}

	// Build paragraphs
	var paragraphs strings.Builder
	lines := strings.Split(text.Content, "\n")
	for _, line := range lines {
		paragraphs.WriteString(fmt.Sprintf(`
              <a:p%s>
                <a:r>
                  %s
                  <a:t>%s</a:t>
                </a:r>
              </a:p>`, align, rPr.String(), escapeXML(line)))
	}

	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="%d" name="TextBox %d"/>
          <p:cNvSpPr txBox="1"/>
          <p:nvPr/>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="%d" y="%d"/>
            <a:ext cx="%d" cy="%d"/>
          </a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
          <a:noFill/>
        </p:spPr>
        <p:txBody>
          <a:bodyPr wrap="square" rtlCol="0"/>
          <a:lstStyle/>%s
        </p:txBody>
      </p:sp>`, id, id, x, y, cx, cy, paragraphs.String())
}

func buildImageElement(id int, img *ImageElement, rId string) string {
	x, y, cx, cy := img.Position.ToEMU()

	// Default size if not specified
	if cx == 0 {
		cx = 2743200 // ~3 inches
	}
	if cy == 0 {
		cy = 2743200 // ~3 inches
	}

	altText := img.AltText
	if altText == "" {
		altText = "Image"
	}

	return fmt.Sprintf(`
      <p:pic>
        <p:nvPicPr>
          <p:cNvPr id="%d" name="Picture %d" descr="%s"/>
          <p:cNvPicPr>
            <a:picLocks noChangeAspect="1"/>
          </p:cNvPicPr>
          <p:nvPr/>
        </p:nvPicPr>
        <p:blipFill>
          <a:blip r:embed="%s"/>
          <a:stretch>
            <a:fillRect/>
          </a:stretch>
        </p:blipFill>
        <p:spPr>
          <a:xfrm>
            <a:off x="%d" y="%d"/>
            <a:ext cx="%d" cy="%d"/>
          </a:xfrm>
          <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
        </p:spPr>
      </p:pic>`, id, id, escapeXML(altText), rId, x, y, cx, cy)
}

func buildShapeElement(id int, shape *ShapeElement) string {
	x, y, cx, cy := shape.Position.ToEMU()

	// Default size if not specified
	if cx == 0 {
		cx = 914400 // ~1 inch
	}
	if cy == 0 {
		cy = 914400 // ~1 inch
	}

	// Shape preset
	prstGeom := "rect"
	switch shape.Type {
	case "oval", "ellipse", "circle":
		prstGeom = "ellipse"
	case "line":
		prstGeom = "line"
	case "roundRect":
		prstGeom = "roundRect"
	case "triangle":
		prstGeom = "triangle"
	}

	// Fill
	var fill string
	if shape.FillColor != "" {
		fill = fmt.Sprintf(`<a:solidFill><a:srgbClr val="%s"/></a:solidFill>`, shape.FillColor)
	} else {
		fill = `<a:noFill/>`
	}

	// Line
	var line string
	if shape.LineColor != "" {
		lineWidth := int64(shape.LineWidth * EMUPerPt)
		if lineWidth == 0 {
			lineWidth = 12700 // 1pt default
		}
		line = fmt.Sprintf(`<a:ln w="%d"><a:solidFill><a:srgbClr val="%s"/></a:solidFill></a:ln>`, lineWidth, shape.LineColor)
	}

	return fmt.Sprintf(`
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="%d" name="Shape %d"/>
          <p:cNvSpPr/>
          <p:nvPr/>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="%d" y="%d"/>
            <a:ext cx="%d" cy="%d"/>
          </a:xfrm>
          <a:prstGeom prst="%s"><a:avLst/></a:prstGeom>
          %s
          %s
        </p:spPr>
      </p:sp>`, id, id, x, y, cx, cy, prstGeom, fill, line)
}

func writeEnhancedSlideRels(w *zip.Writer, num int, imageRefs []string, imageIdxMap map[string]int) error {
	var rels strings.Builder
	rels.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>`)

	// Add image relationships
	rIdCounter := 2
	for _, imgPath := range imageRefs {
		if mediaIdx, ok := imageIdxMap[imgPath]; ok {
			ext := "png"
			if idx := strings.LastIndex(imgPath, "."); idx >= 0 {
				ext = strings.ToLower(imgPath[idx+1:])
			}
			rels.WriteString(fmt.Sprintf(`
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image%d.%s"/>`, rIdCounter, mediaIdx, ext))
			rIdCounter++
		}
	}

	rels.WriteString(`
</Relationships>`)

	return writeFile(w, fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", num), rels.String())
}

func writeSlideLayoutWithTheme(w *zip.Writer, theme Theme) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="titleOnly">
  <p:cSld name="Title Only">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`

	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideLayouts/slideLayout1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", relsContent)
}

func writeSlideMasterWithTheme(w *zip.Writer, theme Theme) error {
	// Use theme background color
	bgColor := theme.Colors.Light1
	if bgColor == "" {
		bgColor = "FFFFFF"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgPr>
        <a:solidFill>
          <a:srgbClr val="%s"/>
        </a:solidFill>
        <a:effectLst/>
      </p:bgPr>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
</p:sldMaster>`, bgColor)

	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideMasters/slideMaster1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideMasters/_rels/slideMaster1.xml.rels", relsContent)
}

func writeThemeWithColors(w *zip.Writer, theme Theme) error {
	c := theme.Colors

	// Use defaults if not specified
	if c.Dark1 == "" {
		c.Dark1 = "000000"
	}
	if c.Light1 == "" {
		c.Light1 = "FFFFFF"
	}
	if c.Dark2 == "" {
		c.Dark2 = "44546A"
	}
	if c.Light2 == "" {
		c.Light2 = "E7E6E6"
	}
	if c.Accent1 == "" {
		c.Accent1 = "4472C4"
	}
	if c.Accent2 == "" {
		c.Accent2 = "ED7D31"
	}
	if c.Accent3 == "" {
		c.Accent3 = "A5A5A5"
	}
	if c.Accent4 == "" {
		c.Accent4 = "FFC000"
	}
	if c.Accent5 == "" {
		c.Accent5 = "5B9BD5"
	}
	if c.Accent6 == "" {
		c.Accent6 = "70AD47"
	}
	if c.Hyperlink == "" {
		c.Hyperlink = "0563C1"
	}
	if c.FollowLink == "" {
		c.FollowLink = "954F72"
	}

	themeName := c.Name
	if themeName == "" {
		themeName = "Custom"
	}

	titleFont := theme.TitleFont
	if titleFont == "" {
		titleFont = "Calibri Light"
	}
	bodyFont := theme.BodyFont
	if bodyFont == "" {
		bodyFont = "Calibri"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="%s Theme">
  <a:themeElements>
    <a:clrScheme name="%s">
      <a:dk1><a:srgbClr val="%s"/></a:dk1>
      <a:lt1><a:srgbClr val="%s"/></a:lt1>
      <a:dk2><a:srgbClr val="%s"/></a:dk2>
      <a:lt2><a:srgbClr val="%s"/></a:lt2>
      <a:accent1><a:srgbClr val="%s"/></a:accent1>
      <a:accent2><a:srgbClr val="%s"/></a:accent2>
      <a:accent3><a:srgbClr val="%s"/></a:accent3>
      <a:accent4><a:srgbClr val="%s"/></a:accent4>
      <a:accent5><a:srgbClr val="%s"/></a:accent5>
      <a:accent6><a:srgbClr val="%s"/></a:accent6>
      <a:hlink><a:srgbClr val="%s"/></a:hlink>
      <a:folHlink><a:srgbClr val="%s"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="%s">
      <a:majorFont><a:latin typeface="%s"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="%s"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="%s">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
      </a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
</a:theme>`, themeName, themeName, c.Dark1, c.Light1, c.Dark2, c.Light2,
		c.Accent1, c.Accent2, c.Accent3, c.Accent4, c.Accent5, c.Accent6,
		c.Hyperlink, c.FollowLink, themeName, escapeXML(titleFont), escapeXML(bodyFont), themeName)

	return writeFile(w, "ppt/theme/theme1.xml", content)
}

// GetThemePresets returns available theme preset names
func GetThemePresets() []string {
	names := make([]string, 0, len(ThemePresets))
	for name := range ThemePresets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func writeContentTypes(w *zip.Writer, slideCount int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
  <Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
  <Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`

	for i := 1; i <= slideCount; i++ {
		content += fmt.Sprintf(`
  <Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i)
	}

	content += `
</Types>`

	return writeFile(w, "[Content_Types].xml", content)
}

func writeRels(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
	return writeFile(w, "_rels/.rels", content)
}

func writeCoreProps(w *zip.Writer, opts CreateOptions) error {
	title := opts.Title
	if title == "" {
		title = "Presentation"
	}
	creator := opts.Creator
	if creator == "" {
		creator = "ps-pptx"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:creator>%s</dc:creator>
  <dc:subject>%s</dc:subject>
</cp:coreProperties>`, escapeXML(title), escapeXML(creator), escapeXML(opts.Subject))
	return writeFile(w, "docProps/core.xml", content)
}

func writeAppProps(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
  <Application>ps-pptx</Application>
</Properties>`
	return writeFile(w, "docProps/app.xml", content)
}

func writePresentation(w *zip.Writer, slideCount int) error {
	slideList := ""
	for i := 1; i <= slideCount; i++ {
		slideList += fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 255+i, i)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldMasterIdLst>
    <p:sldMasterId id="2147483648" r:id="rId%d"/>
  </p:sldMasterIdLst>
  <p:sldIdLst>%s</p:sldIdLst>
  <p:sldSz cx="9144000" cy="6858000" type="screen4x3"/>
  <p:notesSz cx="6858000" cy="9144000"/>
</p:presentation>`, slideCount+1, slideList)
	return writeFile(w, "ppt/presentation.xml", content)
}

func writePresentationRels(w *zip.Writer, slideCount int) error {
	rels := ""
	for i := 1; i <= slideCount; i++ {
		rels += fmt.Sprintf(`
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, i, i)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">%s
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
  <Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>
</Relationships>`, rels, slideCount+1, slideCount+2)
	return writeFile(w, "ppt/_rels/presentation.xml.rels", content)
}

func writeSlide(w *zip.Writer, num int, slide SlideContent) error {
	title := escapeXML(slide.Title)
	body := escapeXML(slide.Body)

	// Build body paragraphs
	bodyContent := ""
	if body != "" {
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			bodyContent += fmt.Sprintf(`
              <a:p>
                <a:r>
                  <a:rPr lang="en-US" sz="2400"/>
                  <a:t>%s</a:t>
                </a:r>
              </a:p>`, escapeXML(line))
		}
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr/>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name="Title"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph type="title"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="274638"/>
            <a:ext cx="8229600" cy="1143000"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p>
            <a:r>
              <a:rPr lang="en-US" sz="4400" b="1"/>
              <a:t>%s</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="3" name="Content"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph idx="1"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="457200" y="1600200"/>
            <a:ext cx="8229600" cy="4525963"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>%s
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>`, title, bodyContent)
	return writeFile(w, fmt.Sprintf("ppt/slides/slide%d.xml", num), content)
}

func writeSlideRels(w *zip.Writer, num int) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`
	return writeFile(w, fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", num), content)
}

func writeSlideLayout(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="titleOnly">
  <p:cSld name="Title Only">
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`

	// Write layout rels
	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideLayouts/slideLayout1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", relsContent)
}

func writeSlideMaster(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:bg>
      <p:bgRef idx="1001">
        <a:schemeClr val="bg1"/>
      </p:bgRef>
    </p:bg>
    <p:spTree>
      <p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
</p:sldMaster>`

	relsContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`

	if err := writeFile(w, "ppt/slideMasters/slideMaster1.xml", content); err != nil {
		return err
	}
	return writeFile(w, "ppt/slideMasters/_rels/slideMaster1.xml.rels", relsContent)
}

func writeTheme(w *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
  <a:themeElements>
    <a:clrScheme name="Office">
      <a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
      <a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2>
      <a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1>
      <a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
      <a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
      <a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink>
      <a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="Office">
      <a:majorFont><a:latin typeface="Calibri Light"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="Office">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
        <a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
      </a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
</a:theme>`
	return writeFile(w, "ppt/theme/theme1.xml", content)
}

func writeFile(w *zip.Writer, name, content string) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(content))
	return err
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
