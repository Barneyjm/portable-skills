---
name: portable-audio
description: Read audio file metadata and extract album art. Supports MP3, M4A, FLAC, OGG.
version: 1.0.0
binary: ps-audio
---

# Audio Metadata Skill

Read metadata from audio files and extract album art. No dependencies required.

## Supported Formats

- MP3 (ID3v1, ID3v2)
- M4A/MP4/AAC
- FLAC
- OGG Vorbis
- WAV
- AIFF

## Available Commands

### info
Display metadata from audio files.

```bash
ps-audio info <file> [file...] [--json]
```

**Arguments:**
- `file` - Audio file(s) to read

**Flags:**
- `--json` - Output as JSON

**Examples:**
```bash
# Show metadata for a song
ps-audio info song.mp3

# Show metadata for multiple files
ps-audio info *.mp3

# JSON output
ps-audio info album/*.flac --json
```

**Output:**
```
File: song.mp3
Format: ID3v2.4 (MP3)
Title: Song Name
Artist: Artist Name
Album: Album Name
Year: 2024
Genre: Rock
Track: 3/12
Album Art: Yes
Size: 5.2 MB
```

### art
Extract embedded album art from an audio file.

```bash
ps-audio art <audio-file> [--output <image-file>]
```

**Arguments:**
- `audio-file` - Audio file to extract art from

**Flags:**
- `--output, -o` - Output file path (default: same name as input with .jpg)

**Examples:**
```bash
# Extract album art
ps-audio art song.mp3

# Save with custom name
ps-audio art song.mp3 -o cover.jpg
```

## Common Use Cases

1. **Organize music**: Read tags to organize files
2. **Check metadata**: Verify songs are properly tagged
3. **Extract art**: Get album covers for playlists
4. **Batch info**: Get metadata for multiple files as JSON
