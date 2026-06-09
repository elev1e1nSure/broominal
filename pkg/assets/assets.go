package assets

import (
	"embed"
	"io/fs"
)

//go:embed app-icon.png app-icon.ico
var iconFS embed.FS

// IconPNG returns the embedded PNG icon bytes.
func IconPNG() ([]byte, error) {
	return iconFS.ReadFile("app-icon.png")
}

// IconICO returns the embedded ICO icon bytes.
func IconICO() ([]byte, error) {
	return iconFS.ReadFile("app-icon.ico")
}

// FS exposes the embedded filesystem for direct access.
func FS() fs.FS {
	return iconFS
}
