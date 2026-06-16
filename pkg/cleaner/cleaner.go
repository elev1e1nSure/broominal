// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Run is the top-level clean entry point: it picks quarantine or direct
// delete based on cfg, and writes a JSON report when scanResult is provided.
// cfg may be nil to imply "use quarantine" — the CLI and the TUI both rely
// on this default for first-run, config-not-yet-loaded cases.
func Run(ctx context.Context, items []types.Item, scanResult *types.ScanResult, cfg *config.Config) (*types.CleanResult, error) {
	sort.Slice(items, func(i, j int) bool {
		di := strings.Count(items[i].Path, string(os.PathSeparator))
		dj := strings.Count(items[j].Path, string(os.PathSeparator))
		if di != dj {
			// Process deeper paths first so a parent directory isn't
			// renamed/removed before the files it still contains.
			return di > dj
		}
		return items[i].Path > items[j].Path
	})

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
		if cfg != nil && cfg.QuarantineEnabled {
			var totalSize int64
			for _, it := range items {
				if it.Selected {
					totalSize += it.Size
				}
			}
			if err := quarantine.CheckHealth(ctx, totalSize); err != nil {
				return nil, fmt.Errorf("quarantine health check failed: %w", err)
			}
		}

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

// deleteDirect is the irreversible cleanup path taken when the user has
// opted out of quarantine in config. It exists separately from Move so the
// caller never has to think about which mode is in effect; Run picks the path.
func deleteDirect(ctx context.Context, items []types.Item) (freed int64, files int, skipped int) {
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		if err := os.RemoveAll(item.Path); err != nil {
			slog.Warn("cleaner: skipped locked or inaccessible path", "path", item.Path, "error", err)
			skipped++
			continue
		}
		freed += item.Size
		files++
	}
	return
}
