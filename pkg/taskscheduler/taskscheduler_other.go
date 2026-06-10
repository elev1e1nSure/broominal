//go:build !windows

package taskscheduler

func Set(_ int) error { return nil }
func Delete() error   { return nil }
func Exists() bool    { return false }
