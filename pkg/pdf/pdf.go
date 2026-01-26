// Package pdf provides PDF reading and creation functionality.
package pdf

import (
	"fmt"
	"io"
	"os"
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
