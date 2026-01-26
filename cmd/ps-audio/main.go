package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Barneyjm/portable-skills/pkg/audio"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-audio",
		Short:   "Audio metadata without dependencies",
		Long:    "Read audio file metadata and extract album art. Supports MP3, M4A, FLAC, OGG.",
		Version: version,
	}

	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newArtCmd())
	rootCmd.AddCommand(newInstallSkillCmd())
	rootCmd.AddCommand(newUninstallSkillCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newInfoCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info <file> [file...]",
		Short: "Show audio file metadata",
		Long:  "Display metadata (title, artist, album, etc.) from audio files.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var results []*audio.Metadata

			for _, path := range args {
				meta, err := audio.Info(path)
				if err != nil {
					return fmt.Errorf("failed to read %s: %w", path, err)
				}
				results = append(results, meta)
			}

			if jsonOutput {
				if len(results) == 1 {
					return skill.WriteJSON(results[0])
				}
				return skill.WriteJSON(results)
			}

			// Human readable output
			for i, m := range results {
				if i > 0 {
					fmt.Println()
				}
				printMetadata(m)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func printMetadata(m *audio.Metadata) {
	fmt.Printf("File: %s\n", m.File)
	if m.Format != "" {
		fmt.Printf("Format: %s (%s)\n", m.Format, m.FileType)
	} else if m.FileType != "" {
		fmt.Printf("Type: %s\n", m.FileType)
	}

	if m.Title != "" {
		fmt.Printf("Title: %s\n", m.Title)
	}
	if m.Artist != "" {
		fmt.Printf("Artist: %s\n", m.Artist)
	}
	if m.Album != "" {
		fmt.Printf("Album: %s\n", m.Album)
	}
	if m.AlbumArtist != "" && m.AlbumArtist != m.Artist {
		fmt.Printf("Album Artist: %s\n", m.AlbumArtist)
	}
	if m.Year != 0 {
		fmt.Printf("Year: %d\n", m.Year)
	}
	if m.Genre != "" {
		fmt.Printf("Genre: %s\n", m.Genre)
	}
	if m.Track != 0 {
		if m.TrackTotal != 0 {
			fmt.Printf("Track: %d/%d\n", m.Track, m.TrackTotal)
		} else {
			fmt.Printf("Track: %d\n", m.Track)
		}
	}
	if m.Disc != 0 {
		if m.DiscTotal != 0 {
			fmt.Printf("Disc: %d/%d\n", m.Disc, m.DiscTotal)
		} else {
			fmt.Printf("Disc: %d\n", m.Disc)
		}
	}
	if m.Composer != "" {
		fmt.Printf("Composer: %s\n", m.Composer)
	}
	if m.Comment != "" {
		fmt.Printf("Comment: %s\n", m.Comment)
	}
	if m.HasPicture {
		fmt.Println("Album Art: Yes")
	}
	fmt.Printf("Size: %s\n", formatSize(m.SizeBytes))
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func newArtCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "art <audio-file>",
		Short: "Extract album art",
		Long:  "Extract embedded album art from an audio file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			audioPath := args[0]

			if output == "" {
				// Default: same name as audio file with image extension
				base := filepath.Base(audioPath)
				ext := filepath.Ext(base)
				output = base[:len(base)-len(ext)] + ".jpg"
			}

			if err := audio.ExtractPicture(audioPath, output); err != nil {
				return err
			}

			skill.PrintSuccess("Album art saved to %s", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: <input>.jpg)")

	return cmd
}

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-audio as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/audio/
so it can be invoked with /audio in Claude Code.`,
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
		Long:  `Remove ps-audio from Claude Code skills.`,
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
	return filepath.Join(home, ".claude", "skills", "audio")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	binaryDest := filepath.Join(skillDir, "ps-audio")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-audio")
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

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	fmt.Println("Installing ps-audio skill...")

	if err := copyFile(execPath, binaryDest); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	if err := os.Chmod(binaryDest, 0755); err != nil {
		return fmt.Errorf("failed to set binary permissions: %w", err)
	}
	fmt.Printf("  ✓ Copied binary to %s\n", binaryDest)

	if err := os.WriteFile(skillMDDest, []byte(skillMD), 0644); err != nil {
		return fmt.Errorf("failed to write SKILL.md: %w", err)
	}
	fmt.Printf("  ✓ Wrote SKILL.md to %s\n", skillMDDest)

	fmt.Println("\n✓ Installation complete!")
	fmt.Println("\nYou can now use /audio in Claude Code.")

	return nil
}

func uninstallSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		fmt.Printf("Skill directory does not exist: %s\n", skillDir)
		fmt.Println("Nothing to uninstall.")
		return nil
	}

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
