//go:build windows

package taskscheduler

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const TaskName = "broominal quarantine-cleanup"

// Set creates or updates the daily scheduled task.
// Runs `broominal quarantine-cleanup --force --max-age-days <n>` every day at 03:00.
func Set(maxAgeDays int) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	taskRun := fmt.Sprintf(`"%s" quarantine-cleanup --force --max-age-days %d`, exe, maxAgeDays)
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

// Exists reports whether the scheduled task exists.
func Exists() bool {
	cmd := exec.Command("schtasks", "/query", "/tn", TaskName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}
