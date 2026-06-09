package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
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
	Name   string
	Status Status
	Detail string
}

// Run performs all health checks and returns the results.
func Run() []Check {
	var checks []Check

	checks = append(checks, checkAdmin())
	checks = append(checks, checkDir(quarantine.BaseDir(), "Quarantine"))
	checks = append(checks, checkDir(report.BaseDir(), "Reports"))
	checks = append(checks, checkDir(config.Dir(), "Config"))
	checks = append(checks, checkEnvDir("TEMP", "Temp"))
	checks = append(checks, checkEnvDir("USERPROFILE", "User profile"))
	checks = append(checks, checkManifests())
	checks = append(checks, checkQuarantineStats())

	return checks
}

func checkAdmin() Check {
	cmd := exec.Command("cmd", "/c", "net", "session")
	if err := cmd.Run(); err != nil {
		return Check{
			Name:   "Admin privileges",
			Status: StatusWarn,
			Detail: "Not running as administrator (some paths may be inaccessible)",
		}
	}
	return Check{
		Name:   "Admin privileges",
		Status: StatusPass,
		Detail: "Running as administrator",
	}
}

func checkDir(path, name string) Check {
	if _, err := os.Stat(path); err != nil {
		// try to create
		if err := os.MkdirAll(path, 0700); err != nil {
			return Check{
				Name:   name + " directory",
				Status: StatusFail,
				Detail: fmt.Sprintf("%s: %v", path, err),
			}
		}
	}
	// test write
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return Check{
			Name:   name + " directory",
			Status: StatusFail,
			Detail: fmt.Sprintf("Cannot write to %s: %v", path, err),
		}
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(testFile)
	}()
	return Check{
		Name:   name + " directory",
		Status: StatusPass,
		Detail: path,
	}
}

func checkEnvDir(env, name string) Check {
	val := os.Getenv(env)
	if val == "" {
		return Check{
			Name:   name + " environment",
			Status: StatusFail,
			Detail: env + " not set",
		}
	}
	info, err := os.Stat(val)
	if err != nil || !info.IsDir() {
		return Check{
			Name:   name + " directory",
			Status: StatusFail,
			Detail: fmt.Sprintf("%s is not accessible: %v", val, err),
		}
	}
	return Check{
		Name:   name + " directory",
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
				Name:   "Quarantine manifests",
				Status: StatusPass,
				Detail: "No quarantines yet",
			}
		}
		return Check{
			Name:   "Quarantine manifests",
			Status: StatusFail,
			Detail: err.Error(),
		}
	}
	var invalid int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mf := filepath.Join(qDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(mf)
		if err != nil {
			invalid++
			continue
		}
		var m types.Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			invalid++
		}
	}
	if invalid > 0 {
		return Check{
			Name:   "Quarantine manifests",
			Status: StatusWarn,
			Detail: fmt.Sprintf("%d invalid manifest(s)", invalid),
		}
	}
	return Check{
		Name:   "Quarantine manifests",
		Status: StatusPass,
		Detail: fmt.Sprintf("%d valid quarantine(s)", len(entries)),
	}
}

func checkQuarantineStats() Check {
	qDir := quarantine.BaseDir()
	entries, err := os.ReadDir(qDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Check{
				Name:   "Quarantine stats",
				Status: StatusPass,
				Detail: "0 quarantines",
			}
		}
		return Check{
			Name:   "Quarantine stats",
			Status: StatusFail,
			Detail: err.Error(),
		}
	}
	var totalSize int64
	var totalFiles int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		_ = filepath.Walk(filepath.Join(qDir, e.Name()), func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			totalSize += info.Size()
			totalFiles++
			return nil
		})
	}
	return Check{
		Name:   "Quarantine stats",
		Status: StatusPass,
		Detail: fmt.Sprintf("%d quarantines, %d files, %s", len(entries), totalFiles, scanner.FormatSize(totalSize)),
	}
}

