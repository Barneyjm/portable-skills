---
name: portable-pptx
description: Read and create PowerPoint (.pptx) files. Extract text, get info, create presentations from text.
version: 1.0.0
binary: ps-pptx
---

# PowerPoint Skill

Read and create PowerPoint presentations without any dependencies.

## Available Commands

### info
Display presentation metadata.

```bash
ps-pptx info <file> [file...] [--json]
```

**Arguments:**
- `file` - Path to .pptx file(s)

**Flags:**
- `--json` - Output as JSON

**Examples:**
```bash
# Get info about a presentation
ps-pptx info presentation.pptx

# Get info as JSON
ps-pptx info deck.pptx --json
```

### text
Extract text content from a presentation.

```bash
ps-pptx text <file> [--slide <num>] [--output <file>]
```

**Arguments:**
- `file` - Path to .pptx file

**Flags:**
- `--slide, -s` - Extract only a specific slide (1-indexed)
- `--output, -o` - Write to file instead of stdout

**Examples:**
```bash
# Extract all text
ps-pptx text presentation.pptx

# Extract text from slide 3
ps-pptx text presentation.pptx --slide 3

# Save to file
ps-pptx text presentation.pptx -o content.txt
```

### list
List all slides with titles.

```bash
ps-pptx list <file> [--json]
```

**Arguments:**
- `file` - Path to .pptx file

**Flags:**
- `--json` - Output as JSON

**Examples:**
```bash
# List slides
ps-pptx list presentation.pptx

# List as JSON
ps-pptx list presentation.pptx --json
```

### create
Create a PowerPoint from text input.

```bash
ps-pptx create <input> --output <file.pptx> [options]
```

**Arguments:**
- `input` - Text file path or `-` for stdin

**Flags:**
- `--output, -o` - Output .pptx path (required)
- `--title` - Presentation title
- `--creator` - Creator name

**Input Format:**
Each paragraph (separated by blank lines) becomes a slide.
The first line of each paragraph is the slide title.

**Examples:**
```bash
# Create from text file
ps-pptx create notes.txt -o presentation.pptx

# Create from stdin
cat <<EOF | ps-pptx create - -o deck.pptx --title "My Deck"
Introduction
Welcome to my presentation.

Main Points
- Point one
- Point two
- Point three

Conclusion
Thank you for watching!
EOF
```

## Use Cases

1. **Extract text for analysis**: Pull all text from presentations
2. **Create simple decks**: Generate presentations from plain text
3. **Batch processing**: Process multiple .pptx files in scripts
4. **Template workflows**: Read a template, modify content, save new file

## Notes

- Template use is essentially a read + create operation
- Only .pptx format is supported (not .ppt)
- Created presentations use a minimal theme
- Images and complex formatting are preserved when reading but not when creating
