package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds user-configurable cleanup rules.
type Config struct {
	EnabledCategories    map[string]bool   `json:"enabled_categories"`
	OldInstallerMonths   int               `json:"old_installer_months"`
	LargeFileMinSizeMB   int               `json:"large_file_min_size_mb"`
	LargeFileMonths      int               `json:"large_file_months"`
	OldTempDays          int               `json:"old_temp_days"`
	OldExtensionDays     int               `json:"old_extension_days"`
	Exclusions           []string          `json:"exclusions"`
	AutoRiskOverrides    map[string]string `json:"auto_risk_overrides"`
	QuarantineMaxAgeDays int               `json:"quarantine_max_age_days"`
	Language             string            `json:"language"`
}

// Default returns the built-in default configuration.
func Default() *Config {
	return &Config{
		EnabledCategories: map[string]bool{
			"Temp":                       true,
			"Downloads":                  true,
			"Browser Cache":              true,
			"Recycle Bin":                true,
			"Logs":                       true,
			"Old Installers":             true,
			"Large Old Files":            true,
			"Thumbnails Cache":           true,
			"DirectX Shader Cache":       true,
			"Delivery Optimization":      true,
			"Windows Error Reports":      true,
			"Discord Cache":              true,
			"Steam Cache":                true,
			"VSCode Cache":               true,
			"Edge Code Cache":            true,
			"Chrome Code Cache":          true,
			"Firefox Cache2":             true,
			"Old Temp Files":             true,
			"Empty Folders":              true,
			"npm Cache":                  true,
			"pip Cache":                  true,
			"Windows Update Cache":       false,
			"Crash & Memory Dumps":       false,
			"Nvidia Installer Leftovers": false,
			"Telegram Desktop Cache":     false,
			"Old .tmp Files":             false,
			"Old .log Files":             false,
			"Old .bak Files":             false,
		},
		OldInstallerMonths: 6,
		LargeFileMinSizeMB: 100,
		LargeFileMonths:    6,
		OldTempDays:        7,
		OldExtensionDays:   30,
		Exclusions:         []string{},
		AutoRiskOverrides: map[string]string{
			".git":         "review",
			"node_modules": "review",
		},
		QuarantineMaxAgeDays: 30,
		Language:             "",
	}
}

// AppDir returns the root application data directory.
func AppDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("USERPROFILE")
	}
	return filepath.Join(localAppData, "broominal")
}

// Dir returns the directory where config and other app data live.
func Dir() string {
	return AppDir()
}

// Path returns the full path to config.json.
func Path() string {
	return filepath.Join(AppDir(), "config.json")
}

// Load reads the config from disk or returns defaults if missing.
func Load() (*Config, error) {
	p := Path()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if err := Save(cfg); err != nil {
				slog.Warn("config: failed to persist defaults", "error", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	// Merge missing categories and fields from defaults
	defaults := Default()
	if cfg.EnabledCategories == nil {
		cfg.EnabledCategories = make(map[string]bool)
	}
	for cat, val := range defaults.EnabledCategories {
		if _, ok := cfg.EnabledCategories[cat]; !ok {
			cfg.EnabledCategories[cat] = val
		}
	}
	if cfg.OldTempDays <= 0 {
		cfg.OldTempDays = defaults.OldTempDays
	}
	if cfg.OldExtensionDays <= 0 {
		cfg.OldExtensionDays = defaults.OldExtensionDays
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	tmp := Path() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, Path()); err != nil {
		return fmt.Errorf("rename config: %w", err)
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

// IsExcluded reports whether a path matches any exclusion rule.
func (c *Config) IsExcluded(path string) bool {
	lp := strings.ToLower(path)
	for _, ex := range c.Exclusions {
		lex := strings.ToLower(ex)
		// exact segment match to avoid "temp" matching "template"
		for _, seg := range strings.Split(lp, string(filepath.Separator)) {
			if seg == lex {
				return true
			}
		}
	}
	return false
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
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
