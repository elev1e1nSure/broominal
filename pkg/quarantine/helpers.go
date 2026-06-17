package quarantine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sys/windows"
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
	// Only allow restore to user-writable directories.
	// Intentionally excludes SystemRoot/WINDIR/SYSTEMDRIVE to prevent
	// privilege escalation via manifest poisoning (e.g. writing into System32).
	allowedRoots := []string{
		os.Getenv("TEMP"),
		os.Getenv("TMP"),
		os.Getenv("USERPROFILE"),
		os.Getenv("LOCALAPPDATA"),
		os.Getenv("APPDATA"),
		os.Getenv("ProgramData"),
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
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

// ErrScheduledForReboot is returned when a locked file is scheduled for deletion on next reboot.
var ErrScheduledForReboot = errors.New("scheduled for reboot")

func removeOnReboot(path string) error {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(ptr, nil, windows.MOVEFILE_DELAY_UNTIL_REBOOT)
}

func removeAllOnReboot(dir string) error {
	var errs []error
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if e := removeOnReboot(p); e != nil {
				errs = append(errs, e)
			}
		}
		return nil
	})

	var dirs []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, d := range dirs {
		if e := removeOnReboot(d); e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// forceRemoveAll resets file attributes via SetFileAttributes before removal.
// os.Chmod only clears FILE_ATTRIBUTE_READONLY; HIDDEN and SYSTEM flags also
// cause "Access is denied" on RemoveAll. FILE_ATTRIBUTE_NORMAL clears all of
// them in one call. Directories get FILE_ATTRIBUTE_DIRECTORY instead, since
// FILE_ATTRIBUTE_NORMAL is not valid for directory entries on Windows.
func forceRemoveAll(path string) error {
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		ptr, err := windows.UTF16PtrFromString(p)
		if err != nil {
			return nil
		}
		attr := uint32(windows.FILE_ATTRIBUTE_NORMAL)
		if d.IsDir() {
			attr = windows.FILE_ATTRIBUTE_DIRECTORY
		}
		_ = windows.SetFileAttributes(ptr, attr)
		return nil
	})

	err := removeAllRetry(path)
	if err != nil {
		if rebootErr := removeAllOnReboot(path); rebootErr == nil {
			return ErrScheduledForReboot
		}
		return err
	}
	return nil
}

func uniquePath(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s_%d%s", base, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Lstat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

// uniquePathAtomic generates a guaranteed-unique path by appending an
// atomic sequence number. The caller must hold the journal/manifest lock
// so that the generated path is used before another goroutine allocates
// the next sequence number.
func uniquePathAtomic(dir, srcPath string, seq *int64) string {
	n := atomic.AddInt64(seq, 1)
	base := filepath.Base(srcPath)
	return filepath.Join(dir, fmt.Sprintf("%s_%d.brm", base, n))
}

func copyAndDelete(ctx context.Context, src, dst string) error {
	srcStat, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if srcStat.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to follow symlink: %s", src)
	}
	expectedSize := srcStat.Size()

	if _, err := os.Lstat(dst); err == nil {
		return fmt.Errorf("destination already exists: %s", dst)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	buf := make([]byte, 1024*1024)
	for {
		select {
		case <-ctx.Done():
			_ = out.Close()
			_ = in.Close()
			_ = removeRetry(dst)
			return ctx.Err()
		default:
		}
		n, readErr := in.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	if err := out.Sync(); err != nil {
		return err
	}

	dstStat, err := out.Stat()
	if err != nil {
		return err
	}
	if dstStat.Size() != expectedSize {
		_ = out.Close()
		_ = in.Close()
		_ = removeRetry(dst)
		return fmt.Errorf("size mismatch after copy: expected %d, got %d", expectedSize, dstStat.Size())
	}

	if err := out.Close(); err != nil {
		return err
	}
	_ = in.Close()
	err = removeRetry(src)
	if err != nil {
		_ = removeRetry(dst)
		return err
	}
	return nil
}

func getDiskFreeSpace(dirPath string) (int64, error) {
	var freeBytesAvailableToCaller, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	dirPtr, err := windows.UTF16PtrFromString(dirPath)
	if err != nil {
		return 0, err
	}
	err = windows.GetDiskFreeSpaceEx(dirPtr, &freeBytesAvailableToCaller, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, err
	}
	return int64(freeBytesAvailableToCaller), nil
}
