// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
	"context"
	"log/slog"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Run moves selected items to quarantine (or deletes them directly when quarantine is
// disabled) and persists a report. cfg may be nil, in which case quarantine is used.
func Run(ctx context.Context, items []types.Item, scanResult *types.ScanResult, cfg *config.Config) (*types.CleanResult, error) {
	var result *types.CleanResult

	if cfg != nil && !cfg.QuarantineEnabled {
		freed, files, skipped := deleteDirect(ctx, items)
		result = &types.CleanResult{
			RestoreID: "",
			Freed:     freed,
			Files:     files,
			Skipped:   skipped,
		}
	} else {
		id, freed, files, skipped, err := quarantine.Move(ctx, items)
		if err != nil {
			return nil, err
		}
		result = &types.CleanResult{
			RestoreID: id,
			Freed:     freed,
			Files:     files,
			Skipped:   skipped,
		}
	}

	if scanResult != nil {
		if _, err := report.Save(scanResult, result); err != nil {
			slog.Warn("cleaner: failed to save report", "error", err)
		}
	}
	return result, nil
}

// deleteDirect permanently removes items without quarantine.
func deleteDirect(ctx context.Context, items []types.Item) (freed int64, files int, skipped int) {
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		if err := os.RemoveAll(item.Path); err != nil {
			skipped++
			continue
		}
		freed += item.Size
		files++
	}
	return
}
