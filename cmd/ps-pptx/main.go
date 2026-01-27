package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	rootCmd.AddCommand(newThemesCmd())
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
		output     string
		title      string
		creator    string
		theme      string
		background string
		jsonInput  bool
		images     []string
	)

	cmd := &cobra.Command{
		Use:   "create <input>",
		Short: "Create a PowerPoint from text or JSON input",
		Long: `Create a PowerPoint presentation from a text file, JSON file, or stdin.
Use - for stdin.

TEXT INPUT MODE (default):
Each paragraph (separated by blank lines) becomes a slide.
The first line of each paragraph is the slide title.

JSON INPUT MODE (--json):
Accepts structured JSON for full control over positioning, styling, and elements.
See 'ps-pptx themes --help' for available theme presets.

Example JSON:
{
  "options": {
    "title": "My Presentation",
    "theme": "teal",
    "background": {"color": "023047"}
  },
  "slides": [
    {
      "title": "Welcome",
      "body": "Introduction text",
      "background": {"color": "034764"}
    },
    {
      "elements": [
        {
          "type": "text",
          "text": {
            "content": "Custom positioned text",
            "position": {"x": 1, "y": 2, "width": 8, "height": 1},
            "style": {"fontSize": 36, "bold": true, "color": "FFFFFF"}
          }
        },
        {
          "type": "image",
          "image": {
            "path": "qr.png",
            "position": {"x": 6.5, "y": 1, "width": 3, "height": 3}
          }
        },
        {
          "type": "shape",
          "shape": {
            "type": "rectangle",
            "position": {"x": 0.5, "y": 0.5, "width": 9, "height": 0.1},
            "fillColor": "FFB703"
          }
        }
      ]
    }
  ]
}

IMAGE FLAG FORMAT:
--image slide:N,path:FILE,x:X,y:Y,w:W,h:H
Example: --image slide:5,path:qr.png,x:6.5,y:1,w:3,h:3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			if output == "" {
				return fmt.Errorf("--output is required")
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

			// JSON input mode
			if jsonInput {
				if err := pptx.CreateFromJSON(output, content); err != nil {
					return err
				}
				fmt.Printf("Created %s\n", output)
				return nil
			}

			// Enhanced text mode with theme/background/images
			opts := pptx.EnhancedCreateOptions{
				Title:   title,
				Creator: creator,
				Theme:   theme,
			}

			// Parse background color
			if background != "" {
				opts.Background = &pptx.Background{Color: background}
			}

			// Parse --image flags and create slides with images
			imagesBySlide := make(map[int][]pptx.ImageElement)
			for _, imgSpec := range images {
				slideNum, imgElem, err := parseImageSpec(imgSpec)
				if err != nil {
					return fmt.Errorf("invalid --image format: %w", err)
				}
				imagesBySlide[slideNum] = append(imagesBySlide[slideNum], imgElem)
			}

			// Parse text into slides
			paragraphs := strings.Split(strings.TrimSpace(string(content)), "\n\n")
			var slides []pptx.EnhancedSlide

			for i, para := range paragraphs {
				lines := strings.SplitN(strings.TrimSpace(para), "\n", 2)
				slide := pptx.EnhancedSlide{
					Title: lines[0],
				}
				if len(lines) > 1 {
					slide.Body = lines[1]
				}

				// Add any images for this slide
				if imgs, ok := imagesBySlide[i+1]; ok { // slides are 1-indexed
					for _, img := range imgs {
						imgCopy := img
						slide.Elements = append(slide.Elements, pptx.Element{
							Type:  "image",
							Image: &imgCopy,
						})
					}
				}

				slides = append(slides, slide)
			}

			if err := pptx.CreateEnhanced(output, slides, opts); err != nil {
				return err
			}

			fmt.Printf("Created %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output .pptx path (required)")
	cmd.Flags().StringVar(&title, "title", "", "Presentation title")
	cmd.Flags().StringVar(&creator, "creator", "", "Creator name")
	cmd.Flags().StringVar(&theme, "theme", "", "Theme preset (default, dark, light, teal, coral)")
	cmd.Flags().StringVar(&background, "background", "", "Default background color (hex without #, e.g., 023047)")
	cmd.Flags().BoolVar(&jsonInput, "json", false, "Parse input as JSON for full control")
	cmd.Flags().StringArrayVar(&images, "image", nil, "Add image to slide (slide:N,path:FILE,x:X,y:Y,w:W,h:H)")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

// parseImageSpec parses "slide:N,path:FILE,x:X,y:Y,w:W,h:H" format
func parseImageSpec(spec string) (int, pptx.ImageElement, error) {
	var img pptx.ImageElement
	slideNum := 1

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return 0, img, fmt.Errorf("invalid key:value pair: %s", part)
		}
		key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

		switch strings.ToLower(key) {
		case "slide":
			n, err := strconv.Atoi(value)
			if err != nil {
				return 0, img, fmt.Errorf("invalid slide number: %s", value)
			}
			slideNum = n
		case "path":
			img.Path = value
		case "x":
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, img, fmt.Errorf("invalid x value: %s", value)
			}
			img.Position.X = f
		case "y":
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, img, fmt.Errorf("invalid y value: %s", value)
			}
			img.Position.Y = f
		case "w", "width":
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, img, fmt.Errorf("invalid width value: %s", value)
			}
			img.Position.Width = f
		case "h", "height":
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, img, fmt.Errorf("invalid height value: %s", value)
			}
			img.Position.Height = f
		case "alt":
			img.AltText = value
		default:
			return 0, img, fmt.Errorf("unknown key: %s", key)
		}
	}

	if img.Path == "" {
		return 0, img, fmt.Errorf("path is required")
	}

	return slideNum, img, nil
}

func newThemesCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "themes",
		Short: "List available theme presets",
		Long: `List available theme presets with their color schemes.

Available themes:
  default - Classic Office blue theme
  dark    - Dark mode with syntax-highlighting-inspired colors
  light   - Clean light theme with modern colors
  teal    - Professional teal/gold palette
  coral   - Warm coral/green accent palette`,
		RunE: func(cmd *cobra.Command, args []string) error {
			presets := pptx.GetThemePresets()

			if jsonOutput {
				themes := make(map[string]pptx.Theme)
				for _, name := range presets {
					themes[name] = pptx.ThemePresets[name]
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(themes)
			}

			fmt.Println("Available theme presets:")
			fmt.Println()
			for _, name := range presets {
				theme := pptx.ThemePresets[name]
				fmt.Printf("  %s\n", name)
				fmt.Printf("    Background: #%s  Text: #%s\n", theme.Colors.Light1, theme.Colors.Dark1)
				fmt.Printf("    Accents: #%s #%s #%s\n", theme.Colors.Accent1, theme.Colors.Accent2, theme.Colors.Accent3)
				fmt.Printf("    Fonts: %s / %s\n", theme.TitleFont, theme.BodyFont)
				fmt.Println()
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
