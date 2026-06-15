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
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}
