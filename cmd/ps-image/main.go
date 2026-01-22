package main

import (
	"fmt"
	"os"

	"github.com/Barneyjm/portable-skills/pkg/image"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-image",
		Short:   "Image processing without dependencies",
		Long:    "A portable image processing tool that works without system dependencies.\nSupports resize, convert, and metadata extraction for common image formats.",
		Version: version,
	}

	rootCmd.AddCommand(newResizeCmd())
	rootCmd.AddCommand(newConvertCmd())
	rootCmd.AddCommand(newInfoCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newResizeCmd() *cobra.Command {
	var (
		width   int
		height  int
		output  string
		fit     string
		quality int
	)

	cmd := &cobra.Command{
		Use:   "resize <input>",
		Short: "Resize an image",
		Long:  "Resize an image to specified dimensions. Use - for stdin/stdout.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Handle stdin
			if inputPath == "-" {
				input, err := skill.NewInputFromStdin(".png")
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer input.Close()
				inputPath = input.Path
			}

			// Validate dimensions
			if width <= 0 && height <= 0 {
				return fmt.Errorf("at least one of --width or --height must be specified")
			}

			// Load and resize
			opts := image.ResizeOptions{
				Width:   width,
				Height:  height,
				Fit:     fit,
				Quality: quality,
			}

			resized, err := image.ResizeFile(inputPath, opts)
			if err != nil {
				return err
			}

			// Determine output path
			if output == "" {
				output = image.GenerateOutputPath(inputPath, image.FormatFromExtension(inputPath), "resized")
			}

			// Handle stdout
			if output == "-" {
				return image.Encode(os.Stdout, resized, image.SaveOptions{
					Format:  image.FormatPNG,
					Quality: quality,
				})
			}

			// Save to file
			if err := image.Save(resized, output, image.SaveOptions{Quality: quality}); err != nil {
				return err
			}

			skill.PrintSuccess("Saved resized image to %s", output)
			return nil
		},
	}

	cmd.Flags().IntVarP(&width, "width", "w", 0, "Target width in pixels")
	cmd.Flags().IntVarP(&height, "height", "h", 0, "Target height in pixels")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path (default: <input>_resized.<ext>)")
	cmd.Flags().StringVar(&fit, "fit", "contain", "Fit mode: contain, cover, fill")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Output quality for JPEG (1-100)")

	return cmd
}

func newConvertCmd() *cobra.Command {
	var (
		format  string
		output  string
		quality int
	)

	cmd := &cobra.Command{
		Use:   "convert <input>",
		Short: "Convert image format",
		Long:  "Convert an image to a different format. Supported: png, jpg, gif, bmp, tiff.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Handle stdin
			if inputPath == "-" {
				input, err := skill.NewInputFromStdin(".png")
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer input.Close()
				inputPath = input.Path
			}

			// Validate format
			if format == "" {
				return fmt.Errorf("--format is required")
			}

			targetFormat := image.NormalizeFormat(format)
			if !image.IsSupported(targetFormat) {
				return fmt.Errorf("unsupported format: %s", format)
			}

			// Load image
			img, err := image.Load(inputPath)
			if err != nil {
				return fmt.Errorf("failed to load image: %w", err)
			}

			// Determine output path
			if output == "" {
				output = image.GenerateOutputPath(inputPath, targetFormat, "")
			}

			// Handle stdout
			if output == "-" {
				return image.Encode(os.Stdout, img, image.SaveOptions{
					Format:  targetFormat,
					Quality: quality,
				})
			}

			// Save to file
			if err := image.Save(img, output, image.SaveOptions{
				Format:  targetFormat,
				Quality: quality,
			}); err != nil {
				return err
			}

			skill.PrintSuccess("Converted image to %s", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "Target format: png, jpg, gif, bmp, tiff (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path (default: <input>.<format>)")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Output quality for JPEG (1-100)")
	cmd.MarkFlagRequired("format")

	return cmd
}

func newInfoCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info <input>",
		Short: "Show image metadata",
		Long:  "Extract and display image metadata including dimensions, format, and EXIF data.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Handle stdin
			if inputPath == "-" {
				input, err := skill.NewInputFromStdin(".png")
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer input.Close()
				inputPath = input.Path
			}

			// Get metadata
			meta, err := image.Info(inputPath)
			if err != nil {
				return err
			}

			// Output
			if jsonOutput {
				return skill.WriteJSON(meta)
			}

			// Human-readable output
			fmt.Printf("File: %s\n", inputPath)
			fmt.Printf("Dimensions: %dx%d\n", meta.Width, meta.Height)
			fmt.Printf("Format: %s\n", meta.Format)
			fmt.Printf("Size: %d bytes (%.2f KB)\n", meta.SizeBytes, float64(meta.SizeBytes)/1024)
			fmt.Printf("Has Alpha: %t\n", meta.HasAlpha)

			if meta.EXIF != nil {
				fmt.Println("\nEXIF Data:")
				if meta.EXIF.Make != "" {
					fmt.Printf("  Camera: %s %s\n", meta.EXIF.Make, meta.EXIF.Model)
				}
				if meta.EXIF.DateTime != nil {
					fmt.Printf("  Date: %s\n", meta.EXIF.DateTime.Format("2006-01-02 15:04:05"))
				}
				if meta.EXIF.ExposureTime != "" {
					fmt.Printf("  Exposure: %s\n", meta.EXIF.ExposureTime)
				}
				if meta.EXIF.FNumber != "" {
					fmt.Printf("  Aperture: %s\n", meta.EXIF.FNumber)
				}
				if meta.EXIF.ISOSpeed > 0 {
					fmt.Printf("  ISO: %d\n", meta.EXIF.ISOSpeed)
				}
				if meta.EXIF.FocalLength != "" {
					fmt.Printf("  Focal Length: %s\n", meta.EXIF.FocalLength)
				}
				if meta.EXIF.GPSLatitude != 0 || meta.EXIF.GPSLongitude != 0 {
					fmt.Printf("  GPS: %.6f, %.6f\n", meta.EXIF.GPSLatitude, meta.EXIF.GPSLongitude)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
