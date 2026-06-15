package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestE2EScanCleanRestore(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test-file.txt")
	content := []byte("broominal e2e test 2026")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	prevTemp := os.Getenv("TEMP")
	os.Setenv("TEMP", tmpDir)
	defer os.Setenv("TEMP", prevTemp)

	cfg := config.Default()
	cfg.EnabledCategories = map[string]bool{"Temp": true}
	cfg.QuarantineEnabled = true

	ctx := context.Background()
	res, err := scanner.ScanWithConfig(ctx, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalSize == 0 {
		t.Fatal("scan found no items")
	}
	t.Logf("scanned %d bytes", res.TotalSize)

	var items []types.Item
	for _, cat := range res.Categories {
		for i := range cat.Items {
			cat.Items[i].Selected = true
			items = append(items, cat.Items[i])
		}
	}

	cleanRes, err := Run(ctx, items, res, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cleaned %d files, restore ID: %s", cleanRes.Files, cleanRes.RestoreID)

	if _, err := os.Stat(srcFile); err == nil {
		t.Error("source file still exists after clean")
	}

	restored, skipped, err := quarantine.Restore(cleanRes.RestoreID, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("restored %d, skipped %d", restored, skipped)

	data, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("file not restored: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(data), string(content))
	}
}
