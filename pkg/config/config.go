package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds user-configurable cleanup rules.
type Config struct {
	EnabledCategories    map[string]bool   `json:"enabled_categories"`
	OldInstallerMonths   int               `json:"old_installer_months"`
	LargeFileMinSizeMB   int               `json:"large_file_min_size_mb"`
	LargeFileMonths      int               `json:"large_file_months"`
	Exclusions           []string          `json:"exclusions"`
	AutoRiskOverrides    map[string]string `json:"auto_risk_overrides"`
	QuarantineMaxAgeDays int               `json:"quarantine_max_age_days"`
}

// Default returns the built-in default configuration.
func Default() *Config {
	return &Config{
		EnabledCategories: map[string]bool{
			"Temp":           true,
			"Downloads":      true,
			"Browser Cache":  true,
			"Recycle Bin":    true,
			"Logs":           true,
			"Old Installers": true,
			"Large Old Files": true,
		},
		OldInstallerMonths: 6,
		LargeFileMinSizeMB: 100,
		LargeFileMonths:    6,
		Exclusions:         []string{},
		AutoRiskOverrides: map[string]string{
			".git":          "review",
			"node_modules":  "review",
		},
		QuarantineMaxAgeDays: 30,
	}
}

// Dir returns the directory where config and other app data live.
func Dir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("USERPROFILE")
	}
	return filepath.Join(localAppData, "broominal")
}

// Path returns the full path to config.json.
func Path() string {
	return filepath.Join(Dir(), "config.json")
}

// Load reads the config from disk or returns defaults if missing.
func Load() (*Config, error) {
	p := Path()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			_ = Save(cfg) // try to persist defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(Path(), data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// IsCategoryEnabled reports whether a category is enabled.
func (c *Config) IsCategoryEnabled(name string) bool {
	if c.EnabledCategories == nil {
		return true
	}
	return c.EnabledCategories[name]
}

// RiskOverrideFor returns an override risk for a path, or empty string.
func (c *Config) RiskOverrideFor(path string) string {
	for substr, risk := range c.AutoRiskOverrides {
		if containsIgnoreCase(path, substr) {
			return risk
		}
	}
	return ""
}

func containsIgnoreCase(s, substr string) bool {
	return len(substr) <= len(s) && (containsAt(s, substr, 0) || containsFrom(s, substr, 1))
}

func containsAt(s, substr string, start int) bool {
	for i := 0; i < len(substr); i++ {
		if start+i >= len(s) {
			return false
		}
		if toLower(s[start+i]) != toLower(substr[i]) {
			return false
		}
	}
	return true
}

func containsFrom(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if containsAt(s, substr, i) {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
