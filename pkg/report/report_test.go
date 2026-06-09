package report

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func TestSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	result := &types.ScanResult{
		Categories: []types.CategorySummary{
			{Category: "Temp", Size: 1024, Files: 2, Risk: types.RiskSafe},
		},
		TotalSize:  1024,
		SafeSize:   1024,
		ReviewSize: 0,
		DangerSize: 0,
	}
	cleaned := &types.CleanResult{
		RestoreID: "abc-123",
		Freed:     512,
		Files:     1,
	}

	path, err := Save(result, cleaned)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("report file should exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var report types.ReportData
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("parse report: %v", err)
	}
	if report.Result.TotalSize != 1024 {
		t.Errorf("TotalSize = %d, want 1024", report.Result.TotalSize)
	}
	if report.Cleaned == nil || report.Cleaned.RestoreID != "abc-123" {
		t.Errorf("Cleaned.RestoreID = %v, want abc-123", report.Cleaned)
	}
}

func TestSaveNoClean(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	result := &types.ScanResult{
		Categories: []types.CategorySummary{
			{Category: "Temp", Size: 100, Files: 1, Risk: types.RiskSafe},
		},
		TotalSize: 100,
	}

	path, err := Save(result, nil)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	var report types.ReportData
	_ = json.Unmarshal(data, &report)
	if report.Cleaned != nil {
		t.Error("Cleaned should be nil")
	}
}

func TestPrintSummary(t *testing.T) {
	result := &types.ScanResult{
		TotalSize:  2048,
		SafeSize:   1024,
		ReviewSize: 512,
		DangerSize: 256,
	}
	cleaned := &types.CleanResult{
		RestoreID: "id-1",
		Freed:     1024,
		Files:     3,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintSummary(result, cleaned)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "Broominal Report") {
		t.Error("output should contain 'Broominal Report'")
	}
	if !strings.Contains(out, util.FormatSize(2048)) {
		t.Errorf("output should contain total size")
	}
	if !strings.Contains(out, "id-1") {
		t.Error("output should contain restore ID")
	}
}

func TestPrintSummaryNoClean(t *testing.T) {
	result := &types.ScanResult{
		TotalSize: 100,
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintSummary(result, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if strings.Contains(out, "Freed") {
		t.Error("output should not contain 'Freed' when cleaned is nil")
	}
}

func TestBaseDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	want := filepath.Join(tmp, "broominal", "reports")
	if BaseDir() != want {
		t.Errorf("BaseDir() = %q, want %q", BaseDir(), want)
	}
}

func TestDirPermissions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	result := &types.ScanResult{}
	_, err := Save(result, nil)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(BaseDir())
	if err != nil {
		t.Fatalf("stat reports dir: %v", err)
	}
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0700 {
			t.Errorf("reports dir permissions = %04o, want 0700", perm)
		}
	}
}
