package scanner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanDevCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// VSCode
	for _, sub := range []string{"Cache", "Code Cache"} {
		path := filepath.Join(os.Getenv("APPDATA"), "Code", sub)
		subItems, err := scanDir(ctx, path, "dev_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: vscode cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// npm
	npmPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "npm-cache")
	subItems, err := scanDir(ctx, npmPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: npm cache scan failed", "path", npmPath, "error", err)
	}
	items = append(items, subItems...)
	// pip
	pipPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "pip", "cache")
	subItems, err = scanDir(ctx, pipPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: pip cache scan failed", "path", pipPath, "error", err)
	}
	items = append(items, subItems...)
	// Git
	gitPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Git", "CredentialManager", "cache")
	subItems, err = scanDir(ctx, gitPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: git cache scan failed", "path", gitPath, "error", err)
	}
	items = append(items, subItems...)
	// Visual Studio
	vsPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "VisualStudio")
	subItems, err = scanDir(ctx, vsPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: visual studio cache scan failed", "path", vsPath, "error", err)
	}
	items = append(items, subItems...)
	// JetBrains
	jbRoot := filepath.Join(os.Getenv("LOCALAPPDATA"), "JetBrains")
	entries, _ := os.ReadDir(jbRoot)
	for _, e := range entries {
		if e.IsDir() {
			for _, sub := range []string{"cache", "caches", "log", "logs"} {
				path := filepath.Join(jbRoot, e.Name(), sub)
				subItems, err := scanDir(ctx, path, "dev_cache", types.RiskSafe, nil, true, cfg)
				if err != nil {
					slog.Warn("scanner: jetbrains dev cache scan failed", "path", path, "error", err)
				}
				items = append(items, subItems...)
			}
		}
	}
	// Go Build
	goPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "go-build")
	subItems, err = scanDir(ctx, goPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: go build cache scan failed", "path", goPath, "error", err)
	}
	items = append(items, subItems...)
	// Rust
	rustPath := filepath.Join(os.Getenv("USERPROFILE"), ".cargo", "registry", "cache")
	subItems, err = scanDir(ctx, rustPath, "dev_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: rust cache scan failed", "path", rustPath, "error", err)
	}
	items = append(items, subItems...)
	return items, nil
}

func scanDevPackageCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Docker
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Docker"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "DockerDesktop"),
		filepath.Join(os.Getenv("APPDATA"), "Docker Desktop"),
		filepath.Join(os.Getenv("USERPROFILE"), ".docker"),
	} {
		subItems, err := scanDir(ctx, path, "dev_package_cache", types.RiskReview, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: docker cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// NuGet
	nugetPath := filepath.Join(os.Getenv("USERPROFILE"), ".nuget", "packages")
	subItems, err := scanDir(ctx, nugetPath, "dev_package_cache", types.RiskReview, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: nuget cache scan failed", "path", nugetPath, "error", err)
	}
	items = append(items, subItems...)
	// Unity
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Unity"),
		filepath.Join(os.Getenv("APPDATA"), "Unity"),
		filepath.Join(os.Getenv("APPDATA"), "UnityHub"),
	} {
		subItems, err := scanDir(ctx, path, "dev_package_cache", types.RiskReview, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: unity cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanAndroidCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Android SDK temp downloads
	sdkTemp := filepath.Join(os.Getenv("LOCALAPPDATA"), "Android", "Sdk", ".temp")
	subItems, err := scanDir(ctx, sdkTemp, "android_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: android sdk temp scan failed", "path", sdkTemp, "error", err)
	}
	items = append(items, subItems...)
	// Android build caches
	for _, path := range []string{
		filepath.Join(os.Getenv("USERPROFILE"), ".android", "cache"),
		filepath.Join(os.Getenv("USERPROFILE"), ".gradle", "caches", "build-cache-1"),
	} {
		subItems, err := scanDir(ctx, path, "android_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: android build cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanPowershellHistory(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	psReadLine := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "PowerShell", "PSReadLine")
	entries, err := os.ReadDir(psReadLine)
	if err != nil {
		return nil, nil
	}
	for _, e := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if e.IsDir() || cfg.IsExcluded(filepath.Join(psReadLine, e.Name())) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			slog.Warn("scanner: ps history stat failed", "name", e.Name(), "error", err)
			continue
		}
		items = append(items, types.Item{
			Category: "powershell_history",
			Path:     filepath.Join(psReadLine, e.Name()),
			Size:     info.Size(),
			Risk:     types.RiskSafe,
		})
	}
	return items, nil
}
