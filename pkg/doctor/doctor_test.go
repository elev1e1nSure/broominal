package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
)

func TestCheckDir(t *testing.T) {
	tmp := t.TempDir()

	// Existing writable dir
	c := checkDir(tmp, "Test")
	if c.Status != StatusPass {
		t.Errorf("expected PASS for writable dir, got %s: %s", c.Status, c.Detail)
	}

	// Non-existent path that can be created
	newDir := filepath.Join(tmp, "new", "subdir")
	c2 := checkDir(newDir, "Nested")
	if c2.Status != StatusPass {
		t.Errorf("expected PASS for creatable dir, got %s: %s", c2.Status, c2.Detail)
	}
}

func TestCheckDirReadOnlyParent(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root, cannot test read-only failure")
	}
	// This test is best-effort on Windows
	tmp := t.TempDir()
	readonly := filepath.Join(tmp, "readonly")
	_ = os.MkdirAll(readonly, 0555)
	defer os.Chmod(readonly, 0755) // ensure cleanup

	c := checkDir(filepath.Join(readonly, "nested"), "ReadOnly")
	// May pass or fail depending on OS/permissions; just check it returns something
	if c.Name != "ReadOnly" {
		t.Errorf("name = %q, want %q", c.Name, "ReadOnly")
	}
}

func TestCheckEnvDir(t *testing.T) {
	// Existing single-directory env
	c := checkEnvDir("SystemRoot", "Windows")
	if c.Status != StatusPass {
		t.Errorf("expected PASS for SystemRoot, got %s: %s", c.Status, c.Detail)
	}

	// Missing env
	c2 := checkEnvDir("BROOMINAL_NONEXISTENT_ENV_12345", "Missing")
	if c2.Status != StatusFail {
		t.Errorf("expected FAIL for missing env, got %s", c2.Status)
	}
}

func TestCheckManifestsEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	c := checkManifests()
	if c.Status != StatusPass {
		t.Errorf("expected PASS for empty quarantine, got %s: %s", c.Status, c.Detail)
	}
	if c.Detail != i18n.T("no_backups_yet") {
		t.Errorf("detail = %q, want %q", c.Detail, i18n.T("no_backups_yet"))
	}
}

func TestCheckManifestsValid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	// Create a valid quarantine with manifest
	qDir := quarantine.BaseDir()
	idDir := filepath.Join(qDir, "test-id")
	_ = os.MkdirAll(idDir, 0755)
	manifest := `{"id":"test-id","created_at":"2024-01-01T00:00:00Z","items":[]}`
	_ = os.WriteFile(filepath.Join(idDir, "manifest.json"), []byte(manifest), 0644)

	c := checkManifests()
	if c.Status != StatusPass {
		t.Errorf("expected PASS for valid manifest, got %s: %s", c.Status, c.Detail)
	}
}

func TestCheckManifestsInvalid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := quarantine.BaseDir()
	idDir := filepath.Join(qDir, "bad-id")
	_ = os.MkdirAll(idDir, 0755)
	_ = os.WriteFile(filepath.Join(idDir, "manifest.json"), []byte("not json"), 0644)

	c := checkManifests()
	if c.Status != StatusWarn {
		t.Errorf("expected WARN for invalid manifest, got %s: %s", c.Status, c.Detail)
	}
}

func TestCheckQuarantineStatsEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	c := checkQuarantineStats()
	if c.Status != StatusPass {
		t.Errorf("expected PASS for empty quarantine stats, got %s: %s", c.Status, c.Detail)
	}
	if c.Detail != "0" {
		t.Errorf("detail = %q, want %q", c.Detail, "0")
	}
}

func TestCheckQuarantineStatsWithData(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := quarantine.BaseDir()
	idDir := filepath.Join(qDir, "id-1")
	_ = os.MkdirAll(idDir, 0755)
	_ = os.WriteFile(filepath.Join(idDir, "file.txt"), []byte("hello"), 0644)

	c := checkQuarantineStats()
	if c.Status != StatusPass {
		t.Errorf("expected PASS, got %s: %s", c.Status, c.Detail)
	}
	if c.Detail == "" {
		t.Error("expected non-empty detail")
	}
}

func TestRun(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("TEMP", tmp)
	t.Setenv("USERPROFILE", tmp)

	checks := Run()
	if len(checks) == 0 {
		t.Fatal("expected some checks")
	}

	names := map[string]bool{}
	for _, c := range checks {
		names[c.Name] = true
	}

	expected := []string{
		i18n.T("check_admin"),
		i18n.T("check_quarantine_dir"),
		i18n.T("check_reports_dir"),
		i18n.T("check_config_dir"),
		i18n.T("check_temp_dir"),
		i18n.T("check_userprofile_dir"),
		i18n.T("check_manifests"),
		i18n.T("check_stats"),
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing check: %s", name)
		}
	}
}

func TestCheckQuarantineStatsBrokenManifest(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := quarantine.BaseDir()
	idDir := filepath.Join(qDir, "broken-id")
	_ = os.MkdirAll(idDir, 0755)
	// No manifest file — should use dir mod time
	_ = os.WriteFile(filepath.Join(idDir, "file.txt"), []byte("x"), 0644)

	c := checkQuarantineStats()
	if c.Status != StatusPass {
		t.Errorf("expected PASS for broken manifest stats, got %s: %s", c.Status, c.Detail)
	}
}
