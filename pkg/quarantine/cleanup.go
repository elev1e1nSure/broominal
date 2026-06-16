package quarantine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
			if err := forceRemoveAll(dirPath); err != nil {
				if errors.Is(err, ErrScheduledForReboot) {
					deleted++
					freed += size
				} else if _, statErr := os.Stat(manifestPath); os.IsNotExist(statErr) {
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

// RepairDamaged attempts to recover items from corrupted manifest.json files.
// It uses a tolerant regex parser to extract items that are still present on disk.
// Returns the number of directories repaired, the number of completely dead ones, and any error.
func RepairDamaged() (int, int, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	var repaired int
	var dead int
	var firstErr error

	re := regexp.MustCompile(`"original"\s*:\s*"([^"]+)",\s*"quarantined"\s*:\s*"([^"]+)",\s*"size"\s*:\s*(\d+)`)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue
		}
		dirPath := filepath.Join(qDir, id)
		mfPath := filepath.Join(dirPath, "manifest.json")
		deadPath := filepath.Join(dirPath, "manifest.dead")

		if _, err := os.Stat(deadPath); err == nil {
			dead++
			continue
		}

		data, err := os.ReadFile(mfPath)
		if err == nil {
			var m types.Manifest
			if json.Unmarshal(data, &m) == nil {
				// manifest is intact — skip
				continue
			}
		} else if !os.IsNotExist(err) {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		// It's damaged or missing. Let's try to repair.
		var recovered []types.ManifestItem
		var totalSize int64

		if data != nil {
			matches := re.FindAllStringSubmatch(string(data), -1)
			for _, m := range matches {
				orig := unescapeJSONString(m[1])
				quar := unescapeJSONString(m[2])
				size, _ := strconv.ParseInt(m[3], 10, 64)

				// Verify it exists in quarantine
				if _, statErr := os.Stat(quar); statErr == nil {
					recovered = append(recovered, types.ManifestItem{
						Original:    orig,
						Quarantined: quar,
						Size:        size,
					})
					totalSize += size
				}
			}
		}

		if len(recovered) > 0 {
			info, _ := os.Stat(dirPath)
			createdAt := time.Now()
			if info != nil {
				createdAt = info.ModTime()
			}
			m := types.Manifest{
				ID:        id,
				CreatedAt: createdAt,
				Label:     "Recovered Cleanup " + id,
				TotalSize: totalSize,
				Files:     len(recovered),
				Items:     recovered,
			}
			if err := writeManifest(mfPath, &m); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			} else {
				repaired++
			}
		} else {
			// Unrecoverable
			if err == nil { // mfPath exists
				os.Rename(mfPath, deadPath)
			} else { // mfPath didn't exist, just create a dummy .dead
				os.WriteFile(deadPath, nil, 0600)
			}
			dead++
		}
	}
	return repaired, dead, firstErr
}

func unescapeJSONString(s string) string {
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\/`, `/`)
	s = strings.ReplaceAll(s, `\b`, "\b")
	s = strings.ReplaceAll(s, `\f`, "\f")
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

// PurgeDead removes all quarantine directories that were marked as unrecoverable.
func PurgeDead() (int, error) {
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
		deadPath := filepath.Join(dirPath, "manifest.dead")

		if _, err := os.Stat(deadPath); err != nil {
			// Not marked as dead
			continue
		}

		if err := forceRemoveAll(dirPath); err != nil {
			if errors.Is(err, ErrScheduledForReboot) {
				removed++
				if firstErr == nil {
					firstErr = err
				}
			} else {
				if firstErr == nil {
					firstErr = err
				}
			}
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
	if err := forceRemoveAll(dirPath); err != nil {
		if errors.Is(err, ErrScheduledForReboot) {
			return size, ErrScheduledForReboot
		}
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
