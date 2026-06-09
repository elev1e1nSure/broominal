package update

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	repoOwner = "elev1e1nSure"
	repoName  = "broominal"
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Assets  []Asset
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckForUpdates checks if a new version is available on GitHub
func CheckForUpdates(currentVersion string) (*Release, error) {
	if !HasInternet() {
		return nil, fmt.Errorf("no internet connection")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release info: %w", err)
	}

	slog.Info("update check debug", "latest_tag", release.TagName, "current_version", currentVersion)

	// Normalize and compare versions using semver
	currentVer := strings.TrimSpace(strings.TrimPrefix(currentVersion, "v"))
	latestVer := strings.TrimSpace(strings.TrimPrefix(release.TagName, "v"))

	// Parse versions
	currentSemver, err := semver.NewVersion(currentVer)
	if err != nil {
		// If current version is not valid semver (e.g., "dev"), skip update check
		if currentVer == "dev" || currentVer == "" {
			return nil, nil
		}
		return nil, fmt.Errorf("invalid current version: %w", err)
	}

	latestSemver, err := semver.NewVersion(latestVer)
	if err != nil {
		return nil, fmt.Errorf("invalid latest version: %w", err)
	}

	if !currentSemver.LessThan(latestSemver) {
		slog.Info("update check: no update available", "current", currentVer, "latest", latestVer)
		return nil, nil // No update available
	}

	slog.Info("update check: update available", "latest", release.TagName)
	return &release, nil
}

// HasInternet checks if there's an internet connection by trying GitHub.
func HasInternet() bool {
	conn, err := net.DialTimeout("tcp", "api.github.com:443", 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// DownloadUpdate downloads the Windows binary from the release
func DownloadUpdate(release *Release) (string, error) {
	// Find the Windows asset
	var asset *Asset
	for i := range release.Assets {
		a := &release.Assets[i]
		if strings.Contains(a.Name, "windows") && strings.HasSuffix(a.Name, ".exe") {
			asset = a
			break
		}
	}
	if asset == nil {
		return "", fmt.Errorf("no Windows binary found in release")
	}

	// Download to temp directory
	tmpDir := os.Getenv("TEMP")
	if tmpDir == "" {
		tmpDir = "."
	}

	tmpPath := filepath.Join(tmpDir, "broominal-update.exe")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(asset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write update file: %w", err)
	}

	return tmpPath, nil
}

// InstallUpdate replaces the current executable with the new one
func InstallUpdate(updatePath string) error {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Move current executable to backup
	backupPath := exePath + ".old"
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// Move new executable to current location
	if err := os.Rename(updatePath, exePath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Remove backup
	_ = os.Remove(backupPath)

	return nil
}
