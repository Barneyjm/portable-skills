# Claude Context for portable-skills

This file provides context for AI agents working on this codebase.

## Project Overview

**portable-skills** provides pre-compiled Go binaries that implement common skill capabilities for AI agents. The goal is zero-dependency, cross-platform tools that work without requiring users to have Python, Node, or build toolchains installed.

**Core thesis**: Agent skills should work for normal humans on normal computers, not just developers.

## Architecture

```
portable-skills/
├── cmd/ps-image/main.go      # CLI entry point (uses cobra)
├── pkg/
│   ├── image/                # Core image processing (pure Go, no CGO)
│   │   ├── resize.go         # Resize with fit modes: contain, cover, fill
│   │   ├── convert.go        # Format conversion
│   │   ├── metadata.go       # EXIF extraction
│   │   └── io.go             # Load/save/encode
│   └── skill/                # Skill protocol helpers
│       ├── input.go          # Stdin/file input handling
│       ├── output.go         # JSON/structured output
│       └── manifest.go       # SKILL.md generation
├── skills/image/SKILL.md     # Agent-readable skill definition
├── dist/
│   ├── homebrew/Formula/     # Homebrew formula
│   └── npm/                  # npm wrapper package
└── .github/workflows/        # CI/CD
    ├── build.yml             # Test, build matrix, lint
    ├── release.yml           # Cross-compile + cosign signing
    └── security.yml          # govulncheck, trivy, CodeQL, SBOM
```

## Key Dependencies

All pure Go (no CGO):
- `github.com/disintegration/imaging` - Image resize/convert
- `github.com/rwcarlsen/goexif` - EXIF extraction
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/image` - Additional format support (webp, bmp, tiff)
- `github.com/bogem/id3v2/v2` - MP3 ID3v2 tag reading/writing
- `github.com/dhowden/tag` - Audio metadata reading (multi-format)
- `github.com/abema/go-mp4` - MP4/MOV video container parsing

## Getting Started

```bash
# Build for current platform
go build -o ps-image ./cmd/ps-image

# Run tests
go test -v ./...

# Run linter (CI runs this)
golangci-lint run --timeout=5m
```

## Installing as a Skill

Skills are installed to `~/.claude/skills/<name>/`:

```bash
mkdir -p ~/.claude/skills/image
go build -o ~/.claude/skills/image/ps-image ./cmd/ps-image
cp skills/image/SKILL.md ~/.claude/skills/image/
```

Then invoke with `/image` in Claude Code.

## CLI Usage

```bash
# Resize
ps-image resize photo.jpg --width 800
ps-image resize photo.jpg --width 800 --height 600 --fit cover

# Convert
ps-image convert image.png --format jpg --quality 90

# Info
ps-image info photo.jpg --json

# Stdin/stdout
cat input.png | ps-image resize - --width 400 -o - > output.png
```

## Important Notes

1. **No `-h` for height** - Uses `--height` only (conflicts with `--help`)
2. **Stdin requires `-`** - Pass `-` as input argument when piping
3. **Stdin format auto-detected** - Magic bytes detect PNG, JPEG, GIF, BMP, WebP, TIFF
4. **Resize has `--format` flag** - Control stdout output format (defaults to input format)
5. **WebP decode only** - Can read WebP but not write (falls back to PNG)
6. **go.mod version** - Uses Go 1.24+, workflows use `go-version-file: go.mod`

## CI/CD

- **Lint**: golangci-lint with errcheck enabled (all errors must be handled)
- **Security**: govulncheck, Trivy, CodeQL, dependency review
- **Releases**: Cosign keyless signing, SHA256 checksums, SBOM generation
- **Platforms**: linux/darwin (amd64, arm64), windows (amd64)

## Documentation Requirements

**IMPORTANT:** When making changes to skill functionality, always update:

1. **SKILL.md** - Update the corresponding `skills/<name>/SKILL.md` file with new commands, flags, and examples
2. **README.md** - Update the main README with relevant changes to the skill's section
3. **CLAUDE.md** - Update this file if there are architectural changes or new patterns

This ensures users and agents can discover and use new features correctly.

## Future Work

- `ps-office` - Read/write docx, xlsx
- `ps-video` - Thumbnail generation (requires external codec)
- Extend write support for FLAC, M4A in ps-audio

## Common Tasks

### Adding a new command to ps-image

1. Add function to `pkg/image/`
2. Add command in `cmd/ps-image/main.go` (follow existing pattern)
3. Update `skills/image/SKILL.md`
4. Add tests

### Adding a new skill binary

1. Create `cmd/ps-<name>/main.go`
2. Create `pkg/<name>/` with core logic
3. Create `skills/<name>/SKILL.md`
4. Add to build matrix in `.github/workflows/build.yml`
5. Add formula to `dist/homebrew/Formula/`
