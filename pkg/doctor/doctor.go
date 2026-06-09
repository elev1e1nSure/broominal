package doctor

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"golang.org/x/sys/windows"
)

// Status indicates the result of a check.
type Status string

const (
	StatusPass Status = "PASS"
	StatusWarn Status = "WARN"
	StatusFail Status = "FAIL"
)

// Check represents a single health check result.
type Check struct {
	Name       string
	Status     Status
	Detail     string
	Suggestion string // what the user can do to fix it
	FixKey     string // non-empty if an automatic fix is available
}

// Run performs fast health checks. Heavy checks (quarantine stats) are excluded
// and should be loaded separately via QuarantineStats().
func Run() []Check {
	var checks []Check

	checks = append(checks, checkDirCached(quarantine.BaseDir(), i18n.T("check_quarantine_dir")))
	checks = append(checks, checkDirCached(report.BaseDir(), i18n.T("check_reports_dir")))
	checks = append(checks, checkDirCached(config.Dir(), i18n.T("check_config_dir")))
	checks = append(checks, checkEnvDir("TEMP", i18n.T("check_temp_dir")))
	checks = append(checks, checkEnvDir("USERPROFILE", i18n.T("check_userprofile_dir")))
	checks = append(checks, checkManifests())

	return checks
}

// QuarantineStats returns quarantine size/file stats. This is a heavy operation
// that should be called lazily when needed.
func QuarantineStats() Check {
	return checkQuarantineStats()
}

// IsAdmin returns true if the current process has elevated privileges.
func IsAdmin() bool {
	cmd := exec.Command("cmd", "/c", "net", "session")
	return cmd.Run() == nil
}

// cacheEntry stores cached check results
type cacheEntry struct {
	Path      string
	Name      string
	Timestamp int64
}

var (
	checkCache   = make(map[string]cacheEntry)
	checkCacheMu sync.RWMutex
)

func checkDirCached(path, name string) Check {
	// Check cache first
	checkCacheMu.RLock()
	entry, ok := checkCache[path]
	checkCacheMu.RUnlock()
	if ok {
		// Cache valid for 24 hours
		if nowUnix()-entry.Timestamp < 86400 {
			return Check{
				Name:   entry.Name,
				Status: StatusPass,
				Detail: entry.Path,
			}
		}
	}

	// Run actual check
	result := checkDir(path, name)

	// Cache successful results
	if result.Status == StatusPass {
		checkCacheMu.Lock()
		checkCache[path] = cacheEntry{
			Path:      path,
			Name:      name,
			Timestamp: nowUnix(),
		}
		checkCacheMu.Unlock()
	}

	return result
}

func nowUnix() int64 {
	return time.Now().Unix()
}

func checkDir(path, name string) Check {
	if _, err := os.Stat(path); err != nil {
		// try to create
		if err := os.MkdirAll(path, 0700); err != nil {
			return Check{
				Name:       name,
				Status:     StatusFail,
				Detail:     fmt.Sprintf("%s: %v", path, err),
				Suggestion: i18n.T("suggest_check_permissions"),
			}
		}
	}
	// test write
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return Check{
			Name:       name,
			Status:     StatusFail,
			Detail:     fmt.Sprintf(i18n.T("dir_not_writable"), path, err),
			Suggestion: i18n.T("suggest_check_permissions"),
		}
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(testFile)
	}()
	return Check{
		Name:   name,
		Status: StatusPass,
		Detail: path,
	}
}

func checkEnvDir(env, name string) Check {
	val := os.Getenv(env)
	if val == "" {
		return Check{
			Name:       name,
			Status:     StatusFail,
			Detail:     env + " not set",
			Suggestion: i18n.T("suggest_env_missing"),
		}
	}
	info, err := os.Stat(val)
	if err != nil || !info.IsDir() {
		return Check{
			Name:       name,
			Status:     StatusFail,
			Detail:     fmt.Sprintf("%s is not accessible: %v", val, err),
			Suggestion: i18n.T("suggest_env_missing"),
		}
	}
	return Check{
		Name:   name,
		Status: StatusPass,
		Detail: val,
	}
}

func checkManifests() Check {
	qDir := quarantine.BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Check{
				Name:   i18n.T("check_manifests"),
				Status: StatusPass,
				Detail: i18n.T("no_backups_yet"),
			}
		}
		return Check{
			Name:       i18n.T("check_manifests"),
			Status:     StatusFail,
			Detail:     err.Error(),
			Suggestion: i18n.T("suggest_check_permissions"),
		}
	}
	var valid int
	var invalid int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mf := filepath.Join(qDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(mf)
		if err != nil {
			if os.IsNotExist(err) {
				_ = os.RemoveAll(filepath.Join(qDir, e.Name()))
				continue
			}
			invalid++
			continue
		}
		var m types.Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			invalid++
		} else {
			valid++
		}
	}
	if invalid > 0 {
		return Check{
			Name:       i18n.T("check_manifests"),
			Status:     StatusWarn,
			Detail:     fmt.Sprintf(i18n.T("invalid_manifests"), invalid),
			Suggestion: i18n.T("suggest_remove_damaged"),
		}
	}
	if valid == 0 && len(entries) == 0 {
		return Check{
			Name:   i18n.T("check_manifests"),
			Status: StatusPass,
			Detail: i18n.T("no_backups_yet"),
		}
	}
	return Check{
		Name:   i18n.T("check_manifests"),
		Status: StatusPass,
		Detail: fmt.Sprintf(i18n.T("valid_backups"), valid),
	}
}

// Fix attempts to automatically fix a problem identified by FixKey.
// Currently only "admin" is supported.
func Fix(fixKey string) (string, error) {
	switch fixKey {
	case "admin":
		// Relaunch the current executable with elevated privileges
		exe, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("cannot locate executable: %w", err)
		}
		verbPtr, _ := windows.UTF16PtrFromString("runas")
		exePtr, _ := windows.UTF16PtrFromString(exe)
		if err := windows.ShellExecute(0, verbPtr, exePtr, nil, nil, windows.SW_NORMAL); err != nil {
			return "", fmt.Errorf("failed to elevate: %w", err)
		}
		return "Restarting as administrator...", nil
	default:
		return "", fmt.Errorf("no automatic fix for %s", fixKey)
	}
}

func checkQuarantineStats() Check {
	qDir := quarantine.BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Check{
				Name:   i18n.T("check_stats"),
				Status: StatusPass,
				Detail: "0",
			}
		}
		return Check{
			Name:       i18n.T("check_stats"),
			Status:     StatusFail,
			Detail:     err.Error(),
			Suggestion: i18n.T("suggest_check_permissions"),
		}
	}
	var totalSize int64
	var totalFiles int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if err := filepath.WalkDir(filepath.Join(qDir, e.Name()), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			totalSize += info.Size()
			totalFiles++
			return nil
		}); err != nil {
			slog.Warn("doctor: failed to walk quarantine dir", "path", filepath.Join(qDir, e.Name()), "error", err)
		}
	}
	return Check{
		Name:   i18n.T("check_stats"),
		Status: StatusPass,
		Detail: fmt.Sprintf(i18n.T("backups_files_size"), len(entries), totalFiles, util.FormatSize(totalSize)),
	}
}
