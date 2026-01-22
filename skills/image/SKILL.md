---
name: portable-image
description: Process images without system dependencies. Resize, convert formats, extract metadata.
version: 1.0.0
binary: ps-image
---

# Image Processing Skill

This skill provides image manipulation capabilities via the `ps-image` binary. It requires no system dependencies - the binary is completely self-contained.

## Available Commands

### resize
Resize an image to specified dimensions.

```bash
ps-image resize <input> --width <px> [--height <px>] [--output <path>]
```

**Flags:**
- `-w, --width`: Target width in pixels
- `--height`: Target height in pixels
- `-o, --output`: Output path (default: `<input>_resized.<ext>`)
- `--fit`: Fit mode: `contain` (default), `cover`, `fill`
- `-q, --quality`: Output quality for JPEG (1-100, default: 85)
- `-f, --format`: Output format for stdout (png, jpg, gif, bmp, tiff); auto-detected if not specified

**Fit Modes:**
- `contain`: Fit within dimensions, preserving aspect ratio (default)
- `cover`: Fill dimensions, cropping excess
- `fill`: Stretch to exact dimensions (ignores aspect ratio)

**Examples:**
```bash
# Resize to 800px width, auto height
ps-image resize photo.jpg --width 800

# Resize to exact dimensions with cover mode
ps-image resize photo.jpg --width 800 --height 600 --fit cover

# Read from stdin, write to stdout
cat input.png | ps-image resize - --width 400 -o - > output.png
```

### convert
Convert between image formats.

```bash
ps-image convert <input> --format <fmt> [--output <path>] [--quality <1-100>]
```

**Flags:**
- `-f, --format`: Target format (required): `png`, `jpg`, `gif`, `bmp`, `tiff`
- `-o, --output`: Output path (default: `<input>.<format>`)
- `-q, --quality`: Output quality for JPEG (1-100, default: 85)

**Examples:**
```bash
# Convert PNG to JPEG
ps-image convert image.png --format jpg

# Convert with specific quality
ps-image convert photo.png --format jpg --quality 90

# Convert and specify output path
ps-image convert input.bmp --format png --output result.png
```

### info
Extract image metadata as JSON or human-readable format.

```bash
ps-image info <input> [--json]
```

**Flags:**
- `--json`: Output as JSON (default: human-readable)

**Examples:**
```bash
# Human-readable output
ps-image info photo.jpg

# JSON output for programmatic use
ps-image info photo.jpg --json
```

**JSON Output Schema:**
```json
{
  "width": 1920,
  "height": 1080,
  "format": "jpeg",
  "size_bytes": 245000,
  "has_alpha": false,
  "exif": {
    "make": "Canon",
    "model": "EOS 5D Mark IV",
    "datetime": "2024-01-15T14:30:00Z",
    "exposure_time": "1/250",
    "f_number": "f/2.8",
    "iso_speed": 400,
    "focal_length": "50.0mm",
    "gps_latitude": 37.7749,
    "gps_longitude": -122.4194
  }
}
```

## Supported Formats

| Format | Read | Write |
|--------|------|-------|
| PNG    | ✓    | ✓     |
| JPEG   | ✓    | ✓     |
| GIF    | ✓    | ✓     |
| BMP    | ✓    | ✓     |
| TIFF   | ✓    | ✓     |
| WebP   | ✓    | ✗     |

Note: WebP decode is supported but encoding requires a different format output.

## Stdin/Stdout Support

All commands support `-` for stdin input and `-o -` for stdout output. **Format is auto-detected** from magic bytes when reading from stdin, so you can pipe any supported format:

```bash
# Pipe from another command
curl -s https://example.com/image.png | ps-image resize - --width 400 -o - > thumbnail.png

# Chain commands
cat large.png | ps-image resize - --width 800 -o - | ps-image convert - --format jpg -o - > result.jpg
```

## Installation

Binary is self-contained. No runtime dependencies required.

### Homebrew (macOS/Linux)
```bash
brew install Barneyjm/tap/ps-image
```

### Direct Download
Download from [GitHub Releases](https://github.com/Barneyjm/portable-skills/releases)

### npm
```bash
npx @portableskills/image <command>
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | Error (details printed to stderr) |

## Security

All binaries are:
- Built with reproducible builds
- Scanned for vulnerabilities using govulncheck
- Signed with cosign (Sigstore)
- Published with checksums for verification
