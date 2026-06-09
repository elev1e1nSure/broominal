package quarantine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestUniquePath(t *testing.T) {
	tmp := t.TempDir()
	p1 := uniquePath(tmp, "file.txt")
	if p1 != filepath.Join(tmp, "file.txt") {
		t.Errorf("first uniquePath = %q", p1)
	}
	_ = os.WriteFile(p1, []byte("a"), 0644)

	p2 := uniquePath(tmp, "file.txt")
	if p2 != filepath.Join(tmp, "file_1.txt") {
		t.Errorf("second uniquePath = %q, want %q", p2, filepath.Join(tmp, "file_1.txt"))
	}
	_ = os.WriteFile(p2, []byte("b"), 0644)

	p3 := uniquePath(tmp, "file.txt")
	if p3 != filepath.Join(tmp, "file_2.txt") {
		t.Errorf("third uniquePath = %q, want %q", p3, filepath.Join(tmp, "file_2.txt"))
	}
}

func TestCopyAndDelete(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")
	content := []byte("hello quarantine")
	_ = os.WriteFile(src, content, 0644)

	if err := copyAndDelete(src, dst); err != nil {
		t.Fatalf("copyAndDelete failed: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("dst content = %q, want %q", got, content)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("src should be deleted")
	}
}

func TestMoveDryRun(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	src := filepath.Join(tmp, "keep.txt")
	_ = os.WriteFile(src, []byte("data"), 0644)

	items := []types.Item{{Path: src, Size: 4, Selected: true}}
	id, freed, files, err := Move(items, true)
	if err != nil {
		t.Fatalf("Move dry-run failed: %v", err)
	}
	if id == "" {
		t.Errorf("dry-run id = %q, want non-empty", id)
	}
	if freed != 4 {
		t.Errorf("dry-run freed = %d, want 4", freed)
	}
	if files != 1 {
		t.Errorf("dry-run files = %d, want 1", files)
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		t.Error("dry-run should not move files")
	}
}

func TestMoveReal(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	srcDir := filepath.Join(tmp, "source")
	_ = os.MkdirAll(srcDir, 0755)
	src1 := filepath.Join(srcDir, "a.txt")
	src2 := filepath.Join(srcDir, "b.txt")
	_ = os.WriteFile(src1, []byte("aa"), 0644)
	_ = os.WriteFile(src2, []byte("bb"), 0644)

	items := []types.Item{
		{Path: src1, Size: 2, Selected: true},
		{Path: src2, Size: 2, Selected: true},
	}

	id, freed, files, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id")
	}
	if freed != 4 {
		t.Errorf("freed = %d, want 4", freed)
	}
	if files != 2 {
		t.Errorf("files = %d, want 2", files)
	}

	// Files should be moved
	if _, err := os.Stat(src1); !os.IsNotExist(err) {
		t.Error("src1 should be moved")
	}
	if _, err := os.Stat(src2); !os.IsNotExist(err) {
		t.Error("src2 should be moved")
	}

	// Manifest should exist
	manifestPath := filepath.Join(BaseDir(), id, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest should exist")
	}
}

func TestMoveMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	items := []types.Item{
		{Path: filepath.Join(tmp, "missing.txt"), Size: 10, Selected: true},
	}
	_, freed, files, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if freed != 0 {
		t.Errorf("freed = %d, want 0", freed)
	}
	if files != 0 {
		t.Errorf("files = %d, want 0", files)
	}
}

func TestMoveDuplicateNames(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	srcDir1 := filepath.Join(tmp, "dir1")
	srcDir2 := filepath.Join(tmp, "dir2")
	_ = os.MkdirAll(srcDir1, 0755)
	_ = os.MkdirAll(srcDir2, 0755)
	f1 := filepath.Join(srcDir1, "same.txt")
	f2 := filepath.Join(srcDir2, "same.txt")
	_ = os.WriteFile(f1, []byte("a"), 0644)
	_ = os.WriteFile(f2, []byte("b"), 0644)

	items := []types.Item{
		{Path: f1, Size: 1, Selected: true},
		{Path: f2, Size: 1, Selected: true},
	}

	id, _, _, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	qDir := filepath.Join(BaseDir(), id)
	entries, _ := os.ReadDir(qDir)
	var names []string
	for _, e := range entries {
		if e.Name() != "manifest.json" {
			names = append(names, e.Name())
		}
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 quarantined files, got %d", len(names))
	}
	// One should be same.txt, other same_1.txt
	hasSame := false
	hasSame1 := false
	for _, n := range names {
		if n == "same.txt" {
			hasSame = true
		}
		if n == "same_1.txt" {
			hasSame1 = true
		}
	}
	if !hasSame || !hasSame1 {
		t.Errorf("names = %v, want same.txt and same_1.txt", names)
	}
}

func TestRestoreHappyPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	src := filepath.Join(tmp, "orig.txt")
	_ = os.WriteFile(src, []byte("data"), 0644)

	items := []types.Item{{Path: src, Size: 4, Selected: true}}
	id, _, _, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	restored, skipped, err := Restore(id, false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored != 1 {
		t.Errorf("restored = %d, want 1", restored)
	}
	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		t.Error("original file should be restored")
	}
	// Quarantine dir should be removed
	qDir := filepath.Join(BaseDir(), id)
	if _, err := os.Stat(qDir); !os.IsNotExist(err) {
		t.Error("quarantine dir should be removed after full restore")
	}
}

func TestRestoreConflictSkip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	src := filepath.Join(tmp, "orig.txt")
	_ = os.WriteFile(src, []byte("original"), 0644)

	items := []types.Item{{Path: src, Size: 4, Selected: true}}
	id, _, _, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Recreate original file before restore
	_ = os.WriteFile(src, []byte("new content"), 0644)

	restored, skipped, err := Restore(id, false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored != 0 {
		t.Errorf("restored = %d, want 0", restored)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}

	// Manifest should still contain the skipped item
	manifestPath := filepath.Join(BaseDir(), id, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest should remain for partial restore")
	}
}

func TestRestoreForceOverwrite(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	src := filepath.Join(tmp, "orig.txt")
	_ = os.WriteFile(src, []byte("original"), 0644)

	items := []types.Item{{Path: src, Size: 4, Selected: true}}
	id, _, _, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	_ = os.WriteFile(src, []byte("new content"), 0644)

	restored, skipped, err := Restore(id, true)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored != 1 {
		t.Errorf("restored = %d, want 1", restored)
	}
	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}

	data, _ := os.ReadFile(src)
	if string(data) != "original" {
		t.Errorf("restored content = %q, want original", data)
	}
}

func TestCheckRestoreConflicts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	src := filepath.Join(tmp, "orig.txt")
	_ = os.WriteFile(src, []byte("data"), 0644)

	items := []types.Item{{Path: src, Size: 4, Selected: true}}
	id, _, _, err := Move(items, false)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// No conflict yet
	conflicts, err := CheckRestoreConflicts(id)
	if err != nil {
		t.Fatalf("CheckRestoreConflicts failed: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("conflicts = %d, want 0", len(conflicts))
	}

	// Recreate original
	_ = os.WriteFile(src, []byte("new"), 0644)
	conflicts, err = CheckRestoreConflicts(id)
	if err != nil {
		t.Fatalf("CheckRestoreConflicts failed: %v", err)
	}
	if len(conflicts) != 1 {
		t.Errorf("conflicts = %d, want 1", len(conflicts))
	}
}

func TestListAndGetLast(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	// Empty
	ids, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 ids, got %d", len(ids))
	}

	_, err = GetLast()
	if err == nil {
		t.Fatal("GetLast on empty should error")
	}

	// Create two quarantine dirs with different mod times
	qDir := BaseDir()
	_ = os.MkdirAll(qDir, 0755)
	_ = os.MkdirAll(filepath.Join(qDir, "id-1"), 0755)
	_ = os.Chtimes(filepath.Join(qDir, "id-1"), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
	_ = os.MkdirAll(filepath.Join(qDir, "id-2"), 0755)
	_ = os.Chtimes(filepath.Join(qDir, "id-2"), time.Now(), time.Now())

	ids, err = List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}

	last, err := GetLast()
	if err != nil {
		t.Fatalf("GetLast failed: %v", err)
	}
	if last != "id-2" {
		t.Errorf("GetLast = %q, want id-2", last)
	}
}

func TestCleanupEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	deleted, freed, err := Cleanup(30, false)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("deleted = %d, want 0", deleted)
	}
	if freed != 0 {
		t.Errorf("freed = %d, want 0", freed)
	}
}

func TestCleanupRemovesOld(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := BaseDir()
	oldDir := filepath.Join(qDir, "old-id")
	_ = os.MkdirAll(oldDir, 0755)
	_ = os.WriteFile(filepath.Join(oldDir, "file.txt"), []byte("data"), 0644)
	// Write a valid manifest with old date
	manifest := `{"id":"old-id","created_at":"2020-01-01T00:00:00Z","items":[]}`
	_ = os.WriteFile(filepath.Join(oldDir, "manifest.json"), []byte(manifest), 0644)

	deleted, freed, err := Cleanup(30, false)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}
	if freed == 0 {
		t.Error("expected freed > 0")
	}
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old quarantine should be removed")
	}
}

func TestCleanupDryRun(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := BaseDir()
	oldDir := filepath.Join(qDir, "old-id")
	_ = os.MkdirAll(oldDir, 0755)
	manifest := `{"id":"old-id","created_at":"2020-01-01T00:00:00Z","items":[]}`
	_ = os.WriteFile(filepath.Join(oldDir, "manifest.json"), []byte(manifest), 0644)

	deleted, freed, err := Cleanup(30, true)
	if err != nil {
		t.Fatalf("Cleanup dry-run failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}
	// Dir should still exist after dry-run
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error("dry-run should not remove quarantine")
	}
	_ = freed
}

func TestCleanupKeepsRecent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	qDir := BaseDir()
	recentDir := filepath.Join(qDir, "recent-id")
	_ = os.MkdirAll(recentDir, 0755)
	manifest := `{"id":"recent-id","created_at":"` + time.Now().Format(time.RFC3339) + `","items":[]}`
	_ = os.WriteFile(filepath.Join(recentDir, "manifest.json"), []byte(manifest), 0644)

	deleted, freed, err := Cleanup(30, false)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("deleted = %d, want 0", deleted)
	}
	if freed != 0 {
		t.Errorf("freed = %d, want 0", freed)
	}
	if _, err := os.Stat(recentDir); os.IsNotExist(err) {
		t.Error("recent quarantine should be kept")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"12345678-1234-1234-1234-123456789abc", true},
		{"simple-name", true},
		{"foo..bar", true},
		{"..", false},
		{`..\foo`, false},
		{`../foo`, false},
		{`C:\foo`, false},
	}
	for _, tt := range tests {
		err := validateID(tt.id)
		if tt.valid && err != nil {
			t.Errorf("validateID(%q) = %v, want nil", tt.id, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("validateID(%q) = nil, want error", tt.id)
		}
	}
}

func TestRestoreInvalidID(t *testing.T) {
	_, _, err := Restore(`..\foo`, false)
	if err == nil {
		t.Fatal("expected error for invalid restore id")
	}
}

func TestCheckRestoreConflictsInvalidID(t *testing.T) {
	_, err := CheckRestoreConflicts(`..\foo`)
	if err == nil {
		t.Fatal("expected error for invalid restore id")
	}
}

func TestRestoreManifestPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("TEMP", tmp)
	t.Setenv("USERPROFILE", tmp)

	secretFile := filepath.Join(tmp, "secret.txt")
	_ = os.WriteFile(secretFile, []byte("secret"), 0644)

	qDir := filepath.Join(BaseDir(), "test-id")
	_ = os.MkdirAll(qDir, 0755)

	qFile := filepath.Join(qDir, "q.txt")
	_ = os.WriteFile(qFile, []byte("stolen"), 0644)

	manifest := `{"id":"test-id","created_at":"2020-01-01T00:00:00Z","items":[{"original":"../secret.txt","quarantined":` + fmt.Sprintf("%q", qFile) + `,"size":6}]}`
	_ = os.WriteFile(filepath.Join(qDir, "manifest.json"), []byte(manifest), 0644)

	restored, _, err := Restore("test-id", false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored != 0 {
		t.Errorf("restored = %d, want 0", restored)
	}

	data, _ := os.ReadFile(secretFile)
	if string(data) != "secret" {
		t.Errorf("secret was overwritten: %q", data)
	}
}

func TestRestoreManifestDisallowedPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	// Set all allowed roots to dummy paths so that C:\evil is disallowed
	t.Setenv("TEMP", filepath.Join(tmp, "temp"))
	t.Setenv("TMP", filepath.Join(tmp, "tmp"))
	t.Setenv("USERPROFILE", filepath.Join(tmp, "profile"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "local"))
	t.Setenv("APPDATA", filepath.Join(tmp, "roaming"))
	t.Setenv("ProgramData", filepath.Join(tmp, "programdata"))
	t.Setenv("SystemRoot", filepath.Join(tmp, "windows"))
	t.Setenv("WINDIR", filepath.Join(tmp, "windows"))
	t.Setenv("ProgramFiles", filepath.Join(tmp, "pf"))
	t.Setenv("ProgramFiles(x86)", filepath.Join(tmp, "pf86"))
	t.Setenv("SYSTEMDRIVE", "Q:")

	qDir := filepath.Join(BaseDir(), "test-id2")
	_ = os.MkdirAll(qDir, 0755)

	qFile := filepath.Join(qDir, "q.txt")
	_ = os.WriteFile(qFile, []byte("stolen"), 0644)

	outsideFile := `C:\evil\target.txt`
	manifest := fmt.Sprintf(`{"id":"test-id2","created_at":"2020-01-01T00:00:00Z","items":[{"original":%q,"quarantined":%q,"size":6}]}`, outsideFile, qFile)
	_ = os.WriteFile(filepath.Join(qDir, "manifest.json"), []byte(manifest), 0644)

	restored, _, err := Restore("test-id2", false)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restored != 0 {
		t.Errorf("restored = %d, want 0", restored)
	}
}
