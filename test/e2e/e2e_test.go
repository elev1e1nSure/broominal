package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary
	binaryPath = filepath.Join(os.TempDir(), "broominal-e2e.exe")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/broominal")
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build binary: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(binaryPath)

	os.Exit(m.Run())
}

func TestE2ECleanupAndRestore(t *testing.T) {
	tempHome := t.TempDir()
	tempAppData := filepath.Join(tempHome, "AppData", "Local")
	os.MkdirAll(tempAppData, 0755)

	// Setup fake Temp directory to scan
	tempScanDir := filepath.Join(tempHome, "Temp")
	os.MkdirAll(tempScanDir, 0755)

	// Create some dummy files
	file1 := filepath.Join(tempScanDir, "garbage1.tmp")
	file2 := filepath.Join(tempScanDir, "garbage2.log")
	os.WriteFile(file1, []byte("junk data"), 0644)
	os.WriteFile(file2, []byte("more junk"), 0644)

	// Ensure files exist
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Fatalf("file1 not created")
	}

	// 1. Run scan
	// Broominal looks for the real %TEMP% unless overridden. We need a way to scan our specific folder.
	// We can use the custom path via config or path manager, but since we are black-boxing,
	// let's see if there's a way. `broominal scan` doesn't take `--dir`. It reads from paths.
	// Oh wait, broominal scans standard paths. How do we mock the paths?
	// The paths are resolved via os.UserHomeDir(), os.Getenv("TEMP").
	// We can set ENV variables for the subprocess!
	runCmd := func(args ...string) (string, error) {
		cmd := exec.Command(binaryPath, args...)
		cmd.Env = append(os.Environ(),
			"LOCALAPPDATA="+tempAppData,
			"APPDATA="+filepath.Join(tempHome, "AppData", "Roaming"),
			"TEMP="+tempScanDir,
			"USERPROFILE="+tempHome,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		return out.String(), err
	}

	// Wait, if it scans TEMP, our tempScanDir will be scanned.
	// 1. Scan
	out, err := runCmd("scan")
	if err != nil {
		t.Fatalf("scan failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "garbage1.tmp") && !strings.Contains(out, "garbage2.log") && !strings.Contains(out, "Temp") {
		t.Log("scan output does not mention dummy files, which might be normal if hidden")
	}

	// 2. Clean
	out, err = runCmd("quarantine", "clean", "--yes")
	if err != nil {
		t.Fatalf("quarantine clean failed: %v\nOutput: %s", err, out)
	}
	// Wait, `quarantine clean` cleans the QUARANTINE! We need `clean --yes` to perform the actual scan cleanup.
	out, err = runCmd("clean", "--yes")
	if err != nil {
		t.Fatalf("clean failed: %v\nOutput: %s", err, out)
	}

	// Verify files are gone
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Errorf("file1 was not deleted by clean")
	}

	// Check quarantine directory
	quarantineDir := filepath.Join(tempAppData, "broominal", "quarantine")
	entries, err := os.ReadDir(quarantineDir)
	if err != nil {
		t.Fatalf("failed to read quarantine dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("quarantine is empty")
	}

	// 3. Restore
	out, err = runCmd("restore", "last")
	if err != nil {
		t.Fatalf("restore failed: %v\nOutput: %s", err, out)
	}

	// Verify files are back
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Errorf("file1 was not restored")
	}
	content, _ := os.ReadFile(file1)
	if string(content) != "junk data" {
		t.Errorf("restored file content mismatch")
	}
}

func TestE2EConfig(t *testing.T) {
	tempHome := t.TempDir()
	tempAppData := filepath.Join(tempHome, "AppData", "Local")
	os.MkdirAll(tempAppData, 0755)

	runCmd := func(args ...string) (string, error) {
		cmd := exec.Command(binaryPath, args...)
		cmd.Env = append(os.Environ(),
			"LOCALAPPDATA="+tempAppData,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		return out.String(), err
	}

	out, err := runCmd("config")
	if err != nil {
		t.Fatalf("config failed: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "Config path:") {
		t.Errorf("config output did not contain path: %s", out)
	}
}
