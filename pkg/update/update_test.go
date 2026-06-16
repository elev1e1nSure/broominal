package update

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChecksum(t *testing.T) {
	data := `
1234567890abcdef file.exe
0987654321fedcba other.exe
`
	// Match exactly
	hash, err := parseChecksum(data, "file.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "1234567890abcdef" {
		t.Errorf("expected hash 1234567890abcdef, got %s", hash)
	}

	// Not found
	_, err = parseChecksum(data, "missing.exe")
	if err == nil {
		t.Errorf("expected error for missing file")
	}

	// Single column format
	singleData := `abcdef123456`
	hash2, err := parseChecksum(singleData, "anything.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash2 != "abcdef123456" {
		t.Errorf("expected hash abcdef123456, got %s", hash2)
	}
}

func TestVerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.exe")

	content := []byte("hello world")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Calculate true hash
	h := sha256.New()
	h.Write(content)
	trueHash := hex.EncodeToString(h.Sum(nil))

	// Mock server hosting the checksum
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "bad") {
			w.Write([]byte("invalidhash test.exe\n"))
		} else {
			w.Write([]byte(trueHash + " test.exe\n"))
		}
	}))
	defer server.Close()

	asset := &Asset{
		BrowserDownloadURL: server.URL + "/good.sha256",
	}

	if err := verifyChecksum(asset, filePath); err != nil {
		t.Errorf("expected valid checksum, got: %v", err)
	}

	badAsset := &Asset{
		BrowserDownloadURL: server.URL + "/bad.sha256",
	}
	if err := verifyChecksum(badAsset, filePath); err == nil {
		t.Errorf("expected error for invalid checksum")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content mismatch, expected %s, got %s", string(content), string(dstContent))
	}
}
