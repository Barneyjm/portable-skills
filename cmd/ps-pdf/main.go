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
		output   string
		title    string
		author   string
		pageSize string
		fontSize float64
		font     string
	)

	cmd := &cobra.Command{
		Use:   "create <input>",
		Short: "Create a PDF from text input",
		Long:  "Create a PDF from a text file or stdin. Use - for stdin.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			if output == "" {
				return fmt.Errorf("--output is required")
			}

			opts := pdf.CreateOptions{
				Title:      title,
				Author:     author,
				PageSize:   pageSize,
				FontSize:   fontSize,
				FontFamily: font,
			}

			var err error
			if input == "-" {
				err = pdf.CreateFromReader(output, os.Stdin, opts)
			} else {
				err = pdf.CreateFromFile(output, input, opts)
			}

			if err != nil {
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
	cmd.Flags().Float64Var(&fontSize, "font-size", 12, "Font size in points")
	cmd.Flags().StringVar(&font, "font", "Helvetica", "Font family: Helvetica, Times, Courier")
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
