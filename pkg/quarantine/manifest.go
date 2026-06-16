package quarantine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

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
		if m == nil {
			continue
		}
		infos = append(infos, entryInfo{id: id, time: m.CreatedAt})
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].time.After(infos[j].time)
	})
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.id)
	}
	return ids, nil
}

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

func writeManifest(path string, manifest *types.Manifest) error {
	// Use CreateTemp for a unique temp name — avoids collisions if a previous
	// write crashed and left a stale .tmp file behind.
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "manifest-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp manifest: %w", err)
	}
	tmp := f.Name()
	// Clean up temp file on any error path.
	var writeErr error
	defer func() {
		if writeErr != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if writeErr = enc.Encode(manifest); writeErr != nil {
		return fmt.Errorf("encode manifest: %w", writeErr)
	}
	if writeErr = f.Sync(); writeErr != nil {
		return fmt.Errorf("sync temp manifest: %w", writeErr)
	}
	if writeErr = f.Close(); writeErr != nil {
		return fmt.Errorf("close temp manifest: %w", writeErr)
	}
	// On Windows Rename cannot overwrite an existing file — remove first.
	// There is a narrow window between Remove and Rename where the manifest
	// does not exist; the lock in Move/Restore prevents concurrent readers.
	_ = os.Remove(path)
	if writeErr = os.Rename(tmp, path); writeErr != nil {
		return fmt.Errorf("rename manifest: %w", writeErr)
	}
	return nil
}
