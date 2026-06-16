package quarantine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

// CheckHealth verifies that the quarantine system is healthy and ready to accept files.
// It checks base dir existence, available disk space, and automatically repairs any
// corrupted or interrupted batches from previous crashes.
func CheckHealth(ctx context.Context, requiredBytes int64) error {
	baseDir := BaseDir()

	// 1. Check if base directory exists, create if not
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return fmt.Errorf("quarantine base dir inaccessible: %w", err)
	}

	// 2. Repair interrupted batches (WAL recovery)
	if err := RepairState(); err != nil {
		return fmt.Errorf("repair quarantine state: %w", err)
	}

	// 3. Check disk space
	freeSpace, err := getDiskFreeSpace(baseDir)
	if err != nil {
		slog.Warn("quarantine: could not determine disk free space", "error", err)
		// Non-fatal if we can't query the disk space
	} else {
		// Require at least requiredBytes + a small safety margin (e.g., 50MB)
		margin := int64(50 * 1024 * 1024)
		if freeSpace < requiredBytes+margin {
			return fmt.Errorf("insufficient disk space: need %d bytes, but only %d bytes available", requiredBytes+margin, freeSpace)
		}
	}

	return nil
}

// RepairState scans the base directory for orphaned batches (batches with a journal.jsonl
// but no manifest.json) and finalizes them, synthesizing a manifest so they can be restored.
func RepairState() error {
	baseDir := BaseDir()
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue // Not a valid batch directory
		}

		batchDir := filepath.Join(baseDir, id)
		journalPath := filepath.Join(batchDir, "journal.jsonl")
		manifestPath := filepath.Join(batchDir, "manifest.json")

		// If manifest exists, batch was successfully completed
		if _, err := os.Stat(manifestPath); err == nil {
			// Clean up leftover journal if somehow it wasn't deleted
			_ = os.Remove(journalPath)
			continue
		}

		// If manifest is missing but journal exists, it's an interrupted batch
		if _, err := os.Stat(journalPath); err == nil {
			slog.Info("quarantine: repairing interrupted batch", "id", id)
			if err := repairBatch(id, batchDir, journalPath, manifestPath); err != nil {
				slog.Warn("quarantine: failed to repair batch", "id", id, "error", err)
				// We don't return the error, we just warn and continue repairing others
			}
		} else {
			// Neither manifest nor journal exists. Probably an empty directory or a crash
			// right after directory creation. Remove it to keep things clean.
			isEmpty, _ := isDirEmpty(batchDir)
			if isEmpty {
				_ = os.Remove(batchDir)
			}
		}
	}
	return nil
}

func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func repairBatch(id, batchDir, journalPath, manifestPath string) error {
	jf, err := os.Open(journalPath)
	if err != nil {
		return err
	}
	defer jf.Close()

	scanner := bufio.NewScanner(jf)
	var entries []JournalEntry
	for scanner.Scan() {
		var entry JournalEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip corrupted lines
		}
		entries = append(entries, entry)
	}

	// Reconstruct state
	type fileState struct {
		Original    string
		Quarantined string
		Size        int64
		Category    string
		Done        bool
	}
	stateMap := make(map[string]*fileState)

	for _, e := range entries {
		if e.Op == "begin" {
			stateMap[e.Original] = &fileState{
				Original:    e.Original,
				Quarantined: e.Quarantined,
				Size:        e.Size,
				Category:    e.Category,
				Done:        false,
			}
		} else if e.Op == "done" {
			if s, ok := stateMap[e.Original]; ok {
				s.Done = true
			}
		}
	}

	var manifest types.Manifest
	manifest.ID = id
	manifest.CreatedAt = time.Now()
	manifest.Label = "Recovered Cleanup " + id

	var freed int64
	var files int
	catSet := make(map[string]struct{})

	for _, s := range stateMap {
		if s.Done {
			manifest.Items = append(manifest.Items, types.ManifestItem{
				Original:    s.Original,
				Quarantined: s.Quarantined,
				Size:        s.Size,
			})
			freed += s.Size
			files++
			if s.Category != "" {
				catSet[s.Category] = struct{}{}
			}
			continue
		}

		qStat, qErr := os.Stat(s.Quarantined)
		oStat, oErr := os.Stat(s.Original)

		qExists := qErr == nil
		oExists := oErr == nil

		if qExists && oExists {
			if qStat.Size() == s.Size && oStat.Size() == s.Size {
				if err := removeRetry(s.Original); err == nil {
					manifest.Items = append(manifest.Items, types.ManifestItem{
						Original:    s.Original,
						Quarantined: s.Quarantined,
						Size:        s.Size,
					})
					freed += s.Size
					files++
					if s.Category != "" {
						catSet[s.Category] = struct{}{}
					}
				} else {
					_ = removeRetry(s.Quarantined)
				}
			} else {
				_ = removeRetry(s.Quarantined)
			}
		} else if qExists && !oExists {
			manifest.Items = append(manifest.Items, types.ManifestItem{
				Original:    s.Original,
				Quarantined: s.Quarantined,
				Size:        s.Size,
			})
			freed += s.Size
			files++
			if s.Category != "" {
				catSet[s.Category] = struct{}{}
			}
		}
	}

	manifest.TotalSize = freed
	manifest.Files = files
	for c := range catSet {
		manifest.Categories = append(manifest.Categories, c)
	}
	sort.Strings(manifest.Categories)

	if len(manifest.Items) == 0 {
		jf.Close()
		_ = os.Remove(journalPath)
		_ = os.RemoveAll(batchDir)
		return nil
	}

	if err := writeManifest(manifestPath, &manifest); err != nil {
		return fmt.Errorf("write recovered manifest: %w", err)
	}

	jf.Close()
	_ = os.Remove(journalPath)

	return nil
}
