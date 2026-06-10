//go:build windows

package util

import (
	"errors"

	"golang.org/x/sys/windows"
)

// IsFileLocked reports whether err indicates that a file is locked or in use by another process.
func IsFileLocked(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_LOCK_VIOLATION)
}
