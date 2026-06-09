// Package cleaner orchestrates the quarantine-and-report pipeline.
package cleaner

import (
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
		_, _ = report.Save(scanResult, result)
	}
	return result, nil
}
