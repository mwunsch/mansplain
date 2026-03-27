package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Config represents the on-disk config file.
type Config struct {
	BaseURL string `toml:"base_url,omitempty"`
	APIKey  string `toml:"api_key,omitempty"`
	Model   string `toml:"model,omitempty"`
}

// Flags holds CLI flag overrides.
type Flags struct {
	BaseURL string
	APIKey  string
	Model   string
}

// Resolved is the final configuration ready for use.
type Resolved struct {
	BaseURL string
	APIKey  string
	Model   string
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mansplain")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mansplain")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(configDir(), "config.toml")
}

// ConfigExists returns true if the config file exists on disk.
func ConfigExists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

// Load reads the config file. Returns zero config if missing.
func Load() (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ConfigPath(), err)
	}
	return cfg, nil
}

// Save writes a Config to disk, creating directories as needed.
func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	f, err := os.Create(ConfigPath())
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// Resolve merges flags, env vars, and config file into a final config.
// Priority: flags > env vars > config file > defaults.
func Resolve(flags Flags) (*Resolved, error) {
	file, err := Load()
	if err != nil {
		return nil, err
	}

	r := &Resolved{
		BaseURL: defaultBaseURL,
	}

	// Config file
	if file.BaseURL != "" {
		r.BaseURL = file.BaseURL
	}
	if file.APIKey != "" {
		r.APIKey = file.APIKey
	}
	if file.Model != "" {
		r.Model = file.Model
	}

	// Env vars
	if v := os.Getenv("MANSPLAIN_BASE_URL"); v != "" {
		r.BaseURL = v
	}
	if v := os.Getenv("MANSPLAIN_API_KEY"); v != "" {
		r.APIKey = v
	} else if v := os.Getenv("OPENAI_API_KEY"); v != "" && r.APIKey == "" {
		r.APIKey = v
	}
	if v := os.Getenv("MANSPLAIN_MODEL"); v != "" {
		r.Model = v
	}

	// CLI flags
	if flags.BaseURL != "" {
		r.BaseURL = flags.BaseURL
	}
	if flags.APIKey != "" {
		r.APIKey = flags.APIKey
	}
	if flags.Model != "" {
		r.Model = flags.Model
	}

	if r.APIKey == "" {
		return nil, fmt.Errorf("no API key configured; run `mansplain configure` or set MANSPLAIN_API_KEY")
	}

	return r, nil
}
