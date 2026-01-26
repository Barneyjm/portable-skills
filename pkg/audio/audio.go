// Package audio provides audio metadata reading utilities.
package audio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
)

// Metadata represents audio file metadata.
type Metadata struct {
	File        string  `json:"file"`
	Format      string  `json:"format"`
	FileType    string  `json:"file_type"`
	Title       string  `json:"title,omitempty"`
	Artist      string  `json:"artist,omitempty"`
	Album       string  `json:"album,omitempty"`
	AlbumArtist string  `json:"album_artist,omitempty"`
	Composer    string  `json:"composer,omitempty"`
	Genre       string  `json:"genre,omitempty"`
	Year        int     `json:"year,omitempty"`
	Track       int     `json:"track,omitempty"`
	TrackTotal  int     `json:"track_total,omitempty"`
	Disc        int     `json:"disc,omitempty"`
	DiscTotal   int     `json:"disc_total,omitempty"`
	Duration    float64 `json:"duration_seconds,omitempty"`
	Lyrics      string  `json:"lyrics,omitempty"`
	Comment     string  `json:"comment,omitempty"`
	HasPicture  bool    `json:"has_picture"`
	SizeBytes   int64   `json:"size_bytes"`
}

// Info reads metadata from an audio file.
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

	m, err := tag.ReadFrom(f)
	if err != nil {
		// Return basic info even if tags can't be read
		return &Metadata{
			File:      filepath.Base(path),
			FileType:  detectFileType(path),
			SizeBytes: stat.Size(),
		}, nil
	}

	meta := &Metadata{
		File:        filepath.Base(path),
		Format:      string(m.Format()),
		FileType:    string(m.FileType()),
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		Composer:    m.Composer(),
		Genre:       m.Genre(),
		Year:        m.Year(),
		Comment:     m.Comment(),
		HasPicture:  m.Picture() != nil,
		SizeBytes:   stat.Size(),
	}

	// Get track/disc info
	track, trackTotal := m.Track()
	meta.Track = track
	meta.TrackTotal = trackTotal

	disc, discTotal := m.Disc()
	meta.Disc = disc
	meta.DiscTotal = discTotal

	// Try to get lyrics if available
	if raw := m.Raw(); raw != nil {
		if lyrics, ok := raw["lyrics"].(string); ok {
			meta.Lyrics = lyrics
		}
	}

	return meta, nil
}

func detectFileType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".mp3":
		return "MP3"
	case ".m4a", ".m4b", ".m4p", ".mp4":
		return "M4A/MP4"
	case ".flac":
		return "FLAC"
	case ".ogg", ".oga":
		return "OGG"
	case ".wav":
		return "WAV"
	case ".wma":
		return "WMA"
	case ".aiff", ".aif":
		return "AIFF"
	default:
		return "Unknown"
	}
}

// ExtractPicture extracts album art from an audio file.
func ExtractPicture(audioPath, outputPath string) error {
	f, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return fmt.Errorf("failed to read tags: %w", err)
	}

	pic := m.Picture()
	if pic == nil {
		return fmt.Errorf("no album art found")
	}

	if err := os.WriteFile(outputPath, pic.Data, 0644); err != nil {
		return fmt.Errorf("failed to write picture: %w", err)
	}

	return nil
}

// PictureInfo returns information about embedded album art.
type PictureInfo struct {
	Type        string `json:"type"`
	MIMEType    string `json:"mime_type"`
	Description string `json:"description,omitempty"`
	SizeBytes   int    `json:"size_bytes"`
}

// GetPictureInfo returns info about embedded album art without extracting it.
func GetPictureInfo(path string) (*PictureInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	pic := m.Picture()
	if pic == nil {
		return nil, nil
	}

	return &PictureInfo{
		Type:        pic.Type,
		MIMEType:    pic.MIMEType,
		Description: pic.Description,
		SizeBytes:   len(pic.Data),
	}, nil
}
