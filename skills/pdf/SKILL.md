---
name: portable-pdf
description: Extract text from PDFs and create PDFs from text, Markdown, or JSON. No dependencies required.
---

# PDF Skill

Extract text from PDF files and create professional PDFs from text, Markdown, or structured JSON. No dependencies required.

## Available Commands

### info
Display PDF metadata.

```bash
ps-pdf info <file> [file...] [--json]
```

**Arguments:**
- `file` - Path to PDF file(s)

**Flags:**
- `--json` - Output as JSON

**Examples:**
```bash
# Get info about a PDF
ps-pdf info document.pdf

# Get info about multiple PDFs as JSON
ps-pdf info *.pdf --json
```

### text
Extract text content from a PDF.

```bash
ps-pdf text <file> [--page <num>] [--output <file>]
```

**Arguments:**
- `file` - Path to PDF file

**Flags:**
- `--page, -p` - Extract only a specific page (1-indexed)
- `--output, -o` - Write to file instead of stdout

**Examples:**
```bash
# Extract all text
ps-pdf text document.pdf

# Extract text from page 3
ps-pdf text document.pdf --page 3

# Save extracted text to file
ps-pdf text document.pdf -o content.txt
```

### create
Create a PDF from text, Markdown, or JSON input.

```bash
ps-pdf create <input> --output <file.pdf> [options]
```

**Arguments:**
- `input` - Text/Markdown/JSON file path or `-` for stdin

**Flags:**
- `--output, -o` - Output PDF path (required)
- `--title` - Document title
- `--author` - Document author
- `--page-size` - Page size: A4 (default), Letter, Legal
- `--orientation` - P (portrait, default), L (landscape)
- `--font-size` - Font size in points (default: 12)
- `--font` - Font family: Helvetica (default), Times, Courier
- `--markdown` - Parse input as Markdown
- `--json` - Parse input as JSON for full control
- `--margin-top` - Top margin in mm (default: 20)
- `--margin-bottom` - Bottom margin in mm (default: 20)
- `--margin-left` - Left margin in mm (default: 20)
- `--margin-right` - Right margin in mm (default: 20)
- `--page-numbers` - Add page numbers
- `--page-num-pos` - Position: bottom-left, bottom-center (default), bottom-right
- `--line-spacing` - Line height multiplier (default: 1.2)

**Plain Text Examples:**
```bash
# Create PDF from text file
ps-pdf create notes.txt -o notes.pdf

# Create PDF with custom formatting
ps-pdf create report.txt -o report.pdf --page-size Letter --font Times --font-size 11

# Create PDF with page numbers and margins
ps-pdf create document.txt -o doc.pdf --page-numbers --margin-top 25 --margin-bottom 25
```

**Markdown Examples:**
```bash
# Create PDF from Markdown
ps-pdf create readme.md -o readme.pdf --markdown

# Markdown with page numbers
ps-pdf create docs.md -o docs.pdf --markdown --page-numbers --title "Documentation"
```

Markdown supports:
- Headings (# H1 through ###### H6)
- Bullet lists (-, *, +)
- Numbered lists (1., 2., 3.)
- Code blocks (triple backticks)
- Horizontal rules (---, ***, ___)
- Paragraphs (blank line separation)

**JSON Examples:**
```bash
# Create from JSON file
ps-pdf create layout.json -o report.pdf --json

# Pipe JSON from stdin
echo '{"markdown": "# Hello World"}' | ps-pdf create - -o hello.pdf --json
```

JSON with embedded Markdown:
```json
{
  "options": {
    "title": "My Document",
    "author": "Jane Doe",
    "pageNumbers": true,
    "marginTop": 25,
    "marginBottom": 25
  },
  "markdown": "# Introduction\n\nThis is a paragraph.\n\n- Item 1\n- Item 2\n\n## Section 2\n\nMore content here."
}
```

JSON with explicit paragraphs (full control):
```json
{
  "options": {
    "title": "Report",
    "pageNumbers": true,
    "fontFamily": "Times"
  },
  "paragraphs": [
    { "type": "heading", "level": 1, "content": "Introduction" },
    { "type": "text", "content": "This is the intro paragraph." },
    { "type": "text", "content": "Bold paragraph.", "style": { "bold": true } },
    { "type": "list", "items": ["First item", "Second item"], "ordered": false },
    { "type": "list", "items": ["Step one", "Step two"], "ordered": true },
    { "type": "code", "content": "function hello() {\n  return 'world';\n}" },
    { "type": "hr" },
    { "type": "heading", "level": 2, "content": "Conclusion" },
    { "type": "text", "content": "Centered text.", "alignment": "center" }
  ]
}
```

## Use Cases

1. **Extract text for analysis**: Pull text from PDFs for processing
2. **Convert Markdown to PDF**: Create professional PDFs from Markdown files
3. **Structured documents**: Use JSON mode for complex layouts with headings, lists, and code blocks
4. **Reports with page numbers**: Generate paginated documents
5. **Custom formatting**: Control margins, fonts, and spacing
6. **Batch processing**: Process multiple PDFs in scripts

## Limitations

- Text extraction quality depends on the PDF structure (scanned PDFs may not work well)
- Encrypted PDFs may not be readable
- Limited inline formatting (bold/italic markers in Markdown are stripped for clean text)
- Images not yet supported in PDF creation
