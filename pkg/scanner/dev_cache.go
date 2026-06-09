package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func yarnCachePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Yarn", "Cache")
}

func pnpmStorePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "pnpm", "store")
}

func cargoRegistryCachePath() string {
	return filepath.Join(os.Getenv("USERPROFILE"), ".cargo", "registry", "cache")
}

func goBuildCachePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "go-build")
}

func goModCachePaths() []string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("USERPROFILE"), "go")
	}
	var paths []string
	for _, root := range strings.Split(gopath, string(os.PathListSeparator)) {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		paths = append(paths, filepath.Join(root, "pkg", "mod"))
	}
	return paths
}

func gradleCachePath() string {
	return filepath.Join(os.Getenv("USERPROFILE"), ".gradle", "caches")
}

func mavenRepositoryPath() string {
	return filepath.Join(os.Getenv("USERPROFILE"), ".m2", "repository")
}

func nuGetCachePaths() []string {
	return []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "NuGet", "v3-cache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "NuGet", "Cache"),
	}
}

func composerCachePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Composer", "cache")
}

func scanYarnCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, yarnCachePath(), "yarn_cache", types.RiskSafe, nil, true, cfg)
}

func scanPnpmStore(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, pnpmStorePath(), "pnpm_store", types.RiskSafe, nil, true, cfg)
}

func scanCargoRegistryCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, cargoRegistryCachePath(), "cargo_registry_cache", types.RiskSafe, nil, true, cfg)
}

func scanGoBuildCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, goBuildCachePath(), "go_build_cache", types.RiskSafe, nil, true, cfg)
}

func scanGoModCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range goModCachePaths() {
		sub, err := scanDir(ctx, path, "go_mod_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			return nil, err
		}
		items = append(items, sub...)
	}
	return items, nil
}

func scanGradleCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, gradleCachePath(), "gradle_cache", types.RiskSafe, nil, true, cfg)
}

// scanMavenSnapshots scans only Maven snapshot artifacts and resolver cache — not release artifacts.
func scanMavenSnapshots(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	root := mavenRepositoryPath()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []types.Item
	var count int

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			base := strings.ToLower(d.Name())
			if base == ".cache" || base == "snapshots" {
				return nil
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		relLower := strings.ToLower(filepath.ToSlash(rel))
		if !strings.Contains(relLower, "/snapshots/") && !strings.HasPrefix(relLower, ".cache/") {
			return nil
		}
		if cfg.IsExcluded(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		count++
		if count > maxScanFiles {
			return errScanLimit
		}
		items = append(items, types.Item{
			Category: "maven_snapshots_cache",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskSafe,
		})
		return nil
	})
	if err != nil && err != errScanLimit {
		return nil, err
	}
	return items, nil
}

func scanNuGetCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanPaths(ctx, nuGetCachePaths(), "nuget_cache", types.RiskSafe, true, cfg)
}

func scanComposerCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanDir(ctx, composerCachePath(), "composer_cache", types.RiskSafe, nil, true, cfg)
}

// Registry aliases
func scanPnpmCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanPnpmStore(ctx, cfg)
}

func scanCargoCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanCargoRegistryCache(ctx, cfg)
}

func scanGoModulesCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	sub, err := scanGoBuildCache(ctx, cfg)
	if err != nil {
		return nil, err
	}
	items = append(items, sub...)
	sub, err = scanGoModCache(ctx, cfg)
	if err != nil {
		return nil, err
	}
	items = append(items, sub...)
	return items, nil
}

func scanMavenCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return scanMavenSnapshots(ctx, cfg)
}
