package skill

import (
	"fmt"
	"strings"
)

// Command represents a CLI command in a skill
type Command struct {
	Name        string
	Description string
	Usage       string
	Flags       []Flag
	Examples    []string
}

// Flag represents a command flag
type Flag struct {
	Name        string
	Short       string
	Description string
	Type        string // "string", "int", "bool"
	Default     string
	Required    bool
}

// Manifest represents a SKILL.md definition
type Manifest struct {
	Name        string
	Description string
	Version     string
	Commands    []Command
	Formats     []string // Supported formats (for file-based skills)
	Installation string
}

// GenerateSKILLMD generates a SKILL.md file content
func (m *Manifest) GenerateSKILLMD() string {
	var sb strings.Builder

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", m.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", m.Description))
	if m.Version != "" {
		sb.WriteString(fmt.Sprintf("version: %s\n", m.Version))
	}
	sb.WriteString("---\n\n")

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", toTitle(m.Name)))
	sb.WriteString(fmt.Sprintf("%s\n\n", m.Description))

	// Commands
	sb.WriteString("## Available Commands\n\n")
	for _, cmd := range m.Commands {
		sb.WriteString(fmt.Sprintf("### %s\n", cmd.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", cmd.Description))
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("%s\n", cmd.Usage))
		sb.WriteString("```\n\n")

		if len(cmd.Flags) > 0 {
			sb.WriteString("**Flags:**\n")
			for _, f := range cmd.Flags {
				req := ""
				if f.Required {
					req = " (required)"
				}
				def := ""
				if f.Default != "" {
					def = fmt.Sprintf(" [default: %s]", f.Default)
				}
				if f.Short != "" {
					sb.WriteString(fmt.Sprintf("- `-%s, --%s`: %s%s%s\n", f.Short, f.Name, f.Description, req, def))
				} else {
					sb.WriteString(fmt.Sprintf("- `--%s`: %s%s%s\n", f.Name, f.Description, req, def))
				}
			}
			sb.WriteString("\n")
		}

		if len(cmd.Examples) > 0 {
			sb.WriteString("**Examples:**\n```bash\n")
			for _, ex := range cmd.Examples {
				sb.WriteString(fmt.Sprintf("%s\n", ex))
			}
			sb.WriteString("```\n\n")
		}
	}

	// Supported formats
	if len(m.Formats) > 0 {
		sb.WriteString("## Supported Formats\n\n")
		sb.WriteString(strings.Join(m.Formats, ", "))
		sb.WriteString("\n\n")
	}

	// Installation
	sb.WriteString("## Installation\n\n")
	if m.Installation != "" {
		sb.WriteString(m.Installation)
	} else {
		sb.WriteString("Binary is self-contained. No runtime dependencies required.\n")
	}

	return sb.String()
}

func toTitle(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}
