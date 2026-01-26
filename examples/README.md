# Portable Skills Examples

This directory contains example outputs from each portable skill.

## Files

| File | Created By | Description |
|------|------------|-------------|
| `example.pdf` | `ps-pdf` | PDF document created from text |
| `qr-code.png` | `ps-qr` | QR code linking to this repo |
| `example.pdf.sha256.json` | `ps-hash` | SHA256 checksum of the PDF |
| `demo-archive.zip` | `ps-archive` | Zip containing the PDF and QR code |

## How These Were Created

```bash
# Create a PDF from text
echo "Hello World" | ps-pdf create - -o example.pdf --title "Example"

# Generate a QR code
ps-qr generate "https://github.com/Barneyjm/portable-skills" -o qr-code.png

# Calculate checksum
ps-hash calc example.pdf --json > example.pdf.sha256.json

# Create an archive
ps-archive zip example.pdf qr-code.png -o demo-archive.zip
```

## Try It Yourself

Install all skills:
```bash
npm install -g portable-skills
ps-image install-skill
ps-hash install-skill
ps-qr install-skill
ps-archive install-skill
ps-audio install-skill
ps-pdf install-skill
```

Then use `/image`, `/hash`, `/qr`, `/archive`, `/audio`, or `/pdf` in Claude Code.
