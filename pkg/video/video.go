// Package video provides video metadata reading utilities.
package video

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abema/go-mp4"
)

// Metadata represents video file metadata.
type Metadata struct {
	File           string        `json:"file"`
	Format         string        `json:"format"`
	Duration       float64       `json:"duration_seconds"`
	DurationStr    string        `json:"duration"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	Resolution     string        `json:"resolution"`
	VideoCodec     string        `json:"video_codec,omitempty"`
	AudioCodec     string        `json:"audio_codec,omitempty"`
	FrameRate      float64       `json:"frame_rate,omitempty"`
	VideoBitrate   int64         `json:"video_bitrate_kbps,omitempty"`
	AudioBitrate   int64         `json:"audio_bitrate_kbps,omitempty"`
	TotalBitrate   int64         `json:"total_bitrate_kbps,omitempty"`
	AudioChannels  int           `json:"audio_channels,omitempty"`
	AudioSampleRate int          `json:"audio_sample_rate,omitempty"`
	CreationTime   *time.Time    `json:"creation_time,omitempty"`
	ModificationTime *time.Time  `json:"modification_time,omitempty"`
	SizeBytes      int64         `json:"size_bytes"`
	HasVideo       bool          `json:"has_video"`
	HasAudio       bool          `json:"has_audio"`
}

// Track represents a single track in the video file.
type Track struct {
	ID       uint32 `json:"id"`
	Type     string `json:"type"` // "video", "audio", "subtitle", "other"
	Codec    string `json:"codec,omitempty"`
	Duration float64 `json:"duration_seconds,omitempty"`
	Language string `json:"language,omitempty"`
	// Video-specific
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	FrameRate float64 `json:"frame_rate,omitempty"`
	// Audio-specific
	Channels   int `json:"channels,omitempty"`
	SampleRate int `json:"sample_rate,omitempty"`
}

// DetailedInfo contains detailed information about a video file.
type DetailedInfo struct {
	Metadata
	Tracks []Track `json:"tracks"`
}

// Info reads metadata from a video file.
func Info(path string) (*Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	format := detectFormat(path)
	if format == "" {
		return nil, fmt.Errorf("unsupported video format: %s", filepath.Ext(path))
	}

	meta := &Metadata{
		File:      filepath.Base(path),
		Format:    format,
		SizeBytes: stat.Size(),
	}

	// Parse MP4/MOV container
	if err := parseMP4(f, meta); err != nil {
		return nil, fmt.Errorf("failed to parse video: %w", err)
	}

	// Calculate total bitrate if we have duration
	if meta.Duration > 0 {
		meta.TotalBitrate = int64(float64(stat.Size()*8/1000) / meta.Duration)
	}

	return meta, nil
}

// DetailedInfo reads detailed metadata including all tracks.
func GetDetailedInfo(path string) (*DetailedInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	format := detectFormat(path)
	if format == "" {
		return nil, fmt.Errorf("unsupported video format: %s", filepath.Ext(path))
	}

	info := &DetailedInfo{
		Metadata: Metadata{
			File:      filepath.Base(path),
			Format:    format,
			SizeBytes: stat.Size(),
		},
		Tracks: []Track{},
	}

	if err := parseMP4Detailed(f, info); err != nil {
		return nil, fmt.Errorf("failed to parse video: %w", err)
	}

	// Calculate total bitrate
	if info.Duration > 0 {
		info.TotalBitrate = int64(float64(stat.Size()*8/1000) / info.Duration)
	}

	return info, nil
}

func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".m4v":
		return "MP4"
	case ".mov":
		return "MOV"
	case ".m4a":
		return "M4A"
	case ".3gp":
		return "3GP"
	case ".3g2":
		return "3G2"
	default:
		return ""
	}
}

func parseMP4(r io.ReadSeeker, meta *Metadata) error {
	// Read all boxes
	boxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov()})
	if err != nil {
		return fmt.Errorf("failed to find moov box: %w", err)
	}

	if len(boxes) == 0 {
		return fmt.Errorf("no moov box found")
	}

	// Reset to read from moov
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Get movie header for global timescale and duration
	mvhdBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeMvhd()})
	if err == nil && len(mvhdBoxes) > 0 {
		if mvhd, ok := mvhdBoxes[0].Payload.(*mp4.Mvhd); ok {
			timescale := mvhd.Timescale
			var duration uint64
			if mvhd.GetVersion() == 0 {
				duration = uint64(mvhd.DurationV0)
			} else {
				duration = mvhd.DurationV1
			}
			if timescale > 0 {
				meta.Duration = float64(duration) / float64(timescale)
				meta.DurationStr = formatDuration(meta.Duration)
			}

			// Creation and modification times
			// MP4 epoch is 1904-01-01, we need to convert to Unix epoch
			mp4Epoch := time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
			if mvhd.GetVersion() == 0 {
				if mvhd.CreationTimeV0 > 0 {
					ct := mp4Epoch.Add(time.Duration(mvhd.CreationTimeV0) * time.Second)
					meta.CreationTime = &ct
				}
				if mvhd.ModificationTimeV0 > 0 {
					mt := mp4Epoch.Add(time.Duration(mvhd.ModificationTimeV0) * time.Second)
					meta.ModificationTime = &mt
				}
			} else {
				if mvhd.CreationTimeV1 > 0 {
					ct := mp4Epoch.Add(time.Duration(mvhd.CreationTimeV1) * time.Second)
					meta.CreationTime = &ct
				}
				if mvhd.ModificationTimeV1 > 0 {
					mt := mp4Epoch.Add(time.Duration(mvhd.ModificationTimeV1) * time.Second)
					meta.ModificationTime = &mt
				}
			}
		}
	}

	// Parse tracks
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Get track headers
	tkhdBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeTkhd()})
	if err == nil {
		for _, box := range tkhdBoxes {
			if tkhd, ok := box.Payload.(*mp4.Tkhd); ok {
				width := int(tkhd.GetWidth())
				height := int(tkhd.GetHeight())
				if width > 0 && height > 0 {
					meta.Width = width
					meta.Height = height
					meta.Resolution = fmt.Sprintf("%dx%d", width, height)
					meta.HasVideo = true
				}
			}
		}
	}

	// Get video codec info
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}
	meta.VideoCodec = findVideoCodec(r)

	// Get audio codec info
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}
	audioCodec, channels, sampleRate := findAudioCodec(r)
	if audioCodec != "" {
		meta.AudioCodec = audioCodec
		meta.AudioChannels = channels
		meta.AudioSampleRate = sampleRate
		meta.HasAudio = true
	}

	// Get frame rate from stts (sample timing)
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}
	meta.FrameRate = calculateFrameRate(r, meta.Duration)

	return nil
}

func parseMP4Detailed(r io.ReadSeeker, info *DetailedInfo) error {
	// First get basic metadata
	if err := parseMP4(r, &info.Metadata); err != nil {
		return err
	}

	// Reset and get track info
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Find all trak boxes
	trakBoxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak()})
	if err != nil {
		return nil // Not an error, just no tracks
	}

	for _, trakBox := range trakBoxes {
		track := Track{}

		// Get track header
		if _, err := r.Seek(int64(trakBox.Offset), io.SeekStart); err != nil {
			continue
		}

		tkhdBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeTrak(), mp4.BoxTypeTkhd()})
		if err != nil || len(tkhdBoxes) == 0 {
			continue
		}

		if tkhd, ok := tkhdBoxes[0].Payload.(*mp4.Tkhd); ok {
			track.ID = tkhd.TrackID
			width := int(tkhd.GetWidth())
			height := int(tkhd.GetHeight())
			if width > 0 && height > 0 {
				track.Type = "video"
				track.Width = width
				track.Height = height
			}
		}

		// Determine track type from handler
		if _, err := r.Seek(int64(trakBox.Offset), io.SeekStart); err != nil {
			continue
		}

		hdlrBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeHdlr()})
		if err == nil && len(hdlrBoxes) > 0 {
			if hdlr, ok := hdlrBoxes[0].Payload.(*mp4.Hdlr); ok {
				switch hdlr.HandlerType {
				case [4]byte{'v', 'i', 'd', 'e'}:
					track.Type = "video"
				case [4]byte{'s', 'o', 'u', 'n'}:
					track.Type = "audio"
				case [4]byte{'s', 'b', 't', 'l'}, [4]byte{'t', 'e', 'x', 't'}:
					track.Type = "subtitle"
				default:
					track.Type = "other"
				}
			}
		}

		if track.Type == "" {
			track.Type = "other"
		}

		info.Tracks = append(info.Tracks, track)
	}

	return nil
}

func findVideoCodec(r io.ReadSeeker) string {
	// Check for AVC/H.264
	avc1Boxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAvc1()})
	if err == nil && len(avc1Boxes) > 0 {
		return "H.264/AVC"
	}

	// Reset and check for HEVC/H.265
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return ""
	}
	hvc1Boxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeHvc1()})
	if err == nil && len(hvc1Boxes) > 0 {
		return "H.265/HEVC"
	}

	// Reset and check for VP9
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return ""
	}
	vp09Boxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeVp09()})
	if err == nil && len(vp09Boxes) > 0 {
		return "VP9"
	}

	// Reset and check for AV1
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return ""
	}
	av01Boxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAv01()})
	if err == nil && len(av01Boxes) > 0 {
		return "AV1"
	}

	return ""
}

func findAudioCodec(r io.ReadSeeker) (codec string, channels int, sampleRate int) {
	// Check for AAC (mp4a)
	mp4aBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeMp4a()})
	if err == nil && len(mp4aBoxes) > 0 {
		if mp4a, ok := mp4aBoxes[0].Payload.(*mp4.AudioSampleEntry); ok {
			channels = int(mp4a.ChannelCount)
			sampleRate = int(mp4a.SampleRate >> 16)
		}
		return "AAC", channels, sampleRate
	}

	// Reset and check for AC-3
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", 0, 0
	}
	ac3Boxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAC3()})
	if err == nil && len(ac3Boxes) > 0 {
		return "AC-3", 0, 0
	}

	// Reset and check for Opus
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", 0, 0
	}
	opusBoxes, err := mp4.ExtractBox(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeOpus()})
	if err == nil && len(opusBoxes) > 0 {
		return "Opus", 0, 0
	}

	return "", 0, 0
}

func calculateFrameRate(r io.ReadSeeker, duration float64) float64 {
	if duration <= 0 {
		return 0
	}

	// Get stts box for sample timing
	sttsBoxes, err := mp4.ExtractBoxWithPayload(r, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(), mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStts()})
	if err != nil || len(sttsBoxes) == 0 {
		return 0
	}

	stts, ok := sttsBoxes[0].Payload.(*mp4.Stts)
	if !ok {
		return 0
	}

	// Count total samples
	var totalSamples uint32
	for _, entry := range stts.Entries {
		totalSamples += entry.SampleCount
	}

	if totalSamples == 0 {
		return 0
	}

	// Frame rate = total samples / duration
	return float64(totalSamples) / duration
}

func formatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, secs, millis)
	}
	return fmt.Sprintf("%d:%02d.%03d", minutes, secs, millis)
}

// IsSupportedFormat checks if the file extension is a supported video format.
func IsSupportedFormat(path string) bool {
	return detectFormat(path) != ""
}

// SupportedFormats returns a list of supported video formats.
func SupportedFormats() []string {
	return []string{"MP4", "M4V", "MOV", "M4A", "3GP", "3G2"}
}
