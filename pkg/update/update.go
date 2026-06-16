package update

import (
	"crypto/sha256"
	"encoding/hex"
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

	// Limit response size to 1 MB to prevent OOM from malicious/compromised servers.
	var release Release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1*1024*1024)).Decode(&release); err != nil {
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
	// Fallback: any .exe if no "windows"-named asset found
	if asset == nil {
		for i := range release.Assets {
			a := &release.Assets[i]
			if strings.HasSuffix(a.Name, ".exe") {
				asset = a
				break
			}
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

	tmpPath := filepath.Join(tmpDir, asset.Name)

	// Reject download if server claims the file is unreasonably large.
	const maxBinarySize = 100 * 1024 * 1024 // 100 MB
	if asset.Size > maxBinarySize {
		return "", fmt.Errorf("update binary too large: %d bytes (max %d)", asset.Size, maxBinarySize)
	}

	client := &http.Client{Timeout: 10 * time.Minute}
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

	// Limit download to maxBinarySize even if Content-Length header is absent/wrong.
	if _, err := io.Copy(f, io.LimitReader(resp.Body, maxBinarySize+1)); err != nil {
		return "", fmt.Errorf("failed to write update file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Checksum verification is mandatory. Releases without a checksum asset
	// are rejected to prevent unsigned binary execution.
	var checksumAsset *Asset
	for i := range release.Assets {
		a := &release.Assets[i]
		if a.Name == asset.Name+".sha256" || a.Name == "checksums.txt" {
			checksumAsset = a
			break
		}
	}
	if checksumAsset == nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("update: no checksum asset found for %s — refusing to install unverified binary", asset.Name)
	}
	if err := verifyChecksum(checksumAsset, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("checksum verification failed: %w", err)
	}
	slog.Info("update: checksum verified")

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

	// Copy new executable to current location (cross-drive safe)
	if err := copyFile(updatePath, exePath); err != nil {
		_ = os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	_ = os.Remove(backupPath)
	_ = os.Remove(updatePath)
	return nil
}

func parseChecksum(data, fileName string) (string, error) {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 1 {
			return fields[0], nil
		}
		if len(fields) >= 2 && strings.EqualFold(fields[1], fileName) {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found", fileName)
}

func verifyChecksum(asset *Asset, filePath string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status for checksum: %d", resp.StatusCode)
	}
	// Checksum files are tiny — limit to 4 KB to prevent OOM.
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
	if err != nil {
		return fmt.Errorf("failed to read checksum: %w", err)
	}
	expected, err := parseChecksum(string(data), filepath.Base(filePath))
	if err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Close()
}
