package main

import (
	_ "embed"
	"os"

	"github.com/mwunsch/mansplain/cmd"
	"github.com/mwunsch/mansplain/internal/ui"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

//go:embed prompts/generate.md
var generatePrompt string

func main() {
	cmd.SetVersion(version)
	cmd.SetGeneratePrompt(generatePrompt)
	if err := cmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
