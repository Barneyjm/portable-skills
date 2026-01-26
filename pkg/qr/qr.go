// Package qr provides QR code generation utilities.
package qr

import (
	"fmt"
	"image/png"
	"os"

	"github.com/skip2/go-qrcode"
)

// Level represents the error correction level.
type Level int

const (
	LevelL Level = iota // 7% recovery
	LevelM              // 15% recovery
	LevelQ              // 25% recovery
	LevelH              // 30% recovery
)

// ParseLevel parses a level string.
func ParseLevel(s string) (Level, error) {
	switch s {
	case "L", "l", "low":
		return LevelL, nil
	case "M", "m", "medium":
		return LevelM, nil
	case "Q", "q", "quartile":
		return LevelQ, nil
	case "H", "h", "high":
		return LevelH, nil
	default:
		return LevelM, fmt.Errorf("invalid level: %s (use L, M, Q, or H)", s)
	}
}

func toQRLevel(l Level) qrcode.RecoveryLevel {
	switch l {
	case LevelL:
		return qrcode.Low
	case LevelM:
		return qrcode.Medium
	case LevelQ:
		return qrcode.High
	case LevelH:
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

// Options for QR code generation.
type Options struct {
	Size    int   // Image size in pixels (default 256)
	Level   Level // Error correction level (default M)
	Invert  bool  // Invert colors (white on black)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size:  256,
		Level: LevelM,
	}
}

// Generate creates a QR code image from content and saves to file.
func Generate(content string, outputPath string, opts Options) error {
	if opts.Size <= 0 {
		opts.Size = 256
	}

	qr, err := qrcode.New(content, toQRLevel(opts.Level))
	if err != nil {
		return fmt.Errorf("failed to create QR code: %w", err)
	}

	if opts.Invert {
		qr.ForegroundColor, qr.BackgroundColor = qr.BackgroundColor, qr.ForegroundColor
	}

	img := qr.Image(opts.Size)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// GenerateToStdout writes a QR code PNG to stdout.
func GenerateToStdout(content string, opts Options) error {
	if opts.Size <= 0 {
		opts.Size = 256
	}

	qr, err := qrcode.New(content, toQRLevel(opts.Level))
	if err != nil {
		return fmt.Errorf("failed to create QR code: %w", err)
	}

	if opts.Invert {
		qr.ForegroundColor, qr.BackgroundColor = qr.BackgroundColor, qr.ForegroundColor
	}

	img := qr.Image(opts.Size)

	if err := png.Encode(os.Stdout, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// GenerateASCII creates an ASCII representation of a QR code.
func GenerateASCII(content string, opts Options) (string, error) {
	qr, err := qrcode.New(content, toQRLevel(opts.Level))
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	return qr.ToSmallString(opts.Invert), nil
}
