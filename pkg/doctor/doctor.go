package doctor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
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

// Run performs all health checks including quarantine stats.
func Run() []Check {
	var checks []Check

	checks = append(checks, checkDirCached(quarantine.BaseDir(), i18n.T("check_quarantine_dir")))
	checks = append(checks, checkDirCached(report.BaseDir(), i18n.T("check_reports_dir")))
	checks = append(checks, checkDirCached(config.Dir(), i18n.T("check_config_dir")))
	checks = append(checks, checkEnvDir("TEMP", i18n.T("check_temp_dir")))
	checks = append(checks, checkEnvDir("USERPROFILE", i18n.T("check_userprofile_dir")))
	checks = append(checks, checkManifests())
	checks = append(checks, checkQuarantineStats())

	return checks
}

// IsAdmin returns true if the current process has elevated privileges.
func IsAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// cacheEntry records a successful permission probe so the next doctor run can
// skip a directory-write test that touches the filesystem.
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
	// Avoid re-probing write permissions on every doctor run — results for a
	// given path are stable on the same machine, so a 24h cache is a safe
	// freshness window that matches what users expect from a health check.
	checkCacheMu.RLock()
	entry, ok := checkCache[path]
	checkCacheMu.RUnlock()
	if ok {
		// 86_400 seconds = 24h. See comment above.
		if nowUnix()-entry.Timestamp < 86400 {
			return Check{
				Name:   entry.Name,
				Status: StatusPass,
				Detail: entry.Path,
			}
		}
	}

	// Cold path: the probe below actually touches the filesystem, so we
	// only run it when the cache is missing or stale.
	result := checkDir(path, name)

	// Only cache successes. Failures often resolve themselves (permissions
	// corrected, env var set, etc.) and we want the next run to re-probe
	// rather than report a stale failure.
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
		// Auto-create missing dirs so first-run "doctor" doesn't report
		// false negatives for paths the user simply hasn't touched yet —
		// the real failure we want to surface is a permission denial.
		if err := os.MkdirAll(path, 0700); err != nil {
			return Check{
				Name:       name,
				Status:     StatusFail,
				Detail:     fmt.Sprintf("%s: %v", path, err),
				Suggestion: i18n.T("suggest_check_permissions"),
			}
		}
	}
	// A directory that exists may still be read-only for this user (ACLs,
	// ACL-inherited deny, etc.) — an actual file write is the cheapest
	// reliable signal that the cleaner can deposit quarantine data here.
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
	var dead int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		deadPath := filepath.Join(qDir, e.Name(), "manifest.dead")
		if _, err := os.Stat(deadPath); err == nil {
			dead++
			continue
		}

		mf := filepath.Join(qDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(mf)
		if err != nil {
			// Don't delete dirs with missing manifest — they may be mid-write
			// (Move() creates the dir before writing the manifest). Report as
			// invalid so the user can clean them up via 'quarantine-cleanup'.
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
	if dead > 0 {
		return Check{
			Name:   i18n.T("check_manifests"),
			Status: StatusFail,
			Detail: fmt.Sprintf(i18n.T("dead_manifests"), dead),
			FixKey: "purge_dead",
		}
	}
	if invalid > 0 {
		return Check{
			Name:   i18n.T("check_manifests"),
			Status: StatusWarn,
			Detail: fmt.Sprintf(i18n.T("invalid_manifests"), invalid),
			FixKey: "repair_damaged",
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

// Fix is a dispatcher from a doctor check to its self-repair action. Only
// "admin" is wired up today; the remaining checks report suggestions for the
// user to follow manually because their fixes require steps we can't take
// programmatically (folder permissions, env-var setup, etc.).
func Fix(fixKey string) (string, error) {
	switch fixKey {
	case "admin":
		// Trigger the UAC elevation prompt for the current binary via
		// ShellExecute's "runas" verb — this is the supported way to ask
		// Windows to relaunch us elevated without shipping a manifest.
		exe, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("cannot locate executable: %w", err)
		}
		verbPtr, _ := windows.UTF16PtrFromString("runas")
		// If we're hosted by Windows Terminal, spawn the elevated copy as a
		// new WT tab instead of a standalone conhost window — that matches
		// what WT users expect and keeps the TUI rendering pipeline working.
		if os.Getenv("WT_SESSION") != "" {
			wtArgs := fmt.Sprintf(`new-tab -- "%s"`, exe)
			wtPtr, _ := windows.UTF16PtrFromString("wt.exe")
			argPtr, _ := windows.UTF16PtrFromString(wtArgs)
			if err := windows.ShellExecute(0, verbPtr, wtPtr, argPtr, nil, windows.SW_NORMAL); err == nil {
				return "Restarting as administrator...", nil
			}
		}
		exePtr, _ := windows.UTF16PtrFromString(exe)
		if err := windows.ShellExecute(0, verbPtr, exePtr, nil, nil, windows.SW_NORMAL); err != nil {
			return "", fmt.Errorf("failed to elevate: %w", err)
		}
		return "Restarting as administrator...", nil
	case "repair_damaged":
		repaired, dead, err := quarantine.RepairDamaged()
		if err != nil {
			return "", fmt.Errorf("repair failed: %w", err)
		}
		return fmt.Sprintf(i18n.T("repair_damaged_ok"), repaired, dead), nil
	case "purge_dead":
		n, err := quarantine.PurgeDead()
		if err != nil {
			if errors.Is(err, quarantine.ErrScheduledForReboot) {
				return i18n.T("purge_scheduled_reboot"), nil
			}
			return "", fmt.Errorf("purge failed: %w", err)
		}
		return fmt.Sprintf(i18n.T("purge_damaged_ok"), n), nil
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
