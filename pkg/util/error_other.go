//go:build !windows

package util

// IsFileLocked reports whether err indicates that a file is locked or in use by another process.
// On non-Windows platforms this always returns false.
func IsFileLocked(err error) bool {
	return false
}
