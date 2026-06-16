package quarantine

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

func Cleanup(maxAgeDays int) (int, int64, error) {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	return cleanupQuarantines(func(createdAt time.Time) bool {
		return createdAt.IsZero() || createdAt.Before(cutoff)
	})
}

func CleanupAll() (int, int64, error) {
	return cleanupQuarantines(func(createdAt time.Time) bool {
		return true
	})
}

func cleanupQuarantines(shouldDelete func(time.Time) bool) (int, int64, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	var deleted int
	var freed int64
	var errs []error

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue
		}
		dirPath := filepath.Join(qDir, id)
		manifestPath := filepath.Join(dirPath, "manifest.json")

		var createdAt time.Time
		mf, err := os.Open(manifestPath)
		if err != nil {
			info, _ := os.Stat(dirPath)
			if info != nil {
				createdAt = info.ModTime()
			}
		} else {
			var manifest types.Manifest
			_ = json.NewDecoder(mf).Decode(&manifest)
			createdAt = manifest.CreatedAt
			_ = mf.Close()
		}

		if shouldDelete(createdAt) {
			var size int64
			_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				info, err := d.Info()
				if err != nil {
					return nil
				}
				size += info.Size()
				return nil
			})
			if err := removeAllRetry(dirPath); err != nil {
				// If the manifest is already gone the entry won't appear in the list anymore —
				// treat as a functional success and only warn about leftover locked files.
				if _, statErr := os.Stat(manifestPath); os.IsNotExist(statErr) {
					slog.Warn("quarantine cleanup: manifest removed but dir has locked files", "path", dirPath, "error", err)
					deleted++
					freed += size
				} else {
					errs = append(errs, err)
				}
			} else {
				deleted++
				freed += size
			}
		}
	}

	if len(errs) > 0 {
		slog.Warn("quarantine cleanup: some directories could not be removed", "errors", len(errs))
		return deleted, freed, errs[0]
	}
	return deleted, freed, nil
}

// PurgeDamaged removes all quarantine directories whose manifest is missing or
// unparseable. These entries cannot be restored and serve no purpose, but they
// cause doctor to report WARN on every run until cleaned up.
func PurgeDamaged() (int, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var removed int
	var firstErr error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue
		}
		dirPath := filepath.Join(qDir, id)
		mf := filepath.Join(dirPath, "manifest.json")
		data, err := os.ReadFile(mf)
		if err == nil {
			var m types.Manifest
			if json.Unmarshal(data, &m) == nil {
				// manifest is intact — skip
				continue
			}
		}
		if err := forceRemoveAll(dirPath); err != nil && firstErr == nil {
			firstErr = err
		} else {
			removed++
		}
	}
	return removed, firstErr
}

func Delete(id string) (int64, error) {
	if err := validateID(id); err != nil {
		return 0, err
	}
	dirPath := filepath.Join(BaseDir(), id)
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("backup not found: %s", id)
		}
		return 0, err
	}
	var size int64
	if info.IsDir() {
		filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			fi, err := d.Info()
			if err != nil {
				return nil
			}
			size += fi.Size()
			return nil
		})
	}
	if err := removeAllRetry(dirPath); err != nil {
		// If manifest is gone the entry won't reappear — locked leftover files are best-effort.
		manifestPath := filepath.Join(dirPath, "manifest.json")
		if _, statErr := os.Stat(manifestPath); os.IsNotExist(statErr) {
			slog.Warn("quarantine: manifest removed but dir has locked files", "path", dirPath, "error", err)
			return size, nil
		}
		return 0, err
	}
	return size, nil
}
