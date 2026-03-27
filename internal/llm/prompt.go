package llm

import (
	"fmt"
	"strings"
)

// systemPrompt holds the loaded system prompt. Set via SetSystemPrompt.
var systemPrompt string

// SetSystemPrompt sets the system prompt used for generation.
func SetSystemPrompt(prompt string) {
	systemPrompt = prompt
}

// SystemPrompt returns the system prompt for man page generation.
func SystemPrompt() string {
	return systemPrompt
}

// BuildUserPrompt constructs the user message from source material.
func BuildUserPrompt(req GenerateRequest) string {
	var b strings.Builder

	upper := strings.ToUpper(req.Name)
	fmt.Fprintf(&b, "Generate an mdoc(7) man page for %s (section %d).\n", req.Name, req.Section)
	fmt.Fprintf(&b, "Start with exactly:\n.Dd $Mdocdate$\n.Dt %s %d\n.Os\n.Sh NAME\n.Nm %s\n\n", upper, req.Section, req.Name)

	for _, src := range req.Sources {
		switch src.Type {
		case "help":
			fmt.Fprintf(&b, "## --help output\n\n```\n%s\n```\n\n", src.Content)
		case "readme":
			fmt.Fprintf(&b, "## README\n\n%s\n\n", src.Content)
		case "stdin":
			fmt.Fprintf(&b, "## Source material\n\n%s\n\n", src.Content)
		}
	}

	return b.String()
}
