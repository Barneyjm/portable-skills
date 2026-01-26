package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Barneyjm/portable-skills/pkg/qr"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-qr",
		Short:   "QR code generation without dependencies",
		Long:    "Generate QR codes from text, URLs, or data. Output as PNG or ASCII.",
		Version: version,
	}

	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newInstallSkillCmd())
	rootCmd.AddCommand(newUninstallSkillCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newGenerateCmd() *cobra.Command {
	var (
		output string
		size   int
		level  string
		invert bool
		ascii  bool
	)

	cmd := &cobra.Command{
		Use:   "generate <content>",
		Short: "Generate a QR code",
		Long:  "Generate a QR code from text, URL, or other content.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content := args[0]

			lvl, err := qr.ParseLevel(level)
			if err != nil {
				return err
			}

			opts := qr.Options{
				Size:   size,
				Level:  lvl,
				Invert: invert,
			}

			// ASCII output
			if ascii {
				result, err := qr.GenerateASCII(content, opts)
				if err != nil {
					return err
				}
				fmt.Print(result)
				return nil
			}

			// PNG output
			if output == "-" || output == "" {
				// Output to stdout
				return qr.GenerateToStdout(content, opts)
			}

			// Output to file
			if err := qr.Generate(content, output, opts); err != nil {
				return err
			}

			skill.PrintSuccess("QR code saved to %s", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().IntVarP(&size, "size", "s", 256, "Image size in pixels")
	cmd.Flags().StringVarP(&level, "level", "l", "M", "Error correction level: L, M, Q, H")
	cmd.Flags().BoolVar(&invert, "invert", false, "Invert colors (white on black)")
	cmd.Flags().BoolVarP(&ascii, "ascii", "a", false, "Output as ASCII art (for terminal)")

	return cmd
}

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-qr as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/qr/
so it can be invoked with /qr in Claude Code.`,
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
		Long:  `Remove ps-qr from Claude Code skills.`,
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
	return filepath.Join(home, ".claude", "skills", "qr")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	binaryDest := filepath.Join(skillDir, "ps-qr")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-qr")
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

	fmt.Println("Installing ps-qr skill...")

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
	fmt.Println("\nYou can now use /qr in Claude Code.")

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
