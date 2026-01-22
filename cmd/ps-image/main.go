package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Barneyjm/portable-skills/pkg/image"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

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
	rootCmd.AddCommand(newInstallSkillCmd())
	rootCmd.AddCommand(newUninstallSkillCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newResizeCmd() *cobra.Command {
	var (
		width        int
		height       int
		output       string
		fit          string
		quality      int
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "resize <input>",
		Short: "Resize an image",
		Long:  "Resize an image to specified dimensions. Use - for stdin/stdout.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]
			var detectedFormat string

			// Handle stdin
			if inputPath == "-" {
				input, err := skill.NewInputFromStdin()
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer func() { _ = input.Close() }()
				inputPath = input.Path
				detectedFormat = input.DetectedFormat
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
				// Determine output format for stdout
				var stdoutFormat image.Format
				if outputFormat != "" {
					// User specified format explicitly
					stdoutFormat = image.NormalizeFormat(outputFormat)
					if !image.IsSupported(stdoutFormat) {
						return fmt.Errorf("unsupported output format: %s", outputFormat)
					}
				} else if detectedFormat != "" {
					// Use detected format from stdin
					stdoutFormat = image.NormalizeFormat(detectedFormat)
				} else {
					// Use input file's format
					stdoutFormat = image.FormatFromExtension(inputPath)
				}

				// WebP can't be encoded, fall back to PNG
				if stdoutFormat == image.FormatWEBP {
					stdoutFormat = image.FormatPNG
				}

				return image.Encode(os.Stdout, resized, image.SaveOptions{
					Format:  stdoutFormat,
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
	cmd.Flags().IntVar(&height, "height", 0, "Target height in pixels")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path (default: <input>_resized.<ext>)")
	cmd.Flags().StringVar(&fit, "fit", "contain", "Fit mode: contain, cover, fill")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Output quality for JPEG (1-100)")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Output format for stdout (png, jpg, gif, bmp, tiff)")

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
				input, err := skill.NewInputFromStdin()
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer func() { _ = input.Close() }()
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
	_ = cmd.MarkFlagRequired("format")

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
				input, err := skill.NewInputFromStdin()
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				defer func() { _ = input.Close() }()
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

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-image as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/image/
so it can be invoked with /image in Claude Code.

This is a "good skill citizen" - it only modifies ~/.claude/skills/image/
and never touches other skills.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installSkill(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing installation without prompting")

	return cmd
}

func newUninstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "uninstall-skill",
		Short: "Uninstall from Claude Code skills",
		Long: `Remove ps-image from Claude Code skills.

This removes ~/.claude/skills/image/ and nothing else.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallSkill(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Remove without prompting")

	return cmd
}

func getSkillDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "skills", "image")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Check for existing installation
	binaryDest := filepath.Join(skillDir, "ps-image")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		// Check if skill directory exists and has files
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-image")
			}
			if _, err := os.Stat(skillMDDest); err == nil {
				existing = append(existing, "SKILL.md")
			}

			if len(existing) > 0 {
				fmt.Printf("Existing installation found at %s:\n", skillDir)
				for _, f := range existing {
					fmt.Printf("  - %s\n", f)
				}
				fmt.Print("\nOverwrite? [y/N] ")
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" && response != "yes" {
					fmt.Println("Installation cancelled.")
					return nil
				}
			}
		}
	}

	// Create skill directory
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Copy the current binary
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	fmt.Println("Installing ps-image skill...")

	// Copy binary
	if err := copyFile(execPath, binaryDest); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	if err := os.Chmod(binaryDest, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}
	fmt.Printf("  ✓ Copied binary to %s\n", binaryDest)

	// Write embedded SKILL.md
	if err := os.WriteFile(skillMDDest, []byte(skillMD), 0644); err != nil {
		return fmt.Errorf("failed to write SKILL.md: %w", err)
	}
	fmt.Printf("  ✓ Wrote SKILL.md to %s\n", skillMDDest)

	fmt.Println("\n✓ Installation complete!")
	fmt.Println("\nYou can now use /image in Claude Code.")

	return nil
}

func uninstallSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Check if directory exists
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		fmt.Printf("Skill directory does not exist: %s\n", skillDir)
		fmt.Println("Nothing to uninstall.")
		return nil
	}

	// List contents
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return fmt.Errorf("failed to read skill directory: %w", err)
	}

	if !force {
		fmt.Printf("Will remove: %s\n\n", skillDir)
		fmt.Println("Contents:")
		for _, entry := range entries {
			fmt.Printf("  - %s\n", entry.Name())
		}
		fmt.Print("\nRemove this skill? [y/N] ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	// Remove the directory
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	fmt.Println("\n✓ Skill uninstalled.")

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
