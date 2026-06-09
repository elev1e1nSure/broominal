// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
	"log/slog"

	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Run moves selected items to quarantine and optionally persists a report.
// If dryRun is true, nothing is moved and no report is saved.
func Run(items []types.Item, dryRun bool, scanResult *types.ScanResult) (*types.CleanResult, error) {
	id, freed, files, err := quarantine.Move(items, dryRun)
	if err != nil {
		return nil, err
	}
	result := &types.CleanResult{
		RestoreID: id,
		Freed:     freed,
		Files:     files,
	}
	if !dryRun && scanResult != nil {
		if _, err := report.Save(scanResult, result); err != nil {
			slog.Warn("cleaner: failed to save report", "error", err)
		}
	}
	return result, nil
}
