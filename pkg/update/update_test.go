package update

import (
	"testing"
)

func TestHasInternet(t *testing.T) {
	// This test may fail in offline environments
	result := HasInternet()
	t.Logf("HasInternet: %v", result)
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
