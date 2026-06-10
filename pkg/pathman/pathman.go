package pathman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

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
	if err := k.SetStringValue("Path", val); err != nil {
		return err
	}
	return broadcastEnvChange()
}

func broadcastEnvChange() error {
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x1A
		SMTO_ABORTIFHUNG = 0x0002
	)
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")
	var result uintptr
	ptr, _ := syscall.UTF16PtrFromString("Environment")
	_, _, e1 := proc.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(ptr)),
		uintptr(SMTO_ABORTIFHUNG),
		5000,
		uintptr(unsafe.Pointer(&result)),
	)
	if e1 != nil && e1.Error() != "The operation completed successfully." {
		return e1
	}
	return nil
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
