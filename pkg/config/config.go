package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/categories"
)

// Config holds user-configurable cleanup rules.
type Config struct {
	EnabledCategories    map[string]bool   `json:"enabled_categories"`
	ActivePreset         string            `json:"active_preset"`
	OldInstallerMonths   int               `json:"old_installer_months"`
	LargeFileMinSizeMB   int               `json:"large_file_min_size_mb"`
	LargeFileMonths      int               `json:"large_file_months"`
	Exclusions           []string          `json:"exclusions"`
	AutoRiskOverrides    map[string]string `json:"auto_risk_overrides"`
	QuarantineMaxAgeDays int               `json:"quarantine_max_age_days"`
	Language             string            `json:"language"`
}

// Default returns the built-in default configuration.
func Default() *Config {
	ec := make(map[string]bool, len(categories.All))
	for _, def := range categories.All {
		ec[def.Name] = def.MinPreset == categories.Quick
	}
	return &Config{
		EnabledCategories: ec,
		ActivePreset:      string(PresetQuick),
		Exclusions:        []string{},
		AutoRiskOverrides: map[string]string{
			".git":         "review",
			"node_modules": "review",
		},
		OldInstallerMonths:   6,
		LargeFileMinSizeMB:   100,
		LargeFileMonths:      6,
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
	if cfg.OldInstallerMonths <= 0 {
		cfg.OldInstallerMonths = defaults.OldInstallerMonths
	}
	if cfg.LargeFileMinSizeMB <= 0 {
		cfg.LargeFileMinSizeMB = defaults.LargeFileMinSizeMB
	}
	if cfg.LargeFileMonths <= 0 {
		cfg.LargeFileMonths = defaults.LargeFileMonths
	}
	if cfg.QuarantineMaxAgeDays <= 0 {
		cfg.QuarantineMaxAgeDays = defaults.QuarantineMaxAgeDays
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
	segments := strings.Split(lp, string(filepath.Separator))
	segSet := make(map[string]struct{}, len(segments))
	for _, seg := range segments {
		segSet[seg] = struct{}{}
	}
	for _, ex := range c.Exclusions {
		if _, ok := segSet[strings.ToLower(ex)]; ok {
			return true
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

type Preset string

const (
	PresetQuick    Preset = "quick"
	PresetStandard Preset = "standard"
	PresetDeep     Preset = "deep"
)

func (c *Config) ApplyPreset(p Preset) {
	c.EnabledCategories = make(map[string]bool)
	var level categories.Preset
	switch p {
	case PresetQuick:
		level = categories.Quick
	case PresetStandard:
		level = categories.Standard
	case PresetDeep:
		level = categories.Deep
	}
	for _, def := range categories.All {
		if def.MinPreset <= level {
			c.EnabledCategories[def.Name] = true
		}
	}
	c.ActivePreset = string(p)
}
