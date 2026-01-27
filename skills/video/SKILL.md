---
name: portable-video
description: Read video file metadata including duration, resolution, codecs, and more. Supports MP4, MOV, M4V.
version: 1.0.0
binary: ps-video
---

# Video Metadata Skill

Read metadata from video files without external dependencies. Extract duration, resolution, codecs, frame rate, and more.

## Supported Formats

- MP4 (.mp4, .m4v)
- MOV (.mov)
- M4A (.m4a) - audio-only container
- 3GP (.3gp, .3g2)

## Available Commands

### info
Display metadata from video files.

```bash
ps-video info <file> [file...] [--json] [--detailed]
```

**Arguments:**
- `file` - Video file(s) to read

**Flags:**
- `--json` - Output as JSON
- `--detailed` - Show detailed track information

**Examples:**
```bash
# Show metadata for a video
ps-video info video.mp4

# Show metadata for multiple files
ps-video info *.mp4

# JSON output
ps-video info movie.mp4 --json

# Detailed track information
ps-video info video.mp4 --detailed
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
Audio Sample Rate: 48000 Hz
Bitrate: 5432 kbps
Created: 2024-03-15 14:30:00
Size: 98.45 MB
```

**JSON Output:**
```json
{
  "file": "video.mp4",
  "format": "MP4",
  "duration_seconds": 150.5,
  "duration": "2:30.500",
  "width": 1920,
  "height": 1080,
  "resolution": "1920x1080",
  "video_codec": "H.264/AVC",
  "audio_codec": "AAC",
  "frame_rate": 29.97,
  "total_bitrate_kbps": 5432,
  "audio_channels": 2,
  "audio_sample_rate": 48000,
  "creation_time": "2024-03-15T14:30:00Z",
  "size_bytes": 103234567,
  "has_video": true,
  "has_audio": true
}
```

### formats
List supported video formats.

```bash
ps-video formats
```

**Output:**
```
Supported video formats:
  - MP4
  - M4V
  - MOV
  - M4A
  - 3GP
  - 3G2
```

## Metadata Fields

| Field | Description |
|-------|-------------|
| duration | Video length in seconds and HH:MM:SS.mmm format |
| resolution | Width x Height in pixels |
| video_codec | Video codec (H.264/AVC, H.265/HEVC, VP9, AV1) |
| audio_codec | Audio codec (AAC, AC-3, Opus) |
| frame_rate | Frames per second |
| bitrate | Total bitrate in kbps |
| audio_channels | Number of audio channels |
| audio_sample_rate | Audio sample rate in Hz |
| creation_time | When the video was created |
| modification_time | When the video was last modified |

## Common Use Cases

1. **Verify video specs**: Check resolution, codec, and duration before processing
2. **Organize media**: Sort videos by duration, resolution, or creation date
3. **Quality control**: Verify bitrate and frame rate meet requirements
4. **Media inventory**: Generate JSON metadata for video libraries
5. **Automation**: Script video processing based on metadata

## Scripting Examples

Get duration of all videos:
```bash
for f in *.mp4; do
  echo "$f: $(ps-video info "$f" --json | jq -r '.duration')"
done
```

Find videos with specific resolution:
```bash
ps-video info *.mp4 --json | jq '.[] | select(.width == 1920)'
```

Generate video inventory:
```bash
ps-video info videos/*.mp4 --json > video_inventory.json
```

## Limitations

- Read-only (no metadata writing)
- Limited to MP4-based containers (no MKV, AVI, WebM)
- No thumbnail extraction (requires external codecs)
- Frame rate calculated from sample count (may vary for VFR content)
