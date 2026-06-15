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
				errs = append(errs, err)
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
		return 0, err
	}
	return size, nil
}
