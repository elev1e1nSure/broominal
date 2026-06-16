package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

// BaseDir returns the directory where JSON reports are written. Each save
// creates a fresh timestamped file so the user can diff or archive reports
// across runs.
func BaseDir() string {
	return filepath.Join(config.AppDir(), "reports")
}

// Save writes a JSON report of the scan (and optional clean) to disk and
// returns the resulting file path. The human-readable summary that the `report`
// CLI prints after this is a separate call to PrintSummary — kept out of Save
// so library callers (e.g. the TUI) can persist the file silently.
func Save(result *types.ScanResult, cleaned *types.CleanResult) (string, error) {
	if err := os.MkdirAll(BaseDir(), 0700); err != nil {
		return "", fmt.Errorf("create reports dir: %w", err)
	}
	filename := fmt.Sprintf("report_%s.json", time.Now().Format("20060102_150405.000"))
	path := filepath.Join(BaseDir(), filename)
	if err := SaveTo(path, result, cleaned); err != nil {
		return "", err
	}
	return path, nil
}

// SaveTo writes a JSON report to an explicit file path. The directory must
// already exist. Writes atomically via a temp file + rename.
func SaveTo(path string, result *types.ScanResult, cleaned *types.CleanResult) error {
	data := types.ReportData{
		Timestamp: time.Now(),
		Result:    *result,
		Cleaned:   cleaned,
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	// Write atomically via temp file + rename to avoid corrupted JSON on crash.
	f, err := os.CreateTemp(dir, "report-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	tmp := f.Name()
	var writeErr error
	defer func() {
		if writeErr != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if writeErr = enc.Encode(data); writeErr != nil {
		return fmt.Errorf("encode report: %w", writeErr)
	}
	if writeErr = f.Close(); writeErr != nil {
		return fmt.Errorf("close report: %w", writeErr)
	}
	if writeErr = os.Rename(tmp, path); writeErr != nil {
		return fmt.Errorf("rename report: %w", writeErr)
	}
	return nil
}

// PrintSummary writes a human-readable summary to stdout. Used by the `report`
// CLI after Save so users get immediate feedback without opening the JSON file.
func PrintSummary(result *types.ScanResult, cleaned *types.CleanResult) {
	fmt.Println()
	fmt.Println("═ Broominal Report ════════════════════════")
	fmt.Printf("  Total found:    %s\n", util.FormatSize(result.TotalSize))
	fmt.Printf("  Safe:           %s\n", util.FormatSize(result.SafeSize))
	fmt.Printf("  Review:         %s\n", util.FormatSize(result.ReviewSize))
	fmt.Printf("  Danger:         %s\n", util.FormatSize(result.DangerSize))
	if cleaned != nil {
		fmt.Println()
		fmt.Printf("  Freed:          %s\n", util.FormatSize(cleaned.Freed))
		fmt.Printf("  Files removed:  %d\n", cleaned.Files)
		fmt.Printf("  Restore ID:     %s\n", cleaned.RestoreID)
	}
	fmt.Println("═══════════════════════════════════════")
}
