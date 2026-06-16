//go:build windows

package taskscheduler

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const TaskName = "broominal quarantine-cleanup"

// Set registers a daily 03:00 task that runs
// `broominal quarantine clean --yes --max-age-days N`, so users who never
// invoke cleanup manually still have the quarantine size bounded automatically.
func Set(maxAgeDays int) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	taskRun := fmt.Sprintf(`"%s" quarantine clean --yes --max-age-days %d`, exe, maxAgeDays)
	cmd := exec.Command("schtasks",
		"/create",
		"/tn", TaskName,
		"/tr", taskRun,
		"/sc", "daily",
		"/st", "03:00",
		"/f",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks: %s", string(out))
	}
	return nil
}

// Delete removes the scheduled task. Silently succeeds if the task does not exist.
func Delete() error {
	if !Exists() {
		return nil
	}
	cmd := exec.Command("schtasks", "/delete", "/tn", TaskName, "/f")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks: %s", string(out))
	}
	return nil
}

// Exists is used to avoid a noisy schtasks error when toggling a feature
// whose task was never created — cheaper and clearer than parsing schtasks' stderr.
func Exists() bool {
	cmd := exec.Command("schtasks", "/query", "/tn", TaskName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
