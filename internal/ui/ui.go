package ui

import (
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
)

// Catppuccin Mocha palette
var (
	subtle    = lipgloss.Color("#6c7086") // overlay0
	blue      = lipgloss.Color("#89b4fa")
	mauve     = lipgloss.Color("#cba6f7")
	green     = lipgloss.Color("#a6e3a1")
	red       = lipgloss.Color("#f38ba8")
	yellow    = lipgloss.Color("#f9e2af")
	text      = lipgloss.Color("#cdd6f4")
	surface0  = lipgloss.Color("#313244")
)

// Reusable styles
var (
	StyleSubtle    = lipgloss.NewStyle().Foreground(subtle)
	StyleAccent    = lipgloss.NewStyle().Foreground(blue)
	StyleHighlight = lipgloss.NewStyle().Foreground(mauve)
	StyleSuccess   = lipgloss.NewStyle().Foreground(green)
	StyleError     = lipgloss.NewStyle().Foreground(red)
	StyleWarning   = lipgloss.NewStyle().Foreground(yellow)
	StyleBold      = lipgloss.NewStyle().Bold(true).Foreground(text)
)

// Info prints an informational label + value to stderr.
func Info(label, value string) {
	out := StyleSubtle.Render(label) + " " + value
	lipgloss.Fprintln(os.Stderr, out)
}

// Success prints a green checkmark and message to stderr.
func Success(msg string) {
	icon := StyleSuccess.Render("✓")
	lipgloss.Fprintln(os.Stderr, icon+" "+msg)
}

// Error prints a red X and message to stderr.
func Error(msg string) {
	icon := StyleError.Render("✗")
	lipgloss.Fprintln(os.Stderr, icon+" "+msg)
}

// Warning prints a yellow exclamation and message to stderr.
func Warning(msg string) {
	icon := StyleWarning.Render("!")
	lipgloss.Fprintln(os.Stderr, icon+" "+msg)
}

// ProviderBanner prints a styled connection info line to stderr.
func ProviderBanner(url, model string) {
	urlStr := StyleSubtle.Render(url)

	var line string
	if model != "" {
		line = StyleAccent.Render(model) + StyleSubtle.Render(" · ") + urlStr
	} else {
		line = urlStr
	}

	lipgloss.Fprintln(os.Stderr, line)
}

// TokenUsage prints token usage stats to stderr.
func TokenUsage(prompt, completion, total int, elapsed time.Duration) {
	if total == 0 {
		return
	}
	dur := elapsed.Truncate(time.Millisecond)

	var toks string
	if prompt > 0 && completion > 0 {
		toks = fmt.Sprintf("%d tokens (%d in · %d out)", total, prompt, completion)
	} else {
		toks = fmt.Sprintf("%d tokens", total)
	}

	line := StyleSubtle.Render(fmt.Sprintf("  %s in %v", toks, dur))
	lipgloss.Fprintln(os.Stderr, line)
}
