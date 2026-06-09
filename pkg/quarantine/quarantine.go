package quarantine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// BaseDir — возвращает базовую директорию карантина
func BaseDir() string {
	return filepath.Join(config.AppDir(), "quarantine")
}

// generateBackupID creates a timestamp-based unique ID atomically by
// attempting os.Mkdir; if the directory already exists it appends a suffix.
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

// Move — перемещает файлы в карантин и возвращает restore ID
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
		// Check for symlinks and existence in one call
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
			// fallback to copy+delete if cross-device or file is locked
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

	manifestPath := filepath.Join(qDir, "manifest.json")
	if err := writeManifest(manifestPath, &manifest); err != nil {
		return "", 0, 0, 0, fmt.Errorf("write manifest: %w", err)
	}

	return id, freed, files, skipped, nil
}

// Restore — восстанавливает файлы из карантина по ID.
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
		if !strings.HasPrefix(strings.ToLower(filepath.Clean(it.Quarantined)), strings.ToLower(filepath.Clean(qDir))) {
			slog.Warn("restore: quarantined path outside expected dir, skipping", "path", it.Quarantined)
			remaining = append(remaining, it)
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

// CheckRestoreConflicts — возвращает пути оригинальных файлов, которые уже существуют
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

// List — возвращает список доступных restore ID, отсортированных от новых к старым.
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
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].time.After(infos[j].time)
	})
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.id)
	}
	return ids, nil
}

// Cleanup removes quarantine entries older than maxAgeDays.
// Returns number of deleted quarantines and total freed bytes.
func Cleanup(maxAgeDays int) (int, int64, error) {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	return cleanupQuarantines(func(createdAt time.Time) bool {
		return createdAt.IsZero() || createdAt.Before(cutoff)
	})
}

// CleanupAll removes all quarantine entries regardless of age.
// Returns number of deleted quarantines and total freed bytes.
func CleanupAll() (int, int64, error) {
	return cleanupQuarantines(func(createdAt time.Time) bool {
		return true
	})
}

// cleanupQuarantines removes quarantine entries based on a predicate function.
// The predicate receives the creation time of a quarantine and returns true if it should be deleted.
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
			}
			deleted++
			freed += size
		}
	}

	if len(errs) > 0 {
		slog.Warn("quarantine cleanup: some directories could not be removed", "errors", len(errs))
	}
	return deleted, freed, nil
}

// Delete removes a specific quarantine entry by ID.
// Returns freed bytes or error if not found.
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

// GetLast — возвращает последний restore ID
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
	if id == "" {
		return fmt.Errorf("invalid restore id")
	}
	for _, r := range id {
		if (r >= '0' && r <= '9') || r == '-' {
			continue
		}
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
		rootCleanSep := rootClean
		if !strings.HasSuffix(rootCleanSep, string(filepath.Separator)) {
			rootCleanSep += string(filepath.Separator)
		}
		if strings.HasPrefix(lowerClean, rootCleanSep) || lowerClean == rootClean {
			return true
		}
	}
	// Hardcoded NVIDIA path used by scanner
	if strings.HasPrefix(lowerClean, `c:\nvidia`) {
		return true
	}
	return false
}

func removeRetry(path string) error {
	for i := 0; i < 5; i++ {
		if err := os.Remove(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return os.Remove(path)
}

func removeAllRetry(path string) error {
	for i := 0; i < 5; i++ {
		if err := os.RemoveAll(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return os.RemoveAll(path)
}

func writeManifest(path string, manifest *types.Manifest) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp manifest: %w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("encode manifest: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close temp manifest: %w", err)
	}
	_ = os.Remove(path) // allow overwrite on Windows
	return os.Rename(tmp, path)
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
	// Check if source is a symlink - refuse to follow
	if info, err := os.Lstat(src); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to follow symlink: %s", src)
	}

	// Check if destination already exists (could be symlink attack)
	if _, err := os.Lstat(dst); err == nil {
		return fmt.Errorf("destination already exists: %s", dst)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := in.Close(); err != nil {
		return err
	}
	return removeRetry(src)
}
