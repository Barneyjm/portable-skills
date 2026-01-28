---
name: portable-audio
description: Read and write audio file metadata, extract and set album art. Supports MP3, M4A, FLAC, OGG.
---

# Audio Metadata Skill

Read and write metadata from audio files, extract and set album art. No dependencies required.

## Supported Formats

**Read support:**
- MP3 (ID3v1, ID3v2)
- M4A/MP4/AAC
- FLAC
- OGG Vorbis
- WAV
- AIFF

**Write support:**
- MP3 (ID3v2 tags) - full read/write support

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

### set
Set or update tags on an audio file.

```bash
ps-audio set <file> [flags]
```

**Arguments:**
- `file` - Audio file to update (MP3 only for writes)

**Flags:**
- `--title` - Track title
- `--artist` - Artist name
- `--album` - Album name
- `--album-artist` - Album artist
- `--year` - Release year
- `--track` - Track number
- `--track-total` - Total tracks
- `--disc` - Disc number
- `--disc-total` - Total discs
- `--genre` - Genre
- `--composer` - Composer
- `--comment` - Comment
- `--art` - Path to album art image (JPG, PNG)
- `--json` - Output result as JSON

**Examples:**
```bash
# Set basic metadata
ps-audio set song.mp3 --title "My Song" --artist "Artist Name"

# Set album and track info
ps-audio set song.mp3 --album "Album Name" --year 2024 --track 5 --track-total 12

# Set album art
ps-audio set song.mp3 --art cover.jpg

# Update multiple fields
ps-audio set song.mp3 --genre "Rock" --comment "Great track" --composer "John Doe"

# Get JSON output of changes
ps-audio set song.mp3 --title "New Title" --json
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

### clear
Clear tags from an audio file.

```bash
ps-audio clear <file> [--art]
```

**Arguments:**
- `file` - Audio file to clear (MP3 only)

**Flags:**
- `--art` - Remove only album art, keep other tags

**Examples:**
```bash
# Clear all tags
ps-audio clear song.mp3

# Remove only album art
ps-audio clear song.mp3 --art
```

## Common Use Cases

1. **Organize music**: Read tags to organize files
2. **Fix metadata**: Update incorrect artist/album/title information
3. **Set album art**: Add or replace cover images
4. **Batch tagging**: Script tag updates for multiple files
5. **Extract art**: Get album covers for playlists or display
6. **Clean files**: Remove unwanted tags or album art

## Scripting Examples

Batch update artist for all MP3s in a folder:
```bash
for f in *.mp3; do
  ps-audio set "$f" --artist "Correct Artist"
done
```

Extract all album art from a folder:
```bash
for f in *.mp3; do
  ps-audio art "$f" -o "covers/$(basename "$f" .mp3).jpg"
done
```

Get metadata as JSON for processing:
```bash
ps-audio info *.mp3 --json | jq '.[] | select(.year == 2024)'
```

## Limitations

- Write support is currently MP3-only (FLAC, M4A, OGG are read-only)
- Album art should be JPG or PNG format for best compatibility
- Very large album art images may increase file size significantly
