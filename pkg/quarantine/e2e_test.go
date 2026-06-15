package quarantine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestE2EQuarantineCleanup(t *testing.T) {
	// Override BaseDir to a temp directory
	origBaseDir := filepath.Dir(filepath.Dir(BaseDir()))
	tmpRoot := t.TempDir()
	os.Setenv("LOCALAPPDATA", tmpRoot)
	defer os.Setenv("LOCALAPPDATA", os.Getenv("LOCALAPPDATA"))
	_ = origBaseDir

	qDir := BaseDir()
	if err := os.MkdirAll(qDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create a quarantine entry with an old timestamp
	oldID := "2020-01-01-120000"
	oldPath := filepath.Join(qDir, oldID)
	if err := os.MkdirAll(oldPath, 0700); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(oldPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("old data"), 0644); err != nil {
		t.Fatal(err)
	}
	manifest := types.Manifest{
		ID:        oldID,
		CreatedAt: time.Now().AddDate(-1, 0, 0),
		Items: []types.ManifestItem{
			{Original: `C:\temp\test.txt`, Quarantined: testFile, Size: 8},
		},
		TotalSize: 8,
		Files:     1,
	}
	writeManifest(filepath.Join(oldPath, "manifest.json"), &manifest)

	// Create a recent quarantine entry
	recentID := time.Now().Format("2006-01-02-150405")
	recentPath := filepath.Join(qDir, recentID)
	if err := os.MkdirAll(recentPath, 0700); err != nil {
		t.Fatal(err)
	}
	recentFile := filepath.Join(recentPath, "recent.txt")
	if err := os.WriteFile(recentFile, []byte("recent"), 0644); err != nil {
		t.Fatal(err)
	}
	recentManifest := types.Manifest{
		ID:        recentID,
		CreatedAt: time.Now(),
		Items: []types.ManifestItem{
			{Original: `C:\temp\recent.txt`, Quarantined: recentFile, Size: 6},
		},
		TotalSize: 6,
		Files:     1,
	}
	writeManifest(filepath.Join(recentPath, "manifest.json"), &recentManifest)

	// Cleanup: remove entries older than 30 days
	deleted, freed, err := Cleanup(30)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("deleted %d quarantines, freed %d bytes", deleted, freed)
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	// Verify old dir is gone, recent dir remains
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old quarantine dir should be deleted")
	}
	if _, err := os.Stat(recentPath); err != nil {
		t.Error("recent quarantine dir should still exist")
	}

	// Verify listing: only recent entry
	ids, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Errorf("expected 1 quarantine after cleanup, got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && !strings.HasPrefix(ids[0], time.Now().Format("2006-01-02")) {
		t.Errorf("remaining entry has wrong date prefix: %s", ids[0])
	}
}
