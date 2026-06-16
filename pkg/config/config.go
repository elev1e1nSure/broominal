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
	EnabledCategories  map[string]bool   `json:"enabled_categories"`
	ActivePreset       string            `json:"active_preset"`
	OldInstallerMonths int               `json:"old_installer_months"`
	LargeFileMinSizeMB int               `json:"large_file_min_size_mb"`
	LargeFileMonths    int               `json:"large_file_months"`
	Exclusions         []string          `json:"exclusions"`
	AutoRiskOverrides  map[string]string `json:"auto_risk_overrides"`
	// QuarantineEnabled controls whether cleanups move files to quarantine (true, default)
	// or permanently delete them without recovery (false).
	QuarantineEnabled bool `json:"quarantine_enabled"`
	// QuarantineAutoCleanupDays: 0 = disabled. When > 0, a Windows scheduled task runs
	// `broominal quarantine-cleanup --force --max-age-days N` daily at 03:00.
	QuarantineAutoCleanupDays int    `json:"quarantine_auto_cleanup_days"`
	Language                  string `json:"language"`
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
		OldInstallerMonths:        6,
		LargeFileMinSizeMB:        100,
		LargeFileMonths:           6,
		QuarantineEnabled:         true,
		QuarantineAutoCleanupDays: 7,
		Language:                  "",
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
	// QuarantineEnabled defaults to true; existing configs without the field get true.
	// We detect "missing from JSON" by checking if it was explicitly set to false.
	// json.Unmarshal leaves bool fields as false when absent, so we can't distinguish
	// absent-from-JSON from explicitly-false. Use the raw JSON to detect.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, exists := raw["quarantine_enabled"]; !exists {
			cfg.QuarantineEnabled = true
		}
		if _, exists := raw["quarantine_auto_cleanup_days"]; !exists {
			cfg.QuarantineAutoCleanupDays = 7
		}
	}
	return &cfg, nil
}

// Save persists the config atomically: the JSON is written to <path>.tmp and
// then renamed into place, so a crash mid-write cannot leave a half-written
// config.json that Load would refuse to parse.
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

// IsCategoryEnabled treats a nil map as "all enabled" so a default-constructed
// Config works in tests and first-run code paths without explicit init.
func (c *Config) IsCategoryEnabled(name string) bool {
	if c.EnabledCategories == nil {
		return true
	}
	return c.EnabledCategories[name]
}

// IsExcluded matches any single path component against the user's exclusion
// list. The component-wise check is what makes rules like "node_modules" or
// "dist" work without users having to spell out full directory paths.
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
