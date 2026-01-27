// Package pdf provides PDF reading and creation functionality.
package pdf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/ledongthuc/pdf"
)

// Info contains metadata about a PDF file.
type Info struct {
	Path      string `json:"path"`
	Pages     int    `json:"pages"`
	Encrypted bool   `json:"encrypted,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GetInfo returns metadata about a PDF file.
func GetInfo(path string) (*Info, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer func() { _ = f.Close() }()

	return &Info{
		Path:  path,
		Pages: r.NumPage(),
	}, nil
}

// ExtractText extracts all text content from a PDF file.
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer func() { _ = f.Close() }()

	var buf strings.Builder
	numPages := r.NumPage()

	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Continue with other pages if one fails
			continue
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(strings.TrimSpace(text))
	}

	return buf.String(), nil
}

// ExtractPageText extracts text from a specific page (1-indexed).
func ExtractPageText(path string, pageNum int) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer func() { _ = f.Close() }()

	if pageNum < 1 || pageNum > r.NumPage() {
		return "", fmt.Errorf("page %d out of range (1-%d)", pageNum, r.NumPage())
	}

	page := r.Page(pageNum)
	if page.V.IsNull() {
		return "", fmt.Errorf("page %d is empty or invalid", pageNum)
	}

	text, err := page.GetPlainText(nil)
	if err != nil {
		return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
	}

	return strings.TrimSpace(text), nil
}

// CreateOptions configures PDF creation.
type CreateOptions struct {
	Title      string
	Author     string
	PageSize   string // A4, Letter, Legal
	FontSize   float64
	FontFamily string // Helvetica, Times, Courier
	Margin     float64
}

// EnhancedCreateOptions provides full control over PDF creation.
type EnhancedCreateOptions struct {
	Title        string   `json:"title,omitempty"`
	Author       string   `json:"author,omitempty"`
	Subject      string   `json:"subject,omitempty"`
	PageSize     string   `json:"pageSize,omitempty"`     // A4, Letter, Legal
	Orientation  string   `json:"orientation,omitempty"`  // P (portrait), L (landscape)
	FontFamily   string   `json:"fontFamily,omitempty"`   // Helvetica, Times, Courier
	FontSize     float64  `json:"fontSize,omitempty"`     // Base font size in points
	MarginTop    float64  `json:"marginTop,omitempty"`    // Top margin in mm
	MarginBottom float64  `json:"marginBottom,omitempty"` // Bottom margin in mm
	MarginLeft   float64  `json:"marginLeft,omitempty"`   // Left margin in mm
	MarginRight  float64  `json:"marginRight,omitempty"`  // Right margin in mm
	Columns      int      `json:"columns,omitempty"`      // Number of columns (1-3)
	ColumnGap    float64  `json:"columnGap,omitempty"`    // Gap between columns in mm
	PageNumbers  bool     `json:"pageNumbers,omitempty"`  // Add page numbers
	PageNumPos   string   `json:"pageNumPos,omitempty"`   // Position: bottom-center, bottom-right, bottom-left
	LineSpacing  float64  `json:"lineSpacing,omitempty"`  // Line height multiplier (1.0 = normal)
	Header       *Header  `json:"header,omitempty"`       // Page header
	Footer       *Footer  `json:"footer,omitempty"`       // Page footer
}

// Header defines page header content.
type Header struct {
	Text      string `json:"text,omitempty"`
	FontSize  float64 `json:"fontSize,omitempty"`
	Alignment string `json:"alignment,omitempty"` // left, center, right
}

// Footer defines page footer content.
type Footer struct {
	Text      string `json:"text,omitempty"`
	FontSize  float64 `json:"fontSize,omitempty"`
	Alignment string `json:"alignment,omitempty"` // left, center, right
}

// TextStyle defines inline text styling.
type TextStyle struct {
	Bold      bool    `json:"bold,omitempty"`
	Italic    bool    `json:"italic,omitempty"`
	Underline bool    `json:"underline,omitempty"`
	FontSize  float64 `json:"fontSize,omitempty"`
	Color     string  `json:"color,omitempty"` // Hex color without #
}

// Paragraph represents a styled paragraph or block element.
type Paragraph struct {
	Type       string    `json:"type"`                 // text, heading, list, code, hr
	Content    string    `json:"content,omitempty"`    // Text content
	Level      int       `json:"level,omitempty"`      // Heading level (1-6)
	Style      TextStyle `json:"style,omitempty"`      // Text styling
	Items      []string  `json:"items,omitempty"`      // List items
	Ordered    bool      `json:"ordered,omitempty"`    // True for numbered lists
	Alignment  string    `json:"alignment,omitempty"`  // left, center, right, justify
	Indent     float64   `json:"indent,omitempty"`     // Left indent in mm
}

// DocumentInput is the JSON input format for structured PDF creation.
type DocumentInput struct {
	Options    EnhancedCreateOptions `json:"options,omitempty"`
	Paragraphs []Paragraph           `json:"paragraphs,omitempty"`
	Content    string                `json:"content,omitempty"`    // Plain text content (alternative to paragraphs)
	Markdown   string                `json:"markdown,omitempty"`   // Markdown content (alternative to paragraphs)
}

// ParseJSONInput parses JSON input for document creation.
func ParseJSONInput(data []byte) (*DocumentInput, error) {
	var input DocumentInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return &input, nil
}

// DefaultCreateOptions returns sensible defaults for PDF creation.
func DefaultCreateOptions() CreateOptions {
	return CreateOptions{
		PageSize:   "A4",
		FontSize:   12,
		FontFamily: "Helvetica",
		Margin:     20,
	}
}

// CreateFromText creates a PDF from plain text content.
func CreateFromText(outputPath string, text string, opts CreateOptions) error {
	if opts.PageSize == "" {
		opts.PageSize = "A4"
	}
	if opts.FontSize == 0 {
		opts.FontSize = 12
	}
	if opts.FontFamily == "" {
		opts.FontFamily = "Helvetica"
	}
	if opts.Margin == 0 {
		opts.Margin = 20
	}

	pdf := fpdf.New("P", "mm", opts.PageSize, "")

	if opts.Title != "" {
		pdf.SetTitle(opts.Title, true)
	}
	if opts.Author != "" {
		pdf.SetAuthor(opts.Author, true)
	}

	pdf.SetMargins(opts.Margin, opts.Margin, opts.Margin)
	pdf.SetAutoPageBreak(true, opts.Margin)

	pdf.AddPage()
	pdf.SetFont(opts.FontFamily, "", opts.FontSize)

	// Write text with automatic line wrapping
	pdf.MultiCell(0, opts.FontSize*0.5, text, "", "", false)

	return pdf.OutputFileAndClose(outputPath)
}

// CreateFromReader creates a PDF from text read from an io.Reader.
func CreateFromReader(outputPath string, r io.Reader, opts CreateOptions) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	return CreateFromText(outputPath, string(content), opts)
}

// CreateFromFile creates a PDF from a text file.
func CreateFromFile(outputPath string, inputPath string, opts CreateOptions) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return CreateFromText(outputPath, string(content), opts)
}

// DefaultEnhancedOptions returns sensible defaults for enhanced PDF creation.
func DefaultEnhancedOptions() EnhancedCreateOptions {
	return EnhancedCreateOptions{
		PageSize:     "A4",
		Orientation:  "P",
		FontFamily:   "Helvetica",
		FontSize:     12,
		MarginTop:    20,
		MarginBottom: 20,
		MarginLeft:   20,
		MarginRight:  20,
		Columns:      1,
		ColumnGap:    10,
		LineSpacing:  1.2,
		PageNumPos:   "bottom-center",
	}
}

// CreateEnhanced creates a PDF with full formatting control.
func CreateEnhanced(outputPath string, paragraphs []Paragraph, opts EnhancedCreateOptions) error {
	// Apply defaults
	defaults := DefaultEnhancedOptions()
	if opts.PageSize == "" {
		opts.PageSize = defaults.PageSize
	}
	if opts.Orientation == "" {
		opts.Orientation = defaults.Orientation
	}
	if opts.FontFamily == "" {
		opts.FontFamily = defaults.FontFamily
	}
	if opts.FontSize == 0 {
		opts.FontSize = defaults.FontSize
	}
	if opts.MarginTop == 0 {
		opts.MarginTop = defaults.MarginTop
	}
	if opts.MarginBottom == 0 {
		opts.MarginBottom = defaults.MarginBottom
	}
	if opts.MarginLeft == 0 {
		opts.MarginLeft = defaults.MarginLeft
	}
	if opts.MarginRight == 0 {
		opts.MarginRight = defaults.MarginRight
	}
	if opts.LineSpacing == 0 {
		opts.LineSpacing = defaults.LineSpacing
	}
	if opts.Columns == 0 {
		opts.Columns = 1
	}
	if opts.ColumnGap == 0 {
		opts.ColumnGap = defaults.ColumnGap
	}
	if opts.PageNumPos == "" {
		opts.PageNumPos = defaults.PageNumPos
	}

	p := fpdf.New(opts.Orientation, "mm", opts.PageSize, "")

	// Set metadata
	if opts.Title != "" {
		p.SetTitle(opts.Title, true)
	}
	if opts.Author != "" {
		p.SetAuthor(opts.Author, true)
	}
	if opts.Subject != "" {
		p.SetSubject(opts.Subject, true)
	}

	// Set margins
	p.SetMargins(opts.MarginLeft, opts.MarginTop, opts.MarginRight)
	p.SetAutoPageBreak(true, opts.MarginBottom)

	// Calculate column width
	pageWidth, _ := p.GetPageSize()
	contentWidth := pageWidth - opts.MarginLeft - opts.MarginRight
	columnWidth := contentWidth
	if opts.Columns > 1 {
		columnWidth = (contentWidth - float64(opts.Columns-1)*opts.ColumnGap) / float64(opts.Columns)
	}

	// Add page number function if enabled
	if opts.PageNumbers {
		p.SetFooterFunc(func() {
			p.SetY(-15)
			p.SetFont(opts.FontFamily, "", 10)
			p.SetTextColor(128, 128, 128)
			pageNum := fmt.Sprintf("Page %d", p.PageNo())

			var align string
			switch opts.PageNumPos {
			case "bottom-left":
				align = "L"
			case "bottom-right":
				align = "R"
			default:
				align = "C"
			}
			p.CellFormat(0, 10, pageNum, "", 0, align, false, 0, "")
		})
	}

	// Add header function if specified
	if opts.Header != nil && opts.Header.Text != "" {
		headerOpts := opts.Header
		p.SetHeaderFunc(func() {
			fontSize := headerOpts.FontSize
			if fontSize == 0 {
				fontSize = 10
			}
			p.SetFont(opts.FontFamily, "", fontSize)
			p.SetTextColor(128, 128, 128)

			var align string
			switch headerOpts.Alignment {
			case "left":
				align = "L"
			case "right":
				align = "R"
			default:
				align = "C"
			}
			p.CellFormat(0, 10, headerOpts.Text, "", 1, align, false, 0, "")
			p.Ln(5)
		})
	}

	p.AddPage()

	// Process paragraphs
	// Note: Multi-column layout is defined but uses single column rendering
	// Full multi-column support would require tracking Y positions per column
	_ = columnWidth // Reserved for future multi-column implementation

	for _, para := range paragraphs {
		switch para.Type {
		case "heading":
			renderHeading(p, para, opts)
		case "list":
			renderList(p, para, opts)
		case "code":
			renderCode(p, para, opts)
		case "hr":
			renderHorizontalRule(p, opts)
		default: // "text" or empty
			renderParagraph(p, para, opts)
		}
	}

	return p.OutputFileAndClose(outputPath)
}

// renderHeading renders a heading paragraph.
func renderHeading(p *fpdf.Fpdf, para Paragraph, opts EnhancedCreateOptions) {
	level := para.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	// Calculate font size based on level
	sizes := []float64{24, 20, 16, 14, 12, 11} // H1-H6
	fontSize := sizes[level-1]

	p.Ln(fontSize * 0.3) // Space before heading
	p.SetFont(opts.FontFamily, "B", fontSize)

	align := getAlignment(para.Alignment)
	p.MultiCell(0, fontSize*opts.LineSpacing*0.4, para.Content, "", align, false)
	p.Ln(fontSize * 0.2) // Space after heading

	// Reset font
	p.SetFont(opts.FontFamily, "", opts.FontSize)
}

// renderParagraph renders a text paragraph with optional styling.
func renderParagraph(p *fpdf.Fpdf, para Paragraph, opts EnhancedCreateOptions) {
	fontSize := opts.FontSize
	if para.Style.FontSize > 0 {
		fontSize = para.Style.FontSize
	}

	// Build font style
	var style string
	if para.Style.Bold {
		style += "B"
	}
	if para.Style.Italic {
		style += "I"
	}
	if para.Style.Underline {
		style += "U"
	}

	p.SetFont(opts.FontFamily, style, fontSize)

	// Set color if specified
	if para.Style.Color != "" {
		r, g, b := hexToRGB(para.Style.Color)
		p.SetTextColor(r, g, b)
	}

	// Handle indentation
	if para.Indent > 0 {
		p.SetX(opts.MarginLeft + para.Indent)
	}

	align := getAlignment(para.Alignment)
	lineHeight := fontSize * opts.LineSpacing * 0.4
	p.MultiCell(0, lineHeight, para.Content, "", align, false)
	p.Ln(lineHeight * 0.5) // Paragraph spacing

	// Reset color and font
	p.SetTextColor(0, 0, 0)
	p.SetFont(opts.FontFamily, "", opts.FontSize)
}

// renderList renders a bullet or numbered list.
func renderList(p *fpdf.Fpdf, para Paragraph, opts EnhancedCreateOptions) {
	p.SetFont(opts.FontFamily, "", opts.FontSize)
	lineHeight := opts.FontSize * opts.LineSpacing * 0.4
	indent := 10.0 // Indent for list items

	for i, item := range para.Items {
		var bullet string
		if para.Ordered {
			bullet = fmt.Sprintf("%d. ", i+1)
		} else {
			bullet = "• "
		}

		// Draw bullet/number
		x := p.GetX()
		p.SetX(x + indent)
		p.CellFormat(10, lineHeight, bullet, "", 0, "R", false, 0, "")

		// Draw item text
		p.MultiCell(0, lineHeight, item, "", "L", false)
	}
	p.Ln(lineHeight * 0.5)
}

// renderCode renders a code block with monospace font.
func renderCode(p *fpdf.Fpdf, para Paragraph, opts EnhancedCreateOptions) {
	p.SetFont("Courier", "", opts.FontSize-1)
	p.SetFillColor(245, 245, 245)

	lineHeight := opts.FontSize * opts.LineSpacing * 0.35

	// Add some padding
	p.Ln(2)

	lines := strings.Split(para.Content, "\n")
	for _, line := range lines {
		p.CellFormat(0, lineHeight, "  "+line, "", 1, "L", true, 0, "")
	}

	p.Ln(4)
	p.SetFont(opts.FontFamily, "", opts.FontSize)
}

// renderHorizontalRule renders a horizontal line.
func renderHorizontalRule(p *fpdf.Fpdf, opts EnhancedCreateOptions) {
	p.Ln(5)
	pageWidth, _ := p.GetPageSize()
	y := p.GetY()
	p.Line(opts.MarginLeft, y, pageWidth-opts.MarginRight, y)
	p.Ln(5)
}

// getAlignment converts string alignment to fpdf alignment code.
func getAlignment(align string) string {
	switch strings.ToLower(align) {
	case "center":
		return "C"
	case "right":
		return "R"
	case "justify":
		return "J"
	default:
		return "L"
	}
}

// hexToRGB converts a hex color string to RGB values.
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	var r, g, b int
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

// CreateFromJSON creates a PDF from JSON input.
func CreateFromJSON(outputPath string, jsonData []byte) error {
	input, err := ParseJSONInput(jsonData)
	if err != nil {
		return err
	}

	// Determine content source
	var paragraphs []Paragraph

	if len(input.Paragraphs) > 0 {
		paragraphs = input.Paragraphs
	} else if input.Markdown != "" {
		paragraphs = ParseMarkdown(input.Markdown)
	} else if input.Content != "" {
		// Plain text: create a single text paragraph
		paragraphs = []Paragraph{{
			Type:    "text",
			Content: input.Content,
		}}
	}

	return CreateEnhanced(outputPath, paragraphs, input.Options)
}

// CreateFromMarkdown creates a PDF from Markdown text.
func CreateFromMarkdown(outputPath string, markdown string, opts EnhancedCreateOptions) error {
	paragraphs := ParseMarkdown(markdown)
	return CreateEnhanced(outputPath, paragraphs, opts)
}

// ParseMarkdown parses Markdown text into paragraphs.
func ParseMarkdown(text string) []Paragraph {
	var paragraphs []Paragraph

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Split into blocks by blank lines (for main structure)
	lines := strings.Split(text, "\n")

	var currentBlock []string
	inCodeBlock := false
	codeBlockContent := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// End code block
				paragraphs = append(paragraphs, Paragraph{
					Type:    "code",
					Content: strings.TrimSuffix(codeBlockContent, "\n"),
				})
				codeBlockContent = ""
				inCodeBlock = false
			} else {
				// Flush current block first
				if len(currentBlock) > 0 {
					para := parseBlock(currentBlock)
					if para.Content != "" || len(para.Items) > 0 {
						paragraphs = append(paragraphs, para)
					}
					currentBlock = nil
				}
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBlockContent += line + "\n"
			continue
		}

		// Handle horizontal rules
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			if len(currentBlock) > 0 {
				para := parseBlock(currentBlock)
				if para.Content != "" || len(para.Items) > 0 {
					paragraphs = append(paragraphs, para)
				}
				currentBlock = nil
			}
			paragraphs = append(paragraphs, Paragraph{Type: "hr"})
			continue
		}

		// Handle headings
		if strings.HasPrefix(trimmed, "#") {
			// Flush current block
			if len(currentBlock) > 0 {
				para := parseBlock(currentBlock)
				if para.Content != "" || len(para.Items) > 0 {
					paragraphs = append(paragraphs, para)
				}
				currentBlock = nil
			}

			level := 0
			for _, c := range trimmed {
				if c == '#' {
					level++
				} else {
					break
				}
			}
			if level > 6 {
				level = 6
			}

			content := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			paragraphs = append(paragraphs, Paragraph{
				Type:    "heading",
				Level:   level,
				Content: processInlineFormatting(content),
			})
			continue
		}

		// Handle list items
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "+ ") {
			// Check if we're continuing a list
			if len(currentBlock) > 0 && !isListItem(currentBlock[len(currentBlock)-1]) {
				para := parseBlock(currentBlock)
				if para.Content != "" || len(para.Items) > 0 {
					paragraphs = append(paragraphs, para)
				}
				currentBlock = nil
			}
			currentBlock = append(currentBlock, line)
			continue
		}

		// Handle numbered lists
		if matchNumberedList(trimmed) {
			if len(currentBlock) > 0 && !matchNumberedList(strings.TrimSpace(currentBlock[len(currentBlock)-1])) {
				para := parseBlock(currentBlock)
				if para.Content != "" || len(para.Items) > 0 {
					paragraphs = append(paragraphs, para)
				}
				currentBlock = nil
			}
			currentBlock = append(currentBlock, line)
			continue
		}

		// Blank line: flush current block
		if trimmed == "" {
			if len(currentBlock) > 0 {
				para := parseBlock(currentBlock)
				if para.Content != "" || len(para.Items) > 0 {
					paragraphs = append(paragraphs, para)
				}
				currentBlock = nil
			}
			continue
		}

		// Regular text line
		currentBlock = append(currentBlock, line)
	}

	// Flush remaining block
	if len(currentBlock) > 0 {
		para := parseBlock(currentBlock)
		if para.Content != "" || len(para.Items) > 0 {
			paragraphs = append(paragraphs, para)
		}
	}

	// Handle unclosed code block
	if inCodeBlock && codeBlockContent != "" {
		paragraphs = append(paragraphs, Paragraph{
			Type:    "code",
			Content: strings.TrimSuffix(codeBlockContent, "\n"),
		})
	}

	return paragraphs
}

// parseBlock parses a block of lines into a paragraph.
func parseBlock(lines []string) Paragraph {
	if len(lines) == 0 {
		return Paragraph{}
	}

	// Check if it's a list
	first := strings.TrimSpace(lines[0])
	if isListItem(first) {
		var items []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Remove bullet markers
			for _, prefix := range []string{"- ", "* ", "+ "} {
				if strings.HasPrefix(trimmed, prefix) {
					items = append(items, processInlineFormatting(strings.TrimPrefix(trimmed, prefix)))
					break
				}
			}
		}
		return Paragraph{
			Type:    "list",
			Items:   items,
			Ordered: false,
		}
	}

	if matchNumberedList(first) {
		var items []string
		re := regexp.MustCompile(`^\d+\.\s+`)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if matchNumberedList(trimmed) {
				content := re.ReplaceAllString(trimmed, "")
				items = append(items, processInlineFormatting(content))
			}
		}
		return Paragraph{
			Type:    "list",
			Items:   items,
			Ordered: true,
		}
	}

	// Regular paragraph: join lines with spaces
	var content strings.Builder
	for i, line := range lines {
		if i > 0 {
			content.WriteString(" ")
		}
		content.WriteString(strings.TrimSpace(line))
	}

	return Paragraph{
		Type:    "text",
		Content: processInlineFormatting(content.String()),
	}
}

// isListItem checks if a line is a bullet list item.
func isListItem(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "- ") ||
		strings.HasPrefix(trimmed, "* ") ||
		strings.HasPrefix(trimmed, "+ ")
}

// matchNumberedList checks if a line is a numbered list item.
func matchNumberedList(line string) bool {
	re := regexp.MustCompile(`^\d+\.\s+`)
	return re.MatchString(line)
}

// processInlineFormatting handles bold and italic markers.
// Note: fpdf doesn't support mixed formatting in a single cell,
// so we strip the markers and return clean text.
// For full support, we'd need to use HTML2PDF or split into segments.
func processInlineFormatting(text string) string {
	// Remove bold markers (**text** or __text__)
	boldRe := regexp.MustCompile(`\*\*(.+?)\*\*|__(.+?)__`)
	text = boldRe.ReplaceAllString(text, "$1$2")

	// Remove italic markers (*text* or _text_)
	italicRe := regexp.MustCompile(`\*(.+?)\*|_(.+?)_`)
	text = italicRe.ReplaceAllString(text, "$1$2")

	// Remove inline code markers
	codeRe := regexp.MustCompile("`(.+?)`")
	text = codeRe.ReplaceAllString(text, "$1")

	return text
}

// CreateEnhancedFromText creates a PDF from plain text with enhanced options.
func CreateEnhancedFromText(outputPath string, text string, opts EnhancedCreateOptions) error {
	paragraphs := []Paragraph{{
		Type:    "text",
		Content: text,
	}}
	return CreateEnhanced(outputPath, paragraphs, opts)
}
