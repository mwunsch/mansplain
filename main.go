package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/mwunsch/mansplain/cmd"
	"github.com/mwunsch/mansplain/internal/ui"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

//go:embed SKILL.md
var skillFile string

const (
	promptStart = "<!-- system-prompt:start -->"
	promptEnd   = "<!-- system-prompt:end -->"
)

func extractSystemPrompt(skill string) string {
	start := strings.Index(skill, promptStart)
	end := strings.Index(skill, promptEnd)
	if start < 0 || end < 0 || end <= start {
		// Fallback: use everything after the frontmatter
		if idx := strings.Index(skill, "---\n"); idx >= 0 {
			rest := skill[idx+4:]
			if idx2 := strings.Index(rest, "---\n"); idx2 >= 0 {
				return strings.TrimSpace(rest[idx2+4:])
			}
		}
		return skill
	}
	return strings.TrimSpace(skill[start+len(promptStart) : end])
}

func main() {
	cmd.SetVersion(version)
	cmd.SetGeneratePrompt(extractSystemPrompt(skillFile))
	if err := cmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
