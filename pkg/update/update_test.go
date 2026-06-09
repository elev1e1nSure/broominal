package update

import (
	"testing"
)

func TestHasInternet(t *testing.T) {
	// This test may fail in offline environments
	result := HasInternet()
	t.Logf("HasInternet: %v", result)
}

func TestParseChecksum(t *testing.T) {
	tests := []struct {
		data     string
		fileName string
		expected string
	}{
		{"abc123  file.exe\n", "file.exe", "abc123"},
		{"abc123\n", "file.exe", "abc123"},
		{"abc123  file.exe\nxyz789  other.exe\n", "other.exe", "xyz789"},
	}
	for _, tt := range tests {
		got, err := parseChecksum(tt.data, tt.fileName)
		if err != nil {
			t.Fatalf("parseChecksum error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("parseChecksum(%q, %q) = %q, want %q", tt.data, tt.fileName, got, tt.expected)
		}
	}
}

func TestCheckForUpdates(t *testing.T) {
	release, err := CheckForUpdates("dev")
	if err != nil {
		t.Logf("CheckForUpdates error (expected in offline): %v", err)
		return
	}
	if release != nil {
		t.Logf("New version available: %s", release.TagName)
	} else {
		t.Log("No update available")
	}
}
