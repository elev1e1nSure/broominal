package quarantine

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

func FuzzValidateID(f *testing.F) {
	f.Add("2025-06-15-120000")
	f.Add("2025-06-15-120000-2")
	f.Add("")
	f.Add("../../../etc/passwd")
	f.Add("a/b/c")
	f.Add("\x00")
	f.Add("2025-06-15; rm -rf /")

	f.Fuzz(func(t *testing.T, id string) {
		err := validateID(id)
		if err == nil {
			for _, r := range id {
				if r != '-' && (r < '0' || r > '9') {
					t.Errorf("validateID accepted invalid rune %q in %q", r, id)
				}
			}
		}
		if err != nil && id == "" {
			return // empty is always invalid
		}
		if err != nil && strings.Contains(id, "..") {
			return // path traversal should be rejected
		}
	})
}

func FuzzIsAllowedRestorePath(f *testing.F) {
	f.Add(`C:\Windows\Temp\test.txt`)
	f.Add(`..\..\Windows\System32\cmd.exe`)
	f.Add(`C:\Users\test\..\..\Windows\cmd.exe`)
	f.Add(``)
	f.Add(`\\server\share\file`)
	f.Add(`C:\Windows\System32\cmd.exe\x00hidden`)
	f.Add(`D:\$Recycle.Bin\test`)
	f.Add(strings.Repeat("A", 1000))

	f.Fuzz(func(t *testing.T, path string) {
		result := isAllowedRestorePath(path)

		if result && strings.Contains(filepath.Clean(path), "..") {
			t.Logf("isAllowedRestorePath allowed path with '..' traversal: %q", path)
		}
		if result && !filepath.IsAbs(path) && path != "" {
			t.Logf("isAllowedRestorePath allowed non-absolute path: %q", path)
		}
		if result && strings.Contains(path, "\x00") {
			t.Errorf("isAllowedRestorePath allowed null byte in path: %q", path)
		}
	})
}

func FuzzManifestDecode(f *testing.F) {
	f.Add([]byte(`{"id":"2025-01-01-120000","created_at":"2025-01-01T12:00:00Z","items":[{"original":"C:\\test.txt","quarantined":"C:\\tmp\\test.txt","size":100}]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`{"id":`))
	f.Add([]byte(strings.Repeat("A", 10000)))
	f.Add([]byte(`{"id":"@@@", "items": []}`))
	f.Add([]byte(`{"id":"test", "items": [{"original": "` + strings.Repeat("../", 100) + `file"}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var m types.Manifest
		err := json.Unmarshal(data, &m)
		_ = err // should not panic regardless of input
	})
}
