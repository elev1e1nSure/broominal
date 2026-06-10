package pathman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

func getUserPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	val, _, err := k.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func setUserPath(val string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue("Path", val)
}

func pathEntries(path string) []string {
	var out []string
	for _, p := range strings.Split(path, ";") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func IsInPath() (bool, error) {
	dir, err := exeDir()
	if err != nil {
		return false, err
	}
	pathStr, err := getUserPath()
	if err != nil {
		return false, err
	}
	for _, entry := range pathEntries(pathStr) {
		if strings.EqualFold(filepath.Clean(entry), filepath.Clean(dir)) {
			return true, nil
		}
	}
	return false, nil
}

func AddToPath() error {
	dir, err := exeDir()
	if err != nil {
		return err
	}
	pathStr, err := getUserPath()
	if err != nil {
		return err
	}
	entries := pathEntries(pathStr)
	for _, entry := range entries {
		if strings.EqualFold(filepath.Clean(entry), filepath.Clean(dir)) {
			return nil // already present
		}
	}
	entries = append(entries, dir)
	newPath := strings.Join(entries, ";")
	return setUserPath(newPath)
}

func RemoveFromPath() error {
	dir, err := exeDir()
	if err != nil {
		return err
	}
	pathStr, err := getUserPath()
	if err != nil {
		return err
	}
	entries := pathEntries(pathStr)
	var filtered []string
	found := false
	for _, entry := range entries {
		if strings.EqualFold(filepath.Clean(entry), filepath.Clean(dir)) {
			found = true
			continue
		}
		filtered = append(filtered, entry)
	}
	if !found {
		return fmt.Errorf("directory not found in PATH")
	}
	newPath := strings.Join(filtered, ";")
	return setUserPath(newPath)
}
