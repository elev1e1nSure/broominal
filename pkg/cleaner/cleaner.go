// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Run is the top-level clean entry point: it picks quarantine or direct
// delete based on cfg, and writes a JSON report when scanResult is provided.
// cfg may be nil to imply "use quarantine" — the CLI and the TUI both rely
// on this default for first-run, config-not-yet-loaded cases.
// progress is an optional callback for per-file progress during quarantine
// moves; nil disables progress reporting.
func Run(ctx context.Context, items []types.Item, scanResult *types.ScanResult, cfg *config.Config, progress types.ProgressFn) (*types.CleanResult, error) {
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
		var totalBytes int64
		for _, it := range items {
			if it.Selected {
				totalBytes += it.Size
			}
		}
		freed, files, skipped := deleteDirect(ctx, items, totalBytes, progress)
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

		id, freed, files, skipped, err := quarantine.Move(ctx, items, progress)
		if err != nil {
			return nil, err
		}
		cancelled := false
		if errors.Is(ctx.Err(), context.Canceled) {
			cancelled = true
		}
		result = &types.CleanResult{
			RestoreID: id,
			Freed:     freed,
			Files:     files,
			Skipped:   skipped,
			Cancelled: cancelled,
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
// opted out of quarantine in config. It sends progress updates via the
// optional callback (throttled to ~250 ms), matching quarantine.Move semantics.
func deleteDirect(ctx context.Context, items []types.Item, totalBytes int64, progress types.ProgressFn) (freed int64, files int, skipped int) {
	total := len(items)
	startTime := time.Now()
	lastProg := time.Now()

	for i, item := range items {
		if ctx.Err() != nil {
			break
		}
		if err := os.RemoveAll(item.Path); err != nil {
			slog.Warn("cleaner: skipped locked or inaccessible path", "path", item.Path, "error", err)
			skipped++
		} else {
			freed += item.Size
			files++
		}

		if progress != nil {
			now := time.Now()
			if now.Sub(lastProg) > 250*time.Millisecond || i == total-1 {
				lastProg = now
				progress(types.Progress{
					Stage:      "cleaning",
					Processed:  i + 1,
					Total:      total,
					Bytes:      freed,
					TotalBytes: totalBytes,
					StartedAt:  startTime,
				})
			}
		}
	}
	// Final progress tick to ensure 100% is shown.
	if progress != nil {
		progress(types.Progress{
			Stage:      "cleaning",
			Processed:  total,
			Total:      total,
			Bytes:      freed,
			TotalBytes: totalBytes,
			StartedAt:  startTime,
		})
	}
	return
}
