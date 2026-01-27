// Package audio provides audio metadata reading and writing utilities.
package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
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

// TagUpdate represents a set of tags to update.
type TagUpdate struct {
	Title       *string `json:"title,omitempty"`
	Artist      *string `json:"artist,omitempty"`
	Album       *string `json:"album,omitempty"`
	AlbumArtist *string `json:"album_artist,omitempty"`
	Year        *int    `json:"year,omitempty"`
	Track       *int    `json:"track,omitempty"`
	TrackTotal  *int    `json:"track_total,omitempty"`
	Disc        *int    `json:"disc,omitempty"`
	DiscTotal   *int    `json:"disc_total,omitempty"`
	Genre       *string `json:"genre,omitempty"`
	Composer    *string `json:"composer,omitempty"`
	Comment     *string `json:"comment,omitempty"`
	ArtPath     *string `json:"art_path,omitempty"` // Path to album art image
}

// WriteResult contains the result of a tag write operation.
type WriteResult struct {
	Path    string   `json:"path"`
	Success bool     `json:"success"`
	Updated []string `json:"updated,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// WriteTags updates tags on an audio file.
// Currently supports MP3 files with ID3v2 tags.
func WriteTags(path string, updates TagUpdate) (*WriteResult, error) {
	// Check file extension to determine format
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		return writeMP3Tags(path, updates)
	case ".flac":
		return nil, fmt.Errorf("FLAC tag writing not yet supported (read-only)")
	case ".m4a", ".mp4", ".m4b", ".m4p":
		return nil, fmt.Errorf("M4A/MP4 tag writing not yet supported (read-only)")
	case ".ogg", ".oga":
		return nil, fmt.Errorf("OGG tag writing not yet supported (read-only)")
	default:
		return nil, fmt.Errorf("unsupported format for tag writing: %s", ext)
	}
}

// writeMP3Tags writes ID3v2 tags to an MP3 file.
func writeMP3Tags(path string, updates TagUpdate) (*WriteResult, error) {
	result := &WriteResult{
		Path:    path,
		Updated: []string{},
	}

	// Open existing tag or create new one
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = t.Close() }()

	// Update fields that are set
	if updates.Title != nil {
		t.SetTitle(*updates.Title)
		result.Updated = append(result.Updated, "title")
	}

	if updates.Artist != nil {
		t.SetArtist(*updates.Artist)
		result.Updated = append(result.Updated, "artist")
	}

	if updates.Album != nil {
		t.SetAlbum(*updates.Album)
		result.Updated = append(result.Updated, "album")
	}

	if updates.AlbumArtist != nil {
		// TPE2 frame for album artist
		t.AddTextFrame(t.CommonID("TPE2"), t.DefaultEncoding(), *updates.AlbumArtist)
		result.Updated = append(result.Updated, "album_artist")
	}

	if updates.Year != nil {
		t.SetYear(strconv.Itoa(*updates.Year))
		result.Updated = append(result.Updated, "year")
	}

	if updates.Genre != nil {
		t.SetGenre(*updates.Genre)
		result.Updated = append(result.Updated, "genre")
	}

	if updates.Composer != nil {
		// TCOM frame for composer
		t.AddTextFrame(t.CommonID("Composer"), t.DefaultEncoding(), *updates.Composer)
		result.Updated = append(result.Updated, "composer")
	}

	if updates.Comment != nil {
		// Remove existing comments and add new one
		t.DeleteFrames(t.CommonID("Comments"))
		comment := id3v2.CommentFrame{
			Encoding: t.DefaultEncoding(),
			Language: "eng",
			Text:     *updates.Comment,
		}
		t.AddCommentFrame(comment)
		result.Updated = append(result.Updated, "comment")
	}

	// Handle track number
	if updates.Track != nil || updates.TrackTotal != nil {
		trackStr := ""
		if updates.Track != nil {
			trackStr = strconv.Itoa(*updates.Track)
		}
		if updates.TrackTotal != nil {
			if trackStr == "" {
				trackStr = "0"
			}
			trackStr = fmt.Sprintf("%s/%d", trackStr, *updates.TrackTotal)
		}
		if trackStr != "" {
			t.AddTextFrame(t.CommonID("Track"), t.DefaultEncoding(), trackStr)
			result.Updated = append(result.Updated, "track")
		}
	}

	// Handle disc number
	if updates.Disc != nil || updates.DiscTotal != nil {
		discStr := ""
		if updates.Disc != nil {
			discStr = strconv.Itoa(*updates.Disc)
		}
		if updates.DiscTotal != nil {
			if discStr == "" {
				discStr = "0"
			}
			discStr = fmt.Sprintf("%s/%d", discStr, *updates.DiscTotal)
		}
		if discStr != "" {
			t.AddTextFrame(t.CommonID("TPOS"), t.DefaultEncoding(), discStr)
			result.Updated = append(result.Updated, "disc")
		}
	}

	// Handle album art
	if updates.ArtPath != nil && *updates.ArtPath != "" {
		artData, err := os.ReadFile(*updates.ArtPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read album art: %w", err)
		}

		// Determine MIME type from extension
		artExt := strings.ToLower(filepath.Ext(*updates.ArtPath))
		mimeType := "image/jpeg"
		switch artExt {
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".bmp":
			mimeType = "image/bmp"
		}

		// Remove existing pictures and add new one
		t.DeleteFrames(t.CommonID("Attached picture"))
		pic := id3v2.PictureFrame{
			Encoding:    t.DefaultEncoding(),
			MimeType:    mimeType,
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     artData,
		}
		t.AddAttachedPicture(pic)
		result.Updated = append(result.Updated, "art")
	}

	// Save changes
	if err := t.Save(); err != nil {
		return nil, fmt.Errorf("failed to save tags: %w", err)
	}

	result.Success = true
	return result, nil
}

// ClearTags removes all tags from an audio file.
// Currently supports MP3 files only.
func ClearTags(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".mp3" {
		return fmt.Errorf("tag clearing only supported for MP3 files")
	}

	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = t.Close() }()

	// Delete all frames
	t.DeleteAllFrames()

	if err := t.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// RemovePicture removes album art from an audio file.
// Currently supports MP3 files only.
func RemovePicture(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".mp3" {
		return fmt.Errorf("picture removal only supported for MP3 files")
	}

	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = t.Close() }()

	t.DeleteFrames(t.CommonID("Attached picture"))

	if err := t.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}
