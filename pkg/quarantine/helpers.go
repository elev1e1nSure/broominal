package quarantine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
	if strings.ContainsRune(path, 0) {
		return false
	}
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
	if info, err := os.Lstat(src); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to follow symlink: %s", src)
	}
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
