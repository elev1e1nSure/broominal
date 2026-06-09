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
			// Safe (default)
			"Temp":                       true,
			"Browser Cache":              true,
			"Thumbnails Cache":           true,
			"DirectX Shader Cache":       true,
			"Empty Folders":              true,
			"Delivery Optimization":      true,
			"Icon Cache":                 true,
			"Windows Error Reports":      true,
			"Opera Cache":                true,
			"Brave Cache":                true,
			"Vivaldi Cache":              true,
			"Yandex Cache":               true,
			// Normal
			"Downloads":                  false,
			"Recycle Bin":                false,
			"Logs":                       false,
			"Old Installers":             false,
			"Large Old Files":            false,
			"Messenger Cache":            false,
			"Steam Cache":                false,
			"VSCode Cache":               false,
			"Edge Code Cache":            false,
			"Chrome Code Cache":          false,
			"Firefox Cache2":             false,
			"npm Cache":                  false,
			"pip Cache":                  false,
			"Spotify Cache":              false,
			"OneDrive Cache":             false,
			"Visual Studio Cache":        false,
			"Git Cache":                  false,
			"Windows Prefetch":           false,
			"Windows Update Cache":       false,
			"Crash & Memory Dumps":       false,
			"Nvidia Installer Leftovers": false,
			"Office Cache":               false,
			"OBS Cache":                  false,
			"TeamViewer Logs":            false,
			"Epic Games Cache":           false,
			"Battle.net Cache":           false,
			"Rockstar Cache":             false,
			"EA App Cache":               false,
			"Ubisoft Cache":              false,
			"GOG Galaxy Cache":           false,
			// Hard
			"Docker Cache":               false,
			"JetBrains Cache":            false,
			"Go Build Cache":             false,
			"Rust Cache":                 false,
			"NuGet Cache":                false,
			"Unity Cache":                false,
			"Adobe Cache":                false,
			"Windows Defender":           false,
		},
		Exclusions: []string{},
		AutoRiskOverrides: map[string]string{
			".git":         "review",
			"node_modules": "review",
		},
		OldInstallerMonths:   6,
		LargeFileMinSizeMB:   100,
		LargeFileMonths:      6,
		OldTempDays:          7,
		OldExtensionDays:     30,
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
	if cfg.OldTempDays <= 0 {
		cfg.OldTempDays = defaults.OldTempDays
	}
	if cfg.OldExtensionDays <= 0 {
		cfg.OldExtensionDays = defaults.OldExtensionDays
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
	PresetSafe   Preset = "safe"
	PresetNormal Preset = "normal"
	PresetHard   Preset = "hard"
)

func (c *Config) ApplyPreset(p Preset) {
	c.EnabledCategories = make(map[string]bool)

	safeCategories := []string{
		"Temp",
		"Browser Cache",
		"Thumbnails Cache",
		"DirectX Shader Cache",
		"Empty Folders",
		"Delivery Optimization",
		"Icon Cache",
		"Windows Error Reports",
		"Opera Cache",
		"Brave Cache",
		"Vivaldi Cache",
		"Yandex Cache",
	}

	normalCategories := append(safeCategories,
		"Logs",
		"Steam Cache",
		"Messenger Cache",
		"VSCode Cache",
		"Edge Code Cache",
		"Chrome Code Cache",
		"Firefox Cache2",
		"npm Cache",
		"pip Cache",
		"Spotify Cache",
		"OneDrive Cache",
		"Git Cache",
		"Windows Prefetch",
		"Office Cache",
		"OBS Cache",
		"TeamViewer Logs",
		"Epic Games Cache",
		"Battle.net Cache",
		"Rockstar Cache",
		"EA App Cache",
		"Ubisoft Cache",
		"GOG Galaxy Cache",
	)

	hardCategories := append(normalCategories,
		"Docker Cache",
		"JetBrains Cache",
		"Go Build Cache",
		"Rust Cache",
		"NuGet Cache",
		"Unity Cache",
		"Visual Studio Cache",
		"Adobe Cache",
		"Downloads",
		"Old Installers",
		"Large Old Files",
		"Windows Update Cache",
		"Crash & Memory Dumps",
		"Nvidia Installer Leftovers",
		"Recycle Bin",
		"Windows Defender",
	)

	var categories []string
	switch p {
	case PresetSafe:
		categories = safeCategories
	case PresetNormal:
		categories = normalCategories
	case PresetHard:
		categories = hardCategories
	}

	for _, cat := range categories {
		c.EnabledCategories[cat] = true
	}
}
