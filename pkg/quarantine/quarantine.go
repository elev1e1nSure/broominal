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

// Move перемещает файлы в карантин и возвращает restore ID.
func Move(ctx context.Context, items []types.Item) (string, int64, int, int, error) {
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

	var freed int64
	var files int
	var skipped int
	for _, it := range items {
		if ctx.Err() != nil {
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

		qPath := uniquePath(qDir, filepath.Base(it.Path))

		moved := true
		if err := os.Rename(it.Path, qPath); err != nil {
			if err := copyAndDelete(it.Path, qPath); err != nil {
				moved = false
				slog.Warn("quarantine: failed to move file", "path", it.Path, "error", err)
			}
		}

		if moved {
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
		return "", 0, 0, 0, fmt.Errorf("write manifest: %w", err)
	}

	return id, freed, files, skipped, nil
}

// Restore восстанавливает файлы из карантина по ID.
func Restore(id string, forceOverwrite bool) (int, int, error) {
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
		mf.Close()
		return 0, 0, fmt.Errorf("decode manifest: %w", err)
	}
	mf.Close()

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

// CheckRestoreConflicts возвращает пути оригинальных файлов, которые уже существуют.
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
