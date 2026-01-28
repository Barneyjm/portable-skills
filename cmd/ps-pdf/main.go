package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Barneyjm/portable-skills/pkg/pdf"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-pdf",
		Short:   "PDF text extraction and creation without dependencies",
		Long:    "Extract text from PDFs and create PDFs from text. No dependencies required.",
		Version: version,
	}

	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newTextCmd())
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
		Short: "Display PDF metadata",
		Long:  "Show information about PDF files including page count.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var results []*pdf.Info

			for _, path := range args {
				info, err := pdf.GetInfo(path)
				if err != nil {
					info = &pdf.Info{
						Path:  path,
						Error: err.Error(),
					}
				}
				results = append(results, info)
			}

			if jsonOutput {
				if len(results) == 1 {
					return skill.WriteJSON(results[0])
				}
				return skill.WriteJSON(results)
			}

			for _, info := range results {
				if info.Error != "" {
					fmt.Printf("%s: error - %s\n", info.Path, info.Error)
				} else {
					fmt.Printf("%s: %d page(s)\n", info.Path, info.Pages)
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
		pageNum int
		output  string
	)

	cmd := &cobra.Command{
		Use:   "text <file>",
		Short: "Extract text content from a PDF",
		Long:  "Extract all text from a PDF file, or from a specific page.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			var text string
			var err error

			if pageNum > 0 {
				text, err = pdf.ExtractPageText(path, pageNum)
			} else {
				text, err = pdf.ExtractText(path)
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

	cmd.Flags().IntVarP(&pageNum, "page", "p", 0, "Extract only a specific page (1-indexed)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write to file instead of stdout")

	return cmd
}

func newCreateCmd() *cobra.Command {
	var (
		output       string
		title        string
		author       string
		pageSize     string
		fontSize     float64
		font         string
		jsonInput    bool
		markdown     bool
		marginTop    float64
		marginBottom float64
		marginLeft   float64
		marginRight  float64
		pageNumbers  bool
		pageNumPos   string
		lineSpacing  float64
		orientation  string
	)

	cmd := &cobra.Command{
		Use:   "create <input>",
		Short: "Create a PDF from text, Markdown, or JSON input",
		Long: `Create a PDF from a text file, Markdown file, JSON file, or stdin. Use - for stdin.

TEXT INPUT (default):
Plain text is rendered as a single block with automatic line wrapping.

MARKDOWN INPUT (--markdown):
Parses Markdown syntax including:
  - Headings (# H1, ## H2, ... ###### H6)
  - Lists (bullet and numbered)
  - Code blocks (triple backticks)
  - Horizontal rules (---, ***, ___)
  - Paragraphs separated by blank lines

JSON INPUT (--json):
Accepts structured JSON for full control over document layout.

Example JSON:
{
  "options": {
    "title": "My Document",
    "author": "Jane Doe",
    "pageNumbers": true,
    "marginTop": 25,
    "marginBottom": 25
  },
  "markdown": "# Hello World\n\nThis is a paragraph.\n\n- Item 1\n- Item 2"
}

Or with explicit paragraphs:
{
  "options": { "title": "Report", "pageNumbers": true },
  "paragraphs": [
    { "type": "heading", "level": 1, "content": "Introduction" },
    { "type": "text", "content": "This is the intro paragraph." },
    { "type": "list", "items": ["First item", "Second item"], "ordered": false },
    { "type": "code", "content": "function hello() {\n  return 'world';\n}" },
    { "type": "hr" }
  ]
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			if output == "" {
				return fmt.Errorf("--output is required")
			}

			// Read input content
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
				if err := pdf.CreateFromJSON(output, content); err != nil {
					return err
				}
				fmt.Printf("Created %s\n", output)
				return nil
			}

			// Build enhanced options
			opts := pdf.EnhancedCreateOptions{
				Title:       title,
				Author:      author,
				PageSize:    pageSize,
				Orientation: orientation,
				FontFamily:  font,
				FontSize:    fontSize,
				PageNumbers: pageNumbers,
				PageNumPos:  pageNumPos,
				LineSpacing: lineSpacing,
			}

			// Only set margins if explicitly provided via flags
			if cmd.Flags().Changed("margin-top") {
				opts.MarginTop = &marginTop
			}
			if cmd.Flags().Changed("margin-bottom") {
				opts.MarginBottom = &marginBottom
			}
			if cmd.Flags().Changed("margin-left") {
				opts.MarginLeft = &marginLeft
			}
			if cmd.Flags().Changed("margin-right") {
				opts.MarginRight = &marginRight
			}

			// Markdown input mode
			if markdown {
				if err := pdf.CreateFromMarkdown(output, string(content), opts); err != nil {
					return err
				}
				fmt.Printf("Created %s\n", output)
				return nil
			}

			// Plain text mode (enhanced)
			if err := pdf.CreateEnhancedFromText(output, string(content), opts); err != nil {
				return err
			}

			fmt.Printf("Created %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output PDF path (required)")
	cmd.Flags().StringVar(&title, "title", "", "Document title")
	cmd.Flags().StringVar(&author, "author", "", "Document author")
	cmd.Flags().StringVar(&pageSize, "page-size", "A4", "Page size: A4, Letter, Legal")
	cmd.Flags().StringVar(&orientation, "orientation", "P", "Orientation: P (portrait), L (landscape)")
	cmd.Flags().Float64Var(&fontSize, "font-size", 12, "Font size in points")
	cmd.Flags().StringVar(&font, "font", "Helvetica", "Font family: Helvetica, Times, Courier")
	cmd.Flags().BoolVar(&jsonInput, "json", false, "Parse input as JSON for full control")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "Parse input as Markdown")
	cmd.MarkFlagsMutuallyExclusive("json", "markdown")
	cmd.Flags().Float64Var(&marginTop, "margin-top", 20, "Top margin in mm")
	cmd.Flags().Float64Var(&marginBottom, "margin-bottom", 20, "Bottom margin in mm")
	cmd.Flags().Float64Var(&marginLeft, "margin-left", 20, "Left margin in mm")
	cmd.Flags().Float64Var(&marginRight, "margin-right", 20, "Right margin in mm")
	cmd.Flags().BoolVar(&pageNumbers, "page-numbers", false, "Add page numbers")
	cmd.Flags().StringVar(&pageNumPos, "page-num-pos", "bottom-center", "Page number position: bottom-left, bottom-center, bottom-right")
	cmd.Flags().Float64Var(&lineSpacing, "line-spacing", 1.2, "Line height multiplier")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-pdf as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/pdf/
so it can be invoked with /pdf in Claude Code.`,
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
		Long:  `Remove ps-pdf from Claude Code skills.`,
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
	return filepath.Join(home, ".claude", "skills", "pdf")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	binaryDest := filepath.Join(skillDir, "ps-pdf")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-pdf")
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

	fmt.Println("Installing ps-pdf skill...")

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
	fmt.Println("\nYou can now use /pdf in Claude Code.")

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
