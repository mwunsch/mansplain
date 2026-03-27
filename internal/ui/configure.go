package ui

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/mwunsch/mansplain/internal/config"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4o"
)

// RunConfigureForm runs the interactive configuration form.
// Returns the config to save, or nil if the user cancelled.
func RunConfigureForm(existing *config.Config) (*config.Config, error) {
	if fi, err := os.Stdin.Stat(); err == nil {
		if fi.Mode()&os.ModeCharDevice == 0 {
			return nil, fmt.Errorf("mansplain configure requires an interactive terminal")
		}
	}

	// Pre-fill defaults so the user can just hit enter to accept
	baseURL := existing.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	apiKey := existing.APIKey
	model := existing.Model
	if model == "" {
		model = defaultModel
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Base URL").
				Value(&baseURL),
			huh.NewInput().
				Title("API Key").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey),
			huh.NewInput().
				Title("Model").
				Value(&model),
		),
	).WithTheme(huh.ThemeFunc(huh.ThemeCatppuccin)).Run()

	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, nil
		}
		return nil, err
	}

	return &config.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}, nil
}
