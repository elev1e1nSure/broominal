package quarantine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// lockPath returns the path to the quarantine lock file.
func lockPath() string {
	return filepath.Join(BaseDir(), ".lock")
}

// acquireLock creates an exclusive lock file to prevent concurrent Move/Restore
// from different processes (e.g. TUI + scheduled schtasks cleanup).
// Retries for up to 5 seconds, then returns an error.
func acquireLock() error {
	if err := os.MkdirAll(BaseDir(), 0700); err != nil {
		return fmt.Errorf("quarantine lock: create base dir: %w", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		f, err := os.OpenFile(lockPath(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			_ = f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return fmt.Errorf("quarantine lock: %w", err)
		}
		// Lock held by another process — check staleness (> 60s old = stale).
		if info, statErr := os.Stat(lockPath()); statErr == nil {
			if time.Since(info.ModTime()) > 60*time.Second {
				_ = os.Remove(lockPath())
				continue
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("quarantine: timed out waiting for lock (another process is running)")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// releaseLock removes the lock file.
func releaseLock() {
	_ = os.Remove(lockPath())
}

// BaseDir returns the quarantine base directory under the app data folder.
func BaseDir() string {
	return filepath.Join(config.AppDir(), "quarantine")
}

func generateBackupID() (string, error) {
	base := time.Now().Format("2006-01-02-150405")
	qDir := BaseDir()
	if err := os.MkdirAll(qDir, 0700); err != nil {
		return "", fmt.Errorf("create quarantine base dir: %w", err)
	}
	id := base
	suffix := 2
	for {
		dirPath := filepath.Join(qDir, id)
		err := os.Mkdir(dirPath, 0700)
		if err == nil {
			return id, nil
		}
		if os.IsExist(err) {
			id = fmt.Sprintf("%s-%d", base, suffix)
			suffix++
			continue
		}
		return "", fmt.Errorf("create quarantine dir: %w", err)
	}
}

// Move transfers the given items to a fresh quarantine batch and returns its
// restore ID. Items that are missing, locked, or symlinks are counted as
// skipped and excluded from the manifest so a later restore cannot resurrect
// dangerous targets.
func Move(ctx context.Context, items []types.Item) (string, int64, int, int, error) {
	if err := acquireLock(); err != nil {
		return "", 0, 0, 0, err
	}
	defer releaseLock()
	id, err := generateBackupID()
	if err != nil {
		return "", 0, 0, 0, err
	}

	qDir := filepath.Join(BaseDir(), id)

	manifest := types.Manifest{
		ID:        id,
		CreatedAt: time.Now(),
		Label:     "Cleanup " + id,
		Items:     make([]types.ManifestItem, 0, len(items)),
	}

	journal, err := NewJournal(qDir)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("create journal: %w", err)
	}

	var freed int64
	var files int
	var skipped int
	for _, it := range items {
		if ctx.Err() != nil {
			journal.Close()
			return id, freed, files, skipped, ctx.Err()
		}
		if !it.Selected {
			continue
		}
		info, err := os.Lstat(it.Path)
		if err != nil {
			skipped++
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			skipped++
			slog.Warn("quarantine: skipping symlink", "path", it.Path)
			continue
		}

		qPath := uniquePath(qDir, filepath.Base(it.Path)+".brm")

		if err := journal.Begin(it.Path, qPath, it.Size, it.Category); err != nil {
			skipped++
			slog.Warn("quarantine: failed to write journal begin", "error", err)
			continue
		}

		moved := true
		if err := os.Rename(it.Path, qPath); err != nil {
			if err := copyAndDelete(it.Path, qPath); err != nil {
				moved = false
				slog.Warn("quarantine: failed to move file", "path", it.Path, "error", err)
			}
		}

		if moved {
			if err := journal.Commit(it.Path, qPath); err != nil {
				slog.Warn("quarantine: failed to write journal commit", "error", err)
				journal.Close()
				return id, freed, files, skipped, err
			}

			manifest.Items = append(manifest.Items, types.ManifestItem{
				Original:    it.Path,
				Quarantined: qPath,
				Size:        it.Size,
			})
			freed += it.Size
			files++
		} else {
			skipped++
		}
	}
	manifest.TotalSize = freed
	manifest.Files = files

	catSet := make(map[string]struct{})
	for _, it := range items {
		if it.Selected && it.Category != "" {
			catSet[it.Category] = struct{}{}
		}
	}
	manifest.Categories = make([]string, 0, len(catSet))
	for c := range catSet {
		manifest.Categories = append(manifest.Categories, c)
	}
	sort.Strings(manifest.Categories)

	manifestPath := filepath.Join(qDir, "manifest.json")
	if err := writeManifest(manifestPath, &manifest); err != nil {
		journal.Close()
		return "", 0, 0, 0, fmt.Errorf("write manifest: %w", err)
	}

	journal.Close()
	_ = os.Remove(filepath.Join(qDir, "journal.jsonl"))

	return id, freed, files, skipped, nil
}

// Restore returns quarantined files to their original paths using the manifest
// for the given ID. The allow-list in isAllowedRestorePath is the security
// boundary — manifest entries pointing outside user-writable roots stay in
// quarantine to prevent path-traversal abuse via a tampered manifest.
func Restore(id string, forceOverwrite bool) (int, int, error) {
	if err := acquireLock(); err != nil {
		return 0, 0, err
	}
	defer releaseLock()

	if err := validateID(id); err != nil {
		return 0, 0, err
	}
	qDir := filepath.Join(BaseDir(), id)
	manifestPath := filepath.Join(qDir, "manifest.json")

	mf, err := os.Open(manifestPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open manifest: %w", err)
	}
	var manifest types.Manifest
	if err := json.NewDecoder(mf).Decode(&manifest); err != nil {
		_ = mf.Close()
		return 0, 0, fmt.Errorf("decode manifest: %w", err)
	}
	_ = mf.Close()

	var restored int
	var skipped int
	var remaining []types.ManifestItem

	for _, it := range manifest.Items {
		if _, err := os.Stat(it.Quarantined); os.IsNotExist(err) {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(filepath.Clean(it.Quarantined)), strings.ToLower(filepath.Clean(qDir))) {
			slog.Warn("restore: quarantined path outside expected dir, skipping", "path", it.Quarantined)
			remaining = append(remaining, it)
			continue
		}
		if !isAllowedRestorePath(it.Original) {
			remaining = append(remaining, it)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(it.Original), 0700); err != nil {
			remaining = append(remaining, it)
			continue
		}
		_, statErr := os.Stat(it.Original)
		exists := statErr == nil
		if exists && !forceOverwrite {
			skipped++
			remaining = append(remaining, it)
			continue
		}
		if exists && forceOverwrite {
			if err := os.Remove(it.Original); err != nil {
				slog.Warn("restore: failed to remove existing file", "path", it.Original, "error", err)
			}
		}
		if err := os.Rename(it.Quarantined, it.Original); err != nil {
			if err := copyAndDelete(it.Quarantined, it.Original); err != nil {
				remaining = append(remaining, it)
				continue
			}
		}
		restored++
	}

	if len(remaining) > 0 {
		manifest.Items = remaining
		if err := writeManifest(manifestPath, &manifest); err != nil {
			return restored, skipped, fmt.Errorf("update manifest: %w", err)
		}
	} else {
		if err := removeAllRetry(qDir); err != nil {
			slog.Warn("restore: failed to remove quarantine dir", "path", qDir, "error", err)
		}
	}

	return restored, skipped, nil
}

// CheckRestoreConflicts returns original paths that already exist on disk and
// would be overwritten by a restore. The TUI uses this to offer the user a
// per-conflict overwrite/skip choice before any file is touched.
func CheckRestoreConflicts(id string) ([]string, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	qDir := filepath.Join(BaseDir(), id)
	manifestPath := filepath.Join(qDir, "manifest.json")

	mf, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("open manifest: %w", err)
	}
	defer mf.Close()

	var manifest types.Manifest
	if err := json.NewDecoder(mf).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}

	var conflicts []string
	for _, it := range manifest.Items {
		if _, err := os.Stat(it.Original); err == nil {
			conflicts = append(conflicts, it.Original)
		}
	}
	return conflicts, nil
}
