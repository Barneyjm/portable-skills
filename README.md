# Portable Skills

Pre-compiled Go binaries that implement common skill capabilities for AI agents. Single file, zero dependencies, cross-platform.

[![Build](https://github.com/Barneyjm/portable-skills/actions/workflows/build.yml/badge.svg)](https://github.com/Barneyjm/portable-skills/actions/workflows/build.yml)
[![Security](https://github.com/Barneyjm/portable-skills/actions/workflows/security.yml/badge.svg)](https://github.com/Barneyjm/portable-skills/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Barneyjm/portable-skills)](https://goreportcard.com/report/github.com/Barneyjm/portable-skills)

## The Problem

The current Agent Skills ecosystem assumes everyone has:
- Python/Node runtimes installed
- Admin rights to install system dependencies
- Build toolchains (gcc, Xcode, Visual Studio)
- Time to debug `pip install` failures

This excludes 99% of computer users. Skills that need image processing, PDF handling, or anything beyond pure text are effectively broken for non-developers.

## The Solution

```bash
# Instead of this nightmare:
pip install pillow  # needs libjpeg, libpng, zlib...
brew install libvips  # hope you have Homebrew
npm install sharp  # might work, might not

# Just this:
./ps-image resize input.png --width 800
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install Barneyjm/tap/portable-skills
```

### Direct Download

Download from [GitHub Releases](https://github.com/Barneyjm/portable-skills/releases/latest)

```bash
# Linux/macOS
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64
chmod +x ps-image-linux-amd64
sudo mv ps-image-linux-amd64 /usr/local/bin/ps-image

# Verify signature (optional but recommended)
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64.sig
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64.pem
cosign verify-blob --signature ps-image-linux-amd64.sig --certificate ps-image-linux-amd64.pem \
  --certificate-identity-regexp "https://github.com/Barneyjm/portable-skills/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ps-image-linux-amd64
```

### npm

```bash
# One-time execution
npx @portableskills/image resize input.png --width 800

# Or install globally
npm install -g @portableskills/image
ps-image resize input.png --width 800
```

## Available Skills

| Skill | Description | Commands |
|-------|-------------|----------|
| [ps-image](#ps-image) | Image processing | resize, convert, info |
| [ps-pptx](#ps-pptx) | PowerPoint files | create, text, info, list, themes |
| [ps-pdf](#ps-pdf) | PDF files | create (text/markdown/JSON), text, info |
| [ps-qr](#ps-qr) | QR code generation | generate |
| [ps-archive](#ps-archive) | Zip archives | zip, unzip, list |
| [ps-hash](#ps-hash) | File checksums | calc, verify |
| [ps-audio](#ps-audio) | Audio metadata | info, set, art, clear |
| [ps-video](#ps-video) | Video metadata | info, formats |

---

### ps-image

Image processing without dependencies - resize, convert formats, extract metadata.

```bash
# Resize
ps-image resize photo.jpg --width 800
ps-image resize photo.jpg --width 800 --height 600 --fit cover

# Convert
ps-image convert image.png --format jpg --quality 90

# Info
ps-image info photo.jpg --json
```

**Supported Formats:** PNG, JPEG, GIF, BMP, TIFF, WebP (read only)

---

### ps-pptx

Read and create PowerPoint presentations with themes, backgrounds, images, shapes, and full positioning control.

```bash
# Extract text
ps-pptx text presentation.pptx

# Create simple presentation
ps-pptx create notes.txt -o deck.pptx

# Create with theme and background
ps-pptx create notes.txt -o deck.pptx --theme teal --background 023047

# Add images to slides
ps-pptx create notes.txt -o deck.pptx --image slide:5,path:qr.png,x:6.5,y:1,w:3,h:3

# Full control with JSON input
ps-pptx create presentation.json -o deck.pptx --json

# List available themes
ps-pptx themes
```

**Themes:** default, dark, light, teal, coral

**JSON Input** for full control over positioning, styling, and elements:
```json
{
  "options": {"title": "My Deck", "theme": "teal"},
  "slides": [
    {"title": "Welcome", "background": {"color": "023047"}},
    {
      "elements": [
        {"type": "text", "text": {"content": "Hello", "position": {"x": 1, "y": 2, "width": 8, "height": 1}, "style": {"fontSize": 48, "bold": true}}},
        {"type": "image", "image": {"path": "qr.png", "position": {"x": 6.5, "y": 1, "width": 3, "height": 3}}},
        {"type": "shape", "shape": {"type": "rectangle", "position": {"x": 0.5, "y": 0.5, "width": 9, "height": 0.1}, "fillColor": "FFB703"}}
      ]
    }
  ]
}
```

---

### ps-pdf

Extract text from PDFs and create PDFs from text, Markdown, or structured JSON.

```bash
# Extract text
ps-pdf text document.pdf
ps-pdf text document.pdf --page 3

# Get info
ps-pdf info document.pdf --json

# Create PDF from text
ps-pdf create notes.txt -o notes.pdf --title "My Notes"

# Create from Markdown with page numbers
ps-pdf create readme.md -o readme.pdf --markdown --page-numbers

# Create with custom margins
ps-pdf create doc.txt -o doc.pdf --margin-top 25 --margin-bottom 25

# Full control with JSON input
ps-pdf create layout.json -o report.pdf --json
```

**Markdown Support:** Headings, bullet/numbered lists, code blocks, horizontal rules

**JSON Input** for full control over document layout:
```json
{
  "options": {
    "title": "My Document",
    "pageNumbers": true,
    "marginTop": 25
  },
  "markdown": "# Introduction\n\nThis is a paragraph.\n\n- Item 1\n- Item 2"
}
```

Or with explicit paragraphs:
```json
{
  "paragraphs": [
    { "type": "heading", "level": 1, "content": "Title" },
    { "type": "text", "content": "Paragraph text." },
    { "type": "list", "items": ["One", "Two"], "ordered": true },
    { "type": "code", "content": "console.log('hello');" }
  ]
}
```

---

### ps-qr

Generate QR codes from text, URLs, or data.

```bash
# Generate QR code
ps-qr generate "https://example.com" -o qrcode.png

# Larger size with high error correction
ps-qr generate "Hello World" -o hello.png --size 512 --level H

# ASCII art in terminal
ps-qr generate "https://example.com" --ascii

# WiFi QR code
ps-qr generate "WIFI:T:WPA;S:MyNetwork;P:MyPassword;;" -o wifi.png
```

---

### ps-archive

Create and extract zip archives.

```bash
# Create zip
ps-archive zip document.pdf
ps-archive zip photos/ -o vacation-photos.zip

# Extract
ps-archive unzip backup.zip -o ~/restored/

# List contents
ps-archive list backup.zip --json
```

---

### ps-hash

Calculate and verify file checksums.

```bash
# Calculate SHA256 (default)
ps-hash calc document.pdf

# Calculate MD5
ps-hash calc file.zip -a md5

# Verify hash
ps-hash verify download.iso abc123def456...
```

**Algorithms:** md5, sha1, sha256 (default), sha512

---

### ps-audio

Read and write metadata on audio files, extract and set album art.

```bash
# Show metadata
ps-audio info song.mp3
ps-audio info album/*.flac --json

# Set tags (MP3 only for writes)
ps-audio set song.mp3 --title "My Song" --artist "Artist Name"
ps-audio set song.mp3 --album "Album" --year 2024 --track 5

# Set album art
ps-audio set song.mp3 --art cover.jpg

# Extract album art
ps-audio art song.mp3 -o cover.jpg

# Clear tags
ps-audio clear song.mp3           # All tags
ps-audio clear song.mp3 --art     # Just album art
```

**Read Formats:** MP3, M4A, FLAC, OGG, WAV, AIFF
**Write Formats:** MP3 (ID3v2)

---

### ps-video

Read video metadata including duration, resolution, codecs, frame rate, and more.

```bash
# Show metadata
ps-video info video.mp4
ps-video info *.mp4 --json

# Detailed track information
ps-video info movie.mp4 --detailed

# List supported formats
ps-video formats
```

**Output:**
```
File: video.mp4
Format: MP4
Duration: 2:30.500 (150.50 seconds)
Resolution: 1920x1080
Video Codec: H.264/AVC
Frame Rate: 29.97 fps
Audio Codec: AAC
Audio Channels: stereo
Bitrate: 5432 kbps
Size: 98.45 MB
```

**Supported Formats:** MP4, M4V, MOV, 3GP

---

## Security

We take security seriously. All releases include:

### Vulnerability Scanning
- **govulncheck**: Go vulnerability database scanning on every build
- **Trivy**: Container and filesystem vulnerability scanning
- **CodeQL**: Static analysis for security issues
- **Dependency Review**: License and vulnerability checks on PRs

### Code Signing

All binaries are signed using [Sigstore cosign](https://docs.sigstore.dev/cosign/overview/) with keyless signing. This provides:
- Proof that binaries were built by our GitHub Actions
- Tamper detection if binaries are modified
- No private key management (uses OIDC identity)

**Verify a release:**
```bash
# Download binary and signatures
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64.sig
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/ps-image-linux-amd64.pem

# Verify
cosign verify-blob \
  --signature ps-image-linux-amd64.sig \
  --certificate ps-image-linux-amd64.pem \
  --certificate-identity-regexp "https://github.com/Barneyjm/portable-skills/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ps-image-linux-amd64
```

### Checksums

Each release includes a `checksums.txt` file with SHA256 hashes:
```bash
curl -LO https://github.com/Barneyjm/portable-skills/releases/latest/download/checksums.txt
sha256sum -c checksums.txt --ignore-missing
```

### SBOM

Software Bill of Materials (SBOM) in CycloneDX format is included with each release for supply chain transparency.

## Build Platforms

| OS      | amd64 | arm64 |
|---------|-------|-------|
| Linux   | ✓     | ✓     |
| macOS   | ✓     | ✓     |
| Windows | ✓     | -     |

All binaries are:
- Statically compiled (CGO_ENABLED=0)
- Stripped for smaller size
- Self-contained with no runtime dependencies

## For Agent Developers

Each skill includes a `SKILL.md` file that agents can read to understand capabilities:

```bash
# Location of skill manifests
skills/
├── archive/
│   └── SKILL.md
├── audio/
│   └── SKILL.md
├── hash/
│   └── SKILL.md
├── image/
│   └── SKILL.md
├── pdf/
│   └── SKILL.md
├── pptx/
│   └── SKILL.md
├── qr/
│   └── SKILL.md
└── video/
    └── SKILL.md
```

The SKILL.md files follow a consistent format with YAML frontmatter and markdown documentation that describes available commands, flags, and examples.

## Development

### Prerequisites

- Go 1.24+

### Building

```bash
# Build all skills for current platform
go build ./cmd/...

# Build a specific skill
go build -o ps-pptx ./cmd/ps-pptx

# Run tests
go test -v ./...
```

### Project Structure

```
portable-skills/
├── cmd/
│   ├── ps-archive/         # Archive skill CLI
│   ├── ps-audio/           # Audio skill CLI
│   ├── ps-hash/            # Hash skill CLI
│   ├── ps-image/           # Image skill CLI
│   ├── ps-pdf/             # PDF skill CLI
│   ├── ps-pptx/            # PowerPoint skill CLI
│   ├── ps-qr/              # QR code skill CLI
│   └── ps-video/           # Video skill CLI
├── pkg/
│   ├── archive/            # Archive processing library
│   ├── audio/              # Audio metadata library
│   ├── hash/               # Hashing library
│   ├── image/              # Image processing library
│   ├── pdf/                # PDF processing library
│   ├── pptx/               # PowerPoint processing library
│   ├── qr/                 # QR code library
│   ├── skill/              # Skill protocol helpers
│   └── video/              # Video metadata library
├── skills/                 # SKILL.md files for each skill
├── dist/
│   ├── homebrew/           # Homebrew formula
│   └── npm/                # npm wrapper package
└── .github/
    └── workflows/          # CI/CD pipelines
```

## Contributing

Contributions are welcome! Please read our contributing guidelines and ensure your code passes all checks:

```bash
# Run linter
golangci-lint run

# Run tests
go test -v -race ./...

# Run vulnerability check
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.

---

*"It should just work."*
