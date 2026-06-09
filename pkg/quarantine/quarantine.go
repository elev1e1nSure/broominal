package quarantine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// BaseDir возвращает базовую директорию карантина
func BaseDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("USERPROFILE")
	}
	return filepath.Join(localAppData, "broominal", "quarantine")
}

// Move перемещает файлы в карантин и возвращает restore ID
func Move(items []types.Item, dryRun bool) (string, int64, int, error) {
	id := uuid.New().String()

	var freed int64
	var files int

	for _, it := range items {
		if !it.Selected {
			continue
		}
		if _, err := os.Stat(it.Path); os.IsNotExist(err) {
			continue
		}
		freed += it.Size
		files++
	}

	if dryRun {
		return "", freed, files, nil
	}

	qDir := filepath.Join(BaseDir(), id)
	if err := os.MkdirAll(qDir, 0755); err != nil {
		return "", 0, 0, fmt.Errorf("create quarantine dir: %w", err)
	}

	manifest := types.Manifest{
		ID:        id,
		CreatedAt: time.Now(),
		Items:     make([]types.ManifestItem, 0, len(items)),
	}

	for _, it := range items {
		if !it.Selected {
			continue
		}
		if _, err := os.Stat(it.Path); os.IsNotExist(err) {
			continue
		}

		qPath := filepath.Join(qDir, filepath.Base(it.Path))
		// handle duplicate names
		qPath = uniquePath(qDir, filepath.Base(it.Path))

		if err := os.Rename(it.Path, qPath); err != nil {
			// fallback to copy+delete if cross-device
			if err := copyAndDelete(it.Path, qPath); err != nil {
				continue
			}
		}

		manifest.Items = append(manifest.Items, types.ManifestItem{
			Original:    it.Path,
			Quarantined: qPath,
			Size:        it.Size,
		})
	}

	manifestPath := filepath.Join(qDir, "manifest.json")
	mf, err := os.Create(manifestPath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("create manifest: %w", err)
	}
	defer mf.Close()

	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		return "", 0, 0, fmt.Errorf("encode manifest: %w", err)
	}

	return id, freed, files, nil
}

// Restore восстанавливает файлы из карантина по ID.
// Возвращает количество восстановленных, пропущенных (конфликт) и ошибку.
func Restore(id string, forceOverwrite bool) (int, int, error) {
	qDir := filepath.Join(BaseDir(), id)
	manifestPath := filepath.Join(qDir, "manifest.json")

	mf, err := os.Open(manifestPath)
	if err != nil {
		return 0, 0, fmt.Errorf("open manifest: %w", err)
	}
	defer mf.Close()

	var manifest types.Manifest
	if err := json.NewDecoder(mf).Decode(&manifest); err != nil {
		return 0, 0, fmt.Errorf("decode manifest: %w", err)
	}

	var restored int
	var skipped int
	var remaining []types.ManifestItem

	for _, it := range manifest.Items {
		if _, err := os.Stat(it.Quarantined); os.IsNotExist(err) {
			continue
		}
		// ensure original dir exists
		if err := os.MkdirAll(filepath.Dir(it.Original), 0755); err != nil {
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
			_ = os.Remove(it.Original)
		}
		if err := os.Rename(it.Quarantined, it.Original); err != nil {
			// fallback
			if err := copyAndDelete(it.Quarantined, it.Original); err != nil {
				remaining = append(remaining, it)
				continue
			}
		}
		restored++
	}

	if len(remaining) > 0 {
		manifest.Items = remaining
		mf2, err := os.Create(manifestPath)
		if err != nil {
			return restored, skipped, fmt.Errorf("update manifest: %w", err)
		}
		defer mf2.Close()
		enc := json.NewEncoder(mf2)
		enc.SetIndent("", "  ")
		_ = enc.Encode(manifest)
	} else {
		_ = os.RemoveAll(qDir)
	}

	return restored, skipped, nil
}

// CheckRestoreConflicts возвращает пути оригинальных файлов, которые уже существуют
func CheckRestoreConflicts(id string) ([]string, error) {
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

// List возвращает список доступных restore ID
func List() ([]string, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, nil
}

// Cleanup removes quarantine entries older than maxAgeDays.
// Returns number of deleted quarantines and total freed bytes.
func Cleanup(maxAgeDays int, dryRun bool) (int, int64, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	var deleted int
	var freed int64

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		dirPath := filepath.Join(qDir, id)
		manifestPath := filepath.Join(dirPath, "manifest.json")

		var createdAt time.Time
		mf, err := os.Open(manifestPath)
		if err != nil {
			// broken quarantine: use dir mod time
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

		if createdAt.IsZero() || createdAt.Before(cutoff) {
			var size int64
			_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				size += info.Size()
				return nil
			})
			if !dryRun {
				_ = os.RemoveAll(dirPath)
			}
			deleted++
			freed += size
		}
	}

	return deleted, freed, nil
}

// GetLast возвращает последний restore ID
func GetLast() (string, error) {
	ids, err := List()
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("no cleanups to restore")
	}
	// Return most recent (by dir mod time)
	var lastID string
	var lastTime time.Time
	for _, id := range ids {
		info, err := os.Stat(filepath.Join(BaseDir(), id))
		if err != nil {
			continue
		}
		if info.ModTime().After(lastTime) {
			lastTime = info.ModTime()
			lastID = id
		}
	}
	return lastID, nil
}

func uniquePath(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s_%d%s", base, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

func copyAndDelete(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 64*1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			break
		}
	}
	return os.Remove(src)
}
