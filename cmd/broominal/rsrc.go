//go:build windows
// +build windows

//go:generate go run github.com/akavel/rsrc@latest -ico ../../assets/app-icon.ico -o rsrc_windows.syso

package main
