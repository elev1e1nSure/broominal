package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/config"
)

func TestDevCachePaths(t *testing.T) {
	t.Setenv("LOCALAPPDATA", `C:\Users\me\AppData\Local`)
	t.Setenv("USERPROFILE", `C:\Users\me`)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"yarn", yarnCachePath(), `C:\Users\me\AppData\Local\Yarn\Cache`},
		{"pnpm", pnpmStorePath(), `C:\Users\me\AppData\Local\pnpm\store`},
		{"cargo", cargoRegistryCachePath(), `C:\Users\me\.cargo\registry\cache`},
		{"go-build", goBuildCachePath(), `C:\Users\me\AppData\Local\go-build`},
		{"gradle", gradleCachePath(), `C:\Users\me\.gradle\caches`},
		{"maven", mavenRepositoryPath(), `C:\Users\me\.m2\repository`},
		{"nuget-v3", nuGetCachePaths()[0], `C:\Users\me\AppData\Local\NuGet\v3-cache`},
		{"composer", composerCachePath(), `C:\Users\me\AppData\Local\Composer\cache`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s path = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestGoModCachePaths(t *testing.T) {
	t.Setenv("USERPROFILE", `C:\Users\me`)
	t.Setenv("GOPATH", "")

	paths := goModCachePaths()
	if len(paths) != 1 || paths[0] != `C:\Users\me\go\pkg\mod` {
		t.Fatalf("default GOPATH paths = %v, want [C:\\Users\\me\\go\\pkg\\mod]", paths)
	}

	t.Setenv("GOPATH", `D:\go;E:\altgo`)
	paths = goModCachePaths()
	want := []string{`D:\go\pkg\mod`, `E:\altgo\pkg\mod`}
	if len(paths) != len(want) {
		t.Fatalf("paths len = %d, want %d", len(paths), len(want))
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}

func TestScanMavenSnapshots(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp)
	repo := mavenRepositoryPath()
	snapshotFile := filepath.Join(repo, "com", "example", "app", "1.0-SNAPSHOT", "snapshots", "app-1.0.jar")
	cacheFile := filepath.Join(repo, ".cache", "resolver-status.properties")
	releaseFile := filepath.Join(repo, "com", "example", "app", "1.0", "app-1.0.jar")

	for _, p := range []string{snapshotFile, cacheFile, releaseFile} {
		_ = os.MkdirAll(filepath.Dir(p), 0755)
		_ = os.WriteFile(p, []byte("data"), 0644)
	}

	cfg := config.Default()
	items, err := scanMavenSnapshots(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scanMavenSnapshots error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (snapshot + cache), got %d", len(items))
	}
	for _, it := range items {
		if it.Path == releaseFile {
			t.Errorf("release artifact should not be scanned: %s", it.Path)
		}
	}
}

func TestScanDevCacheDirs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("USERPROFILE", tmp)

	cacheRoot := filepath.Join(tmp, "Yarn", "Cache")
	_ = os.MkdirAll(cacheRoot, 0755)
	_ = os.WriteFile(filepath.Join(cacheRoot, "entry"), []byte("yarn"), 0644)

	cfg := config.Default()
	items, err := scanYarnCache(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scanYarnCache error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 yarn cache item, got %d", len(items))
	}
}
