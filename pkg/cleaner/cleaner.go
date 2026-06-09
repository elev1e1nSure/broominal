// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
	"context"
	"log/slog"

	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Run moves selected items to quarantine and persists a report.
func Run(ctx context.Context, items []types.Item, scanResult *types.ScanResult) (*types.CleanResult, error) {
	id, freed, files, skipped, err := quarantine.Move(ctx, items)
	if err != nil {
		return nil, err
	}
	result := &types.CleanResult{
		RestoreID: id,
		Freed:     freed,
		Files:     files,
		Skipped:   skipped,
	}
	if scanResult != nil {
		if _, err := report.Save(scanResult, result); err != nil {
			slog.Warn("cleaner: failed to save report", "error", err)
		}
	}
	return result, nil
}
