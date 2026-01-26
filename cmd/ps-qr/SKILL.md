---
name: portable-qr
description: Generate QR codes from text or URLs. Output as PNG or ASCII art.
version: 1.0.0
binary: ps-qr
---

# QR Code Skill

Generate QR codes from any text, URL, or data. No dependencies required.

## Available Commands

### generate
Create a QR code from content.

```bash
ps-qr generate <content> [--output <file>] [--size <px>] [--level <L|M|Q|H>] [--ascii]
```

**Arguments:**
- `content` - Text, URL, or data to encode

**Flags:**
- `--output, -o` - Output file path (default: stdout)
- `--size, -s` - Image size in pixels (default: 256)
- `--level, -l` - Error correction level: L, M, Q, H (default: M)
- `--invert` - Invert colors (white on black)
- `--ascii, -a` - Output as ASCII art for terminal display

**Examples:**
```bash
# Generate QR code for URL, save to file
ps-qr generate "https://example.com" -o qrcode.png

# Generate larger QR code with high error correction
ps-qr generate "Hello World" -o hello.png --size 512 --level H

# Display QR code in terminal as ASCII
ps-qr generate "https://example.com" --ascii

# Pipe PNG to another program
ps-qr generate "data" | feh -

# Inverted colors (for dark backgrounds)
ps-qr generate "text" -o inverted.png --invert
```

## Error Correction Levels

| Level | Recovery | Best For |
|-------|----------|----------|
| L | 7% | Clean environments, maximum data |
| M | 15% | General use (default) |
| Q | 25% | Printed materials |
| H | 30% | Harsh environments, logos |

Higher levels allow more damage recovery but reduce data capacity.

## Common Use Cases

1. **Share URLs**: Generate QR for websites, app links
2. **WiFi sharing**: Encode WiFi credentials
3. **Contact info**: vCard format QR codes
4. **Plain text**: Short messages, codes
5. **Terminal preview**: ASCII mode for quick checks

## WiFi QR Code Format

```bash
ps-qr generate "WIFI:T:WPA;S:MyNetwork;P:MyPassword;;" -o wifi.png
```

## vCard QR Code Format

```bash
ps-qr generate "BEGIN:VCARD
VERSION:3.0
N:Doe;John
TEL:+1234567890
EMAIL:john@example.com
END:VCARD" -o contact.png
```
