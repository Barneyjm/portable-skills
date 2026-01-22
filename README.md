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
brew install Barneyjm/tap/ps-image
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

### ps-image

Image processing without dependencies - resize, convert formats, extract metadata.

#### Commands

**resize** - Resize an image
```bash
ps-image resize photo.jpg --width 800
ps-image resize photo.jpg --width 800 --height 600 --fit cover
ps-image resize photo.jpg -w 400 -o thumbnail.jpg
```

**convert** - Convert between formats
```bash
ps-image convert image.png --format jpg
ps-image convert image.bmp --format png --output result.png
```

**info** - Extract metadata
```bash
ps-image info photo.jpg
ps-image info photo.jpg --json
```

#### Supported Formats

| Format | Read | Write |
|--------|------|-------|
| PNG    | ✓    | ✓     |
| JPEG   | ✓    | ✓     |
| GIF    | ✓    | ✓     |
| BMP    | ✓    | ✓     |
| TIFF   | ✓    | ✓     |
| WebP   | ✓    | ✗     |

#### Stdin/Stdout Support

```bash
# Pipe from curl
curl -s https://example.com/image.png | ps-image resize - --width 400 -o - > thumbnail.png

# Chain commands
cat large.png | ps-image resize - --width 800 -o - | ps-image convert - --format jpg -o - > result.jpg
```

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
├── image/
│   └── SKILL.md    # Agent-readable skill definition
├── pdf/
│   └── SKILL.md
└── office/
    └── SKILL.md
```

The SKILL.md files follow a consistent format with YAML frontmatter and markdown documentation that describes available commands, flags, and examples.

## Development

### Prerequisites

- Go 1.22+

### Building

```bash
# Build for current platform
go build -o ps-image ./cmd/ps-image

# Build for all platforms
./scripts/build-all.sh

# Run tests
go test -v ./...
```

### Project Structure

```
portable-skills/
├── cmd/
│   └── ps-image/           # CLI entry point
├── pkg/
│   ├── image/              # Core image processing library
│   └── skill/              # Skill protocol helpers
├── skills/
│   └── image/
│       └── SKILL.md        # Agent-readable skill definition
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
