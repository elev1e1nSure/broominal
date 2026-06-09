package quarantine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/elev1e1nSure/pclean/pkg/types"
)

// BaseDir возвращает базовую директорию карантина
func BaseDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("USERPROFILE")
	}
	return filepath.Join(localAppData, "pclean", "quarantine")
}

// Move перемещает файлы в карантин и возвращает restore ID
func Move(items []types.Item) (string, int64, int, error) {
	id := uuid.New().String()
	qDir := filepath.Join(BaseDir(), id)
	if err := os.MkdirAll(qDir, 0755); err != nil {
		return "", 0, 0, fmt.Errorf("create quarantine dir: %w", err)
	}

	manifest := types.Manifest{
		ID:        id,
		CreatedAt: time.Now(),
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
		freed += it.Size
		files++
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

// Restore восстанавливает файлы из карантина по ID
func Restore(id string) error {
	qDir := filepath.Join(BaseDir(), id)
	manifestPath := filepath.Join(qDir, "manifest.json")

	mf, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("open manifest: %w", err)
	}
	defer mf.Close()

	var manifest types.Manifest
	if err := json.NewDecoder(mf).Decode(&manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}

	for _, it := range manifest.Items {
		if _, err := os.Stat(it.Quarantined); os.IsNotExist(err) {
			continue
		}
		// ensure original dir exists
		if err := os.MkdirAll(filepath.Dir(it.Original), 0755); err != nil {
			continue
		}
		if err := os.Rename(it.Quarantined, it.Original); err != nil {
			// fallback
			_ = copyAndDelete(it.Quarantined, it.Original)
		}
	}

	// remove quarantine dir after successful restore
	_ = os.RemoveAll(qDir)
	return nil
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
