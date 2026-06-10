package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestScanCleanRestoreDoctorPipeline(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("TEMP", tmp)

	// Create some temp files for scanning
	for i := 0; i < 3; i++ {
		p := filepath.Join(tmp, fmt.Sprintf("junk%d.log", i))
		_ = os.WriteFile(p, []byte("junk data"), 0644)
	}

	// Wait a bit so files are considered "old temp"
	time.Sleep(100 * time.Millisecond)

	cfg := config.Default()
	ctx := context.Background()

	// Scan
	res, err := scanner.ScanWithConfig(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if res.TotalSize == 0 {
		t.Skip("no items found in temp dir, skipping integration test")
	}

	// Select all found items
	var selected []types.Item
	for _, cat := range res.Categories {
		for _, it := range cat.Items {
			it.Selected = true
			selected = append(selected, it)
		}
	}
	if len(selected) == 0 {
		t.Skip("no items to clean, skipping integration test")
	}

	// Clean
	cleanResult, err := Run(ctx, selected, res)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}
	if cleanResult.Files == 0 {
		t.Error("expected at least one file cleaned")
	}

	// Restore
	restored, _, err := quarantine.Restore(cleanResult.RestoreID, false)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if restored == 0 {
		t.Error("expected at least one file restored")
	}

	// Doctor
	checks := doctor.Run()
	for _, check := range checks {
		if check.Status == doctor.StatusFail {
			t.Logf("doctor fail: %s - %s", check.Name, check.Detail)
		}
		if check.Status == doctor.StatusWarn {
			t.Logf("doctor warn: %s - %s", check.Name, check.Detail)
		}
	}
}
