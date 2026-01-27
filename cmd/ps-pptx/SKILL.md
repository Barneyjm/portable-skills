---
name: portable-pptx
description: Read and create PowerPoint (.pptx) files with themes, images, shapes, and full positioning control.
version: 2.0.0
binary: ps-pptx
---

# PowerPoint Skill

Read and create PowerPoint presentations without any dependencies. Supports themes, background colors, images, shapes, and positioned elements.

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
Create a PowerPoint from text or JSON input.

```bash
ps-pptx create <input> --output <file.pptx> [options]
```

**Arguments:**
- `input` - Text/JSON file path or `-` for stdin

**Flags:**
- `--output, -o` - Output .pptx path (required)
- `--title` - Presentation title
- `--creator` - Creator name
- `--theme` - Theme preset: default, dark, light, teal, coral
- `--background` - Default background color (hex without #)
- `--json` - Parse input as JSON for full control
- `--image` - Add image (repeatable): slide:N,path:FILE,x:X,y:Y,w:W,h:H

**Text Input Format:**
Each paragraph (separated by blank lines) becomes a slide.
The first line of each paragraph is the slide title.

**Examples:**
```bash
# Create from text file
ps-pptx create notes.txt -o presentation.pptx

# Create with dark theme
ps-pptx create notes.txt -o deck.pptx --theme dark

# Create with teal theme and custom background
ps-pptx create notes.txt -o deck.pptx --theme teal --background 023047

# Create from stdin with image
cat <<EOF | ps-pptx create - -o deck.pptx --title "My Deck" --image slide:5,path:qr.png,x:6.5,y:1,w:3,h:3
Introduction
Welcome to my presentation.

Main Points
- Point one
- Point two

Call to Action
Scan the QR code to get started!
EOF
```

**JSON Input Mode:**
Use `--json` for full control over positioning, styling, and elements.

```bash
ps-pptx create presentation.json -o deck.pptx --json
```

**JSON Schema:**
```json
{
  "options": {
    "title": "Presentation Title",
    "creator": "Author Name",
    "theme": "teal",
    "background": {"color": "023047"}
  },
  "slides": [
    {
      "title": "Slide Title",
      "body": "Body text content",
      "background": {"color": "034764"}
    },
    {
      "elements": [
        {
          "type": "text",
          "text": {
            "content": "Custom positioned text",
            "position": {"x": 1, "y": 2, "width": 8, "height": 1},
            "style": {
              "fontSize": 36,
              "fontFace": "Arial Black",
              "bold": true,
              "italic": false,
              "color": "FFFFFF",
              "align": "center"
            }
          }
        },
        {
          "type": "image",
          "image": {
            "path": "qr.png",
            "position": {"x": 6.5, "y": 1, "width": 3, "height": 3},
            "altText": "QR Code"
          }
        },
        {
          "type": "shape",
          "shape": {
            "type": "rectangle",
            "position": {"x": 0.5, "y": 0.5, "width": 9, "height": 0.1},
            "fillColor": "FFB703",
            "lineColor": "000000",
            "lineWidth": 1
          }
        }
      ]
    }
  ]
}
```

**Element Types:**
- `text` - Positioned text box with styling
- `image` - Embedded image (PNG, JPG, GIF, BMP, WebP)
- `shape` - Rectangle, oval, line, roundRect, triangle

**Position Units:**
All positions are in inches from top-left. Standard slide is 10" x 7.5".

### themes
List available theme presets.

```bash
ps-pptx themes [--json]
```

**Available Themes:**
- `default` - Classic Office blue theme
- `dark` - Dark mode with syntax-highlighting-inspired colors
- `light` - Clean light theme with modern colors
- `teal` - Professional teal/gold palette (great for business)
- `coral` - Warm coral/green accent palette

**Examples:**
```bash
# List themes
ps-pptx themes

# Get theme details as JSON
ps-pptx themes --json
```

## Use Cases

1. **Professional presentations**: Use themes and backgrounds for polished decks
2. **QR code slides**: Embed QR codes with precise positioning
3. **Branded content**: Custom colors and fonts for brand consistency
4. **Data visualization**: Combine text, shapes, and images
5. **Batch generation**: Create presentations programmatically via JSON
6. **Extract text for analysis**: Pull all text from presentations

## Notes

- Only .pptx format is supported (not .ppt)
- Images support: PNG, JPG, GIF, BMP, WebP, and base64 data URIs
- Shape types: rectangle, oval, line, roundRect, triangle
- Themes affect default fonts and can be combined with custom backgrounds
- When using elements array, title/body are ignored (use text elements instead)
