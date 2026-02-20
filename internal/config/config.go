package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey string `yaml:"api_key"`
}

// Dir returns the config directory path (~/.config/truelist).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "truelist"), nil
}

// FilePath returns the full path to the config file.
func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file and merges with environment variables.
// Precedence: config file > TRUELIST_API_KEY env var.
func Load() (*Config, error) {
	cfg := &Config{}

	// Try loading from file first.
	fp, err := FilePath()
	if err == nil {
		data, readErr := os.ReadFile(fp)
		if readErr == nil {
			_ = yaml.Unmarshal(data, cfg)
		}
	}

	// Fall back to env var if config file didn't provide an API key.
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("TRUELIST_API_KEY")
	}

	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	fp := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(fp, data, 0o600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the resolved API key or an error if none is set.
func GetAPIKey() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.APIKey == "" {
		return "", fmt.Errorf("no API key configured — run `truelist config set api-key <key>` or set TRUELIST_API_KEY")
	}
	return cfg.APIKey, nil
}
