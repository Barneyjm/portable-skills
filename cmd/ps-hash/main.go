package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Barneyjm/portable-skills/pkg/hash"
	"github.com/Barneyjm/portable-skills/pkg/skill"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed SKILL.md
var skillMD string

func main() {
	rootCmd := &cobra.Command{
		Use:     "ps-hash",
		Short:   "File hashing without dependencies",
		Long:    "Calculate and verify file checksums using MD5, SHA1, SHA256, or SHA512.",
		Version: version,
	}

	rootCmd.AddCommand(newCalcCmd())
	rootCmd.AddCommand(newVerifyCmd())
	rootCmd.AddCommand(newInstallSkillCmd())
	rootCmd.AddCommand(newUninstallSkillCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCalcCmd() *cobra.Command {
	var (
		algo       string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "calc <file> [file...]",
		Short: "Calculate file hash",
		Long:  "Calculate the hash of one or more files. Use - for stdin.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			algorithm := hash.Algorithm(strings.ToLower(algo))
			if !hash.IsSupported(string(algorithm)) {
				return fmt.Errorf("unsupported algorithm: %s (supported: md5, sha1, sha256, sha512)", algo)
			}

			var results []*hash.Result

			for _, path := range args {
				var result *hash.Result
				var err error

				if path == "-" {
					// Read from stdin
					hashStr, err := hash.Reader(os.Stdin, algorithm)
					if err != nil {
						return fmt.Errorf("failed to hash stdin: %w", err)
					}
					result = &hash.Result{
						File:      "-",
						Algorithm: string(algorithm),
						Hash:      hashStr,
					}
				} else {
					result, err = hash.File(path, algorithm)
					if err != nil {
						return fmt.Errorf("failed to hash %s: %w", path, err)
					}
				}

				results = append(results, result)
			}

			if jsonOutput {
				if len(results) == 1 {
					return skill.WriteJSON(results[0])
				}
				return skill.WriteJSON(results)
			}

			// Output in standard checksum format: hash  filename
			for _, r := range results {
				fmt.Printf("%s  %s\n", r.Hash, r.File)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&algo, "algorithm", "a", "sha256", "Hash algorithm: md5, sha1, sha256, sha512")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newVerifyCmd() *cobra.Command {
	var algo string

	cmd := &cobra.Command{
		Use:   "verify <file> <expected-hash>",
		Short: "Verify file hash",
		Long:  "Verify that a file matches an expected hash.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			expected := strings.ToLower(args[1])

			algorithm := hash.Algorithm(strings.ToLower(algo))
			if !hash.IsSupported(string(algorithm)) {
				return fmt.Errorf("unsupported algorithm: %s", algo)
			}

			match, result, err := hash.Verify(path, algorithm, expected)
			if err != nil {
				return err
			}

			if match {
				fmt.Printf("OK: %s matches %s hash\n", path, algorithm)
				return nil
			}

			fmt.Printf("MISMATCH: %s\n", path)
			fmt.Printf("  Expected: %s\n", expected)
			fmt.Printf("  Got:      %s\n", result.Hash)
			os.Exit(1)
			return nil
		},
	}

	cmd.Flags().StringVarP(&algo, "algorithm", "a", "sha256", "Hash algorithm: md5, sha1, sha256, sha512")

	return cmd
}

func newInstallSkillCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install as a Claude Code skill",
		Long: `Install ps-hash as a Claude Code skill.

This copies the binary and SKILL.md to ~/.claude/skills/hash/
so it can be invoked with /hash in Claude Code.`,
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
		Long:  `Remove ps-hash from Claude Code skills.`,
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
	return filepath.Join(home, ".claude", "skills", "hash")
}

func installSkill(force bool) error {
	skillDir := getSkillDir()
	if skillDir == "" {
		return fmt.Errorf("could not determine home directory")
	}

	binaryDest := filepath.Join(skillDir, "ps-hash")
	skillMDDest := filepath.Join(skillDir, "SKILL.md")

	if !force {
		if _, err := os.Stat(skillDir); err == nil {
			existing := []string{}
			if _, err := os.Stat(binaryDest); err == nil {
				existing = append(existing, "ps-hash")
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

	fmt.Println("Installing ps-hash skill...")

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
	fmt.Println("\nYou can now use /hash in Claude Code.")

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
