package image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// Metadata contains image information
type Metadata struct {
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Format    string    `json:"format"`
	SizeBytes int64     `json:"size_bytes"`
	ColorMode string    `json:"color_mode,omitempty"`
	HasAlpha  bool      `json:"has_alpha"`
	EXIF      *EXIFData `json:"exif,omitempty"`
}

// EXIFData contains EXIF metadata from photos
type EXIFData struct {
	Make         string     `json:"make,omitempty"`
	Model        string     `json:"model,omitempty"`
	DateTime     *time.Time `json:"datetime,omitempty"`
	Orientation  int        `json:"orientation,omitempty"`
	ExposureTime string     `json:"exposure_time,omitempty"`
	FNumber      string     `json:"f_number,omitempty"`
	ISOSpeed     int        `json:"iso_speed,omitempty"`
	FocalLength  string     `json:"focal_length,omitempty"`
	GPSLatitude  float64    `json:"gps_latitude,omitempty"`
	GPSLongitude float64    `json:"gps_longitude,omitempty"`
}

// Info extracts metadata from an image file
func Info(inputPath string) (*Metadata, error) {
	// Get file info
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Load image to get dimensions
	img, err := Load(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	bounds := img.Bounds()
	format := detectFormat(inputPath)

	meta := &Metadata{
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		Format:    format,
		SizeBytes: fileInfo.Size(),
		HasAlpha:  hasAlphaChannel(img),
	}

	// Try to extract EXIF data (only works for JPEG/TIFF)
	if format == "jpeg" || format == "tiff" {
		exifData, err := extractEXIF(inputPath)
		if err == nil {
			meta.EXIF = exifData
		}
		// Ignore EXIF errors - not all images have EXIF
	}

	return meta, nil
}

// detectFormat determines the format from file extension
func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".bmp":
		return "bmp"
	case ".tiff", ".tif":
		return "tiff"
	case ".webp":
		return "webp"
	default:
		return "unknown"
	}
}

// hasAlphaChannel checks if an image has an alpha channel
func hasAlphaChannel(img interface{}) bool {
	switch v := img.(type) {
	case interface{ Opaque() bool }:
		return !v.Opaque()
	default:
		return false
	}
}

// extractEXIF reads EXIF metadata from an image file
func extractEXIF(inputPath string) (*EXIFData, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	x, err := exif.Decode(f)
	if err != nil {
		return nil, err
	}

	data := &EXIFData{}

	// Camera make/model
	if tag, err := x.Get(exif.Make); err == nil {
		data.Make, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		data.Model, _ = tag.StringVal()
	}

	// Date/time
	if tm, err := x.DateTime(); err == nil {
		data.DateTime = &tm
	}

	// Orientation
	if tag, err := x.Get(exif.Orientation); err == nil {
		data.Orientation, _ = tag.Int(0)
	}

	// Exposure settings
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		num, denom, _ := tag.Rat2(0)
		if denom != 0 {
			data.ExposureTime = fmt.Sprintf("%d/%d", num, denom)
		}
	}
	if tag, err := x.Get(exif.FNumber); err == nil {
		num, denom, _ := tag.Rat2(0)
		if denom != 0 {
			data.FNumber = fmt.Sprintf("f/%.1f", float64(num)/float64(denom))
		}
	}
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		data.ISOSpeed, _ = tag.Int(0)
	}
	if tag, err := x.Get(exif.FocalLength); err == nil {
		num, denom, _ := tag.Rat2(0)
		if denom != 0 {
			data.FocalLength = fmt.Sprintf("%.1fmm", float64(num)/float64(denom))
		}
	}

	// GPS coordinates
	if lat, long, err := x.LatLong(); err == nil {
		data.GPSLatitude = lat
		data.GPSLongitude = long
	}

	return data, nil
}
