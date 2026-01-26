---
name: portable-pdf
description: Extract text from PDFs and create PDFs from text. No system dependencies required.
version: 1.0.0
binary: ps-pdf
---

# PDF Skill

Extract text from PDF files and create PDFs from text. No dependencies required.

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
Create a PDF from text input.

```bash
ps-pdf create <input> --output <file.pdf> [options]
```

**Arguments:**
- `input` - Text file path or `-` for stdin

**Flags:**
- `--output, -o` - Output PDF path (required)
- `--title` - Document title
- `--author` - Document author
- `--page-size` - Page size: A4 (default), Letter, Legal
- `--font-size` - Font size in points (default: 12)
- `--font` - Font family: Helvetica (default), Times, Courier

**Examples:**
```bash
# Create PDF from text file
ps-pdf create notes.txt -o notes.pdf

# Create PDF from stdin with title
echo "Hello World" | ps-pdf create - -o hello.pdf --title "Greeting"

# Create PDF with custom formatting
ps-pdf create report.txt -o report.pdf --page-size Letter --font Times --font-size 11
```

## Use Cases

1. **Extract text for analysis**: Pull text from PDFs for processing
2. **Convert text to PDF**: Create shareable PDFs from plain text
3. **Check PDF info**: Verify page counts and metadata
4. **Batch processing**: Process multiple PDFs in scripts

## Limitations

- Text extraction quality depends on the PDF structure (scanned PDFs may not work well)
- Created PDFs use basic formatting (no images or complex layouts)
- Encrypted PDFs may not be readable
