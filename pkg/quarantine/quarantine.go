package quarantine

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// BaseDir возвращает базовую директорию карантина
func BaseDir() string {
	return filepath.Join(config.AppDir(), "quarantine")
}

// generateBackupID creates a timestamp-based unique ID.
func generateBackupID() string {
	base := time.Now().Format("2006-01-02-150405")
	qDir := BaseDir()
	id := base
	suffix := 2
	for {
		if _, err := os.Stat(filepath.Join(qDir, id)); os.IsNotExist(err) {
			break
		}
		id = fmt.Sprintf("%s-%d", base, suffix)
		suffix++
	}
	return id
}

// Move перемещает файлы в карантин и возвращает restore ID
func Move(items []types.Item, dryRun bool) (string, int64, int, error) {
	id := generateBackupID()

	if dryRun {
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
		return id, freed, files, nil
	}

	qDir := filepath.Join(BaseDir(), id)
	if err := os.MkdirAll(qDir, 0700); err != nil {
		return "", 0, 0, fmt.Errorf("create quarantine dir: %w", err)
	}

	manifest := types.Manifest{
		ID:        id,
		CreatedAt: time.Now(),
		Label:     "Cleanup " + id,
		Items:     make([]types.ManifestItem, 0, len(items)),
	}

	var freed int64
	var files int
	for _, it := range items {
		if !it.Selected {
			continue
		}
		if _, err := os.Stat(it.Path); os.IsNotExist(err) {
			continue
		}

		qPath := uniquePath(qDir, filepath.Base(it.Path))

		moved := true
		if err := os.Rename(it.Path, qPath); err != nil {
			// fallback to copy+delete if cross-device or file is locked
			if err := copyAndDelete(it.Path, qPath); err != nil {
				moved = false
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
		}
	}
	manifest.TotalSize = freed
	manifest.Files = files

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
		if !isAllowedRestorePath(it.Original) {
			remaining = append(remaining, it)
			continue
		}
		// ensure original dir exists
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
		if err := enc.Encode(manifest); err != nil {
			slog.Warn("restore: failed to update manifest", "path", manifestPath, "error", err)
		}
	} else {
		if err := os.RemoveAll(qDir); err != nil {
			slog.Warn("restore: failed to remove quarantine dir", "path", qDir, "error", err)
		}
	}

	return restored, skipped, nil
}

// CheckRestoreConflicts возвращает пути оригинальных файлов, которые уже существуют
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

// GetManifest reads a manifest by restore ID.
func GetManifest(id string) (*types.Manifest, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	path := filepath.Join(BaseDir(), id, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m types.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// List возвращает список доступных restore ID, отсортированных от новых к старым.
func List() ([]string, error) {
	qDir := BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type entryInfo struct {
		id   string
		time time.Time
	}
	var infos []entryInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue
		}
		m, _ := GetManifest(id)
		if m != nil {
			infos = append(infos, entryInfo{id: id, time: m.CreatedAt})
		} else {
			// fallback to dir mod time
			info, _ := os.Stat(filepath.Join(qDir, id))
			if info != nil {
				infos = append(infos, entryInfo{id: id, time: info.ModTime()})
			}
		}
	}
	// sort newest first
	for i := 0; i < len(infos)-1; i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[j].time.After(infos[i].time) {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.id)
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
		if err := validateID(id); err != nil {
			continue
		}
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
			if err := json.NewDecoder(mf).Decode(&manifest); err != nil {
				slog.Warn("cleanup: failed to decode manifest", "path", filepath.Join(dirPath, "manifest.json"), "error", err)
			} else {
				createdAt = manifest.CreatedAt
			}
			if err := mf.Close(); err != nil {
				slog.Warn("cleanup: failed to close manifest", "path", filepath.Join(dirPath, "manifest.json"), "error", err)
			}
		}

		if createdAt.IsZero() || createdAt.Before(cutoff) {
			var size int64
			if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				size += info.Size()
				return nil
			}); err != nil {
				slog.Warn("quarantine cleanup: failed to walk dir", "path", dirPath, "error", err)
			}
			if !dryRun {
				if err := os.RemoveAll(dirPath); err != nil {
					slog.Warn("cleanup: failed to remove quarantine dir", "path", dirPath, "error", err)
				}
			}
			deleted++
			freed += size
		}
	}

	return deleted, freed, nil
}

// CleanupAll removes all quarantine entries regardless of age.
// Returns number of deleted quarantines and total freed bytes.
func CleanupAll(dryRun bool) (int, int64, error) {
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

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if err := validateID(id); err != nil {
			continue
		}
		dirPath := filepath.Join(qDir, id)

		var size int64
		if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			size += info.Size()
			return nil
		}); err != nil {
			slog.Warn("quarantine cleanup all: failed to walk dir", "path", dirPath, "error", err)
		}
		if !dryRun {
			if err := os.RemoveAll(dirPath); err != nil {
				slog.Warn("cleanup: failed to remove quarantine dir", "path", dirPath, "error", err)
			}
		}
		deleted++
		freed += size
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
	return ids[0], nil
}

func validateID(id string) error {
	clean := filepath.Clean(id)
	if filepath.IsAbs(clean) {
		return fmt.Errorf("invalid restore id")
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid restore id")
	}
	return nil
}

func isAllowedRestorePath(path string) bool {
	clean := filepath.Clean(path)
	if strings.Contains(clean, "..") {
		return false
	}
	if !filepath.IsAbs(clean) {
		return false
	}
	allowedRoots := []string{
		os.Getenv("TEMP"),
		os.Getenv("TMP"),
		os.Getenv("USERPROFILE"),
		os.Getenv("LOCALAPPDATA"),
		os.Getenv("APPDATA"),
		os.Getenv("ProgramData"),
		os.Getenv("SystemRoot"),
		os.Getenv("WINDIR"),
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("SYSTEMDRIVE"),
	}
	lowerClean := strings.ToLower(clean)
	for _, root := range allowedRoots {
		if root == "" {
			continue
		}
		rootClean := strings.ToLower(filepath.Clean(root))
		if strings.HasPrefix(lowerClean, rootClean) {
			return true
		}
	}
	// Hardcoded NVIDIA path used by scanner
	if strings.HasPrefix(lowerClean, `c:\nvidia`) {
		return true
	}
	return false
}

func uniquePath(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
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

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return err
	}

	buf := make([]byte, 64*1024)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				out.Close()
				in.Close()
				return werr
			}
		}
		if err != nil {
			break
		}
	}
	if cerr := out.Close(); cerr != nil {
		in.Close()
		return cerr
	}
	if cerr := in.Close(); cerr != nil {
		return cerr
	}
	return os.Remove(src)
}
