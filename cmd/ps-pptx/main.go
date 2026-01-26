package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Barneyjm/portable-skills/pkg/pptx"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-pptx",
		Short:   "PowerPoint file processing without dependencies",
		Long:    "Read and create PowerPoint (.pptx) files. Extract text, get info, create presentations.",
		Version: version,
	}

	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newTextCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newCreateCmd())
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
		Short: "Display presentation metadata",
		Long:  "Show information about PowerPoint files including slide count and metadata.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var results []*pptx.Info

			for _, path := range args {
				info, err := pptx.GetInfo(path)
				if err != nil {
					info = &pptx.Info{
						Path:  path,
						Error: err.Error(),
					}
				}
				results = append(results, info)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if len(results) == 1 {
					return enc.Encode(results[0])
				}
				return enc.Encode(results)
			}

			for _, info := range results {
				if info.Error != "" {
					fmt.Printf("%s: error - %s\n", info.Path, info.Error)
				} else {
					fmt.Printf("%s: %d slide(s)", info.Path, info.SlideCount)
					if info.Title != "" {
						fmt.Printf(" - %s", info.Title)
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newTextCmd() *cobra.Command {
	var (
		slideNum int
		output   string
	)

	cmd := &cobra.Command{
		Use:   "text <file>",
		Short: "Extract text content from a presentation",
		Long:  "Extract all text from a PowerPoint file, or from a specific slide.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			var text string
			var err error

			if slideNum > 0 {
				text, err = pptx.ExtractSlideText(path, slideNum)
			} else {
				text, err = pptx.ExtractText(path)
			}

			if err != nil {
				return err
			}

			if output != "" {
				return os.WriteFile(output, []byte(text), 0644)
			}

			fmt.Println(text)
			return nil
		},
	}

	cmd.Flags().IntVarP(&slideNum, "slide", "s", 0, "Extract only a specific slide (1-indexed)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write to file instead of stdout")

	return cmd
}

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list <file>",
		Short: "List all slides",
		Long:  "List all slides in a presentation with their titles.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slides, err := pptx.ListSlides(args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(slides)
			}

			for _, slide := range slides {
				if slide.Title != "" {
					fmt.Printf("Slide %d: %s\n", slide.Number, slide.Title)
				} else {
					fmt.Printf("Slide %d: (no title)\n", slide.Number)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newCreateCmd() *cobra.Command {
	var (
		output  string
		title   string
		creator string
	)

	cmd := &cobra.Command{
		Use:   "create <input>",
		Short: "Create a PowerPoint from text input",
		Long: `Create a PowerPoint presentation from a text file or stdin.
Use - for stdin.

Each paragraph (separated by blank lines) becomes a slide.
The first line of each paragraph is the slide title.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			if output == "" {
				return fmt.Errorf("--output is required")
			}

			opts := pptx.CreateOptions{
				Title:   title,
				Creator: creator,
			}

			var content []byte
			var err error

			if input == "-" {
				content, err = io.ReadAll(os.Stdin)
			} else {
				content, err = os.ReadFile(input)
			}
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			if err := pptx.CreateFromText(output, string(content), opts); err != nil {
				return err
			}

			fmt.Printf("Created %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output .pptx path (required)")
	cmd.Flags().StringVar(&title, "title", "", "Presentation title")
	cmd.Flags().StringVar(&creator, "creator", "", "Creator name")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-pptx as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/pptx/
so it can be invoked with /pptx in Claude Code.`,
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
		Long:  `Remove ps-pptx from Claude Code skills.`,
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
	return filepath.Join(home, ".claude", "skills", "pptx")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	binaryDest := filepath.Join(skillDir, "ps-pptx")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-pptx")
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
				if _, err := fmt.Scanln(&response); err != nil || (response != "y" && response != "Y" && response != "yes") {
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

	fmt.Println("Installing ps-pptx skill...")

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
	fmt.Println("\nYou can now use /pptx in Claude Code.")

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
		if _, err := fmt.Scanln(&response); err != nil || (response != "y" && response != "Y" && response != "yes") {
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
