package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.OldInstallerMonths != 6 {
		t.Errorf("OldInstallerMonths = %d, want 6", cfg.OldInstallerMonths)
	}
	if cfg.LargeFileMinSizeMB != 100 {
		t.Errorf("LargeFileMinSizeMB = %d, want 100", cfg.LargeFileMinSizeMB)
	}
	if cfg.LargeFileMonths != 6 {
		t.Errorf("LargeFileMonths = %d, want 6", cfg.LargeFileMonths)
	}
	if cfg.QuarantineMaxAgeDays != 30 {
		t.Errorf("QuarantineMaxAgeDays = %d, want 30", cfg.QuarantineMaxAgeDays)
	}
	if cfg.ActivePreset != string(PresetQuick) {
		t.Errorf("ActivePreset = %q, want %q", cfg.ActivePreset, string(PresetQuick))
	}
	for _, cat := range []string{"Temp", "Browser Cache", "Edge Code Cache", "Chrome Code Cache", "Firefox Cache2", "Windows Prefetch", "AMD GPU Cache"} {
		if !cfg.EnabledCategories[cat] {
			t.Errorf("%q should be enabled by default (Quick preset)", cat)
		}
	}
	for _, cat := range []string{"Downloads", "Recycle Bin", "Messenger Cache", "Zoom Cache", "Windows Defender"} {
		if cfg.EnabledCategories[cat] {
			t.Errorf("%q should be disabled by default (Quick preset)", cat)
		}
	}
}

func TestIsCategoryEnabled(t *testing.T) {
	cfg := &Config{EnabledCategories: map[string]bool{"Temp": true, "Downloads": false}}
	if !cfg.IsCategoryEnabled("Temp") {
		t.Error("Temp should be enabled")
	}
	if cfg.IsCategoryEnabled("Downloads") {
		t.Error("Downloads should be disabled")
	}
	if cfg.IsCategoryEnabled("Missing") {
		t.Error("Missing category with existing map should default to false")
	}

	nilCfg := &Config{EnabledCategories: nil}
	if !nilCfg.IsCategoryEnabled("Anything") {
		t.Error("nil map should default to true")
	}
}

func TestRiskOverrideFor(t *testing.T) {
	cfg := &Config{
		AutoRiskOverrides: map[string]string{
			".git":         "review",
			"node_modules": "review",
		},
	}
	tests := []struct {
		path string
		want string
	}{
		{`C:\project\.git\config`, "review"},
		{`C:\project\.GIT\config`, "review"},
		{`C:\app\node_modules\pkg`, "review"},
		{`C:\app\NODE_MODULES\pkg`, "review"},
		{`C:\safe\file.txt`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := cfg.RiskOverrideFor(tt.path)
			if got != tt.want {
				t.Errorf("RiskOverrideFor(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s, substr string
		want      bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"abc", "abcd", false},
		{"", "a", false},
		{"a", "", true},
	}
	for _, tt := range tests {
		got := containsIgnoreCase(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestLoadSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	// No config exists — Load should return defaults and persist them
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.OldInstallerMonths != 6 {
		t.Errorf("unexpected default: %d", cfg.OldInstallerMonths)
	}

	// Verify file was created
	if _, err := os.Stat(Path()); os.IsNotExist(err) {
		t.Fatal("config file should be auto-created")
	}

	// Modify and Save
	cfg.OldInstallerMonths = 12
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload and verify
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}
	if cfg2.OldInstallerMonths != 12 {
		t.Errorf("OldInstallerMonths = %d, want 12", cfg2.OldInstallerMonths)
	}
}

func TestLoadCorrupt(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	// Write invalid JSON
	_ = os.MkdirAll(Dir(), 0755)
	if err := os.WriteFile(Path(), []byte("not json"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for corrupt config")
	}
}

func TestLoadMissingDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	_, err := Load()
	if err != nil {
		t.Fatalf("Load with missing dir should return defaults: %v", err)
	}
}

func TestPathAndDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	wantDir := filepath.Join(tmp, "broominal")
	if Dir() != wantDir {
		t.Errorf("Dir() = %q, want %q", Dir(), wantDir)
	}
	wantPath := filepath.Join(wantDir, "config.json")
	if Path() != wantPath {
		t.Errorf("Path() = %q, want %q", Path(), wantPath)
	}
}

func TestDirPermissions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	cfg := Default()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(Dir())
	if err != nil {
		t.Fatalf("stat config dir: %v", err)
	}
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0700 {
			t.Errorf("config dir permissions = %04o, want 0700", perm)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	cfg := &Config{
		Exclusions: []string{"node_modules", "temp"},
	}
	tests := []struct {
		path string
		want bool
	}{
		{`C:\project\node_modules\pkg`, true},
		{`C:\project\NODE_MODULES\pkg`, true},
		{`C:\temp\file.txt`, true},
		{`C:\template\file.txt`, false},
		{`C:\safe\file.txt`, false},
	}
	for _, tt := range tests {
		got := cfg.IsExcluded(tt.path)
		if got != tt.want {
			t.Errorf("IsExcluded(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestAppDirFallback(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", "")
	t.Setenv("USERPROFILE", tmp)

	want := filepath.Join(tmp, "broominal")
	if AppDir() != want {
		t.Errorf("AppDir() = %q, want %q", AppDir(), want)
	}
}

func TestLoadReadError(t *testing.T) {
	// Create a config path that is a directory instead of a file
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	_ = os.MkdirAll(Path(), 0755)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when config path is a directory")
	}
}

func TestLoadInvalidThresholds(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	bad := &Config{
		EnabledCategories:    map[string]bool{"Temp": true},
		OldInstallerMonths:   -1,
		LargeFileMinSizeMB:   0,
		LargeFileMonths:      -5,
		OldTempDays:          0,
		OldExtensionDays:     -1,
		QuarantineMaxAgeDays: 0,
	}
	_ = os.MkdirAll(Dir(), 0755)
	data, _ := json.Marshal(bad)
	_ = os.WriteFile(Path(), data, 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.OldInstallerMonths <= 0 {
		t.Errorf("OldInstallerMonths = %d, want > 0", cfg.OldInstallerMonths)
	}
	if cfg.LargeFileMinSizeMB <= 0 {
		t.Errorf("LargeFileMinSizeMB = %d, want > 0", cfg.LargeFileMinSizeMB)
	}
	if cfg.LargeFileMonths <= 0 {
		t.Errorf("LargeFileMonths = %d, want > 0", cfg.LargeFileMonths)
	}
	if cfg.OldTempDays <= 0 {
		t.Errorf("OldTempDays = %d, want > 0", cfg.OldTempDays)
	}
	if cfg.OldExtensionDays <= 0 {
		t.Errorf("OldExtensionDays = %d, want > 0", cfg.OldExtensionDays)
	}
	if cfg.QuarantineMaxAgeDays <= 0 {
		t.Errorf("QuarantineMaxAgeDays = %d, want > 0", cfg.QuarantineMaxAgeDays)
	}
}
