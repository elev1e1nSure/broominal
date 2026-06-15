package scanner

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanBrowserCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, rel := range browserCachePaths {
		path := filepath.Join(os.Getenv("USERPROFILE"), rel)
		if strings.Contains(rel, "Firefox") {
			sub, err := scanFirefoxCache(ctx, path, cfg)
			if err != nil {
				slog.Warn("scanner: firefox cache scan failed", "path", path, "error", err)
			}
			items = append(items, sub...)
		} else {
			sub, err := scanDir(ctx, path, "browser_cache", types.RiskSafe, nil, true, cfg)
			if err != nil {
				slog.Warn("scanner: browser cache scan failed", "path", path, "error", err)
			}
			items = append(items, sub...)
		}
	}
	return items, nil
}

func scanFirefoxCache2(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	root := filepath.Join(os.Getenv("LOCALAPPDATA"), "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var items []types.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(root, e.Name(), "cache2")
		subItems, err := scanDir(ctx, path, "firefox_cache2", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: firefox cache2 scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanEdgeCodeCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data", "Default", "Code Cache")
	items, err := scanDir(ctx, path, "edge_code_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: edge code cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanChromeCodeCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default", "Code Cache")
	items, err := scanDir(ctx, path, "chrome_code_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: chrome code cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanOperaCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Opera Software", "Opera Stable", "Cache")
	items, err := scanDir(ctx, path, "opera_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: opera cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanBraveCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "BraveSoftware", "Brave-Browser", "User Data", "Default", "Cache")
	items, err := scanDir(ctx, path, "brave_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: brave cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanVivaldiCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Vivaldi", "User Data", "Default", "Cache")
	items, err := scanDir(ctx, path, "vivaldi_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: vivaldi cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanYandexCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Yandex", "YandexBrowser", "User Data", "Default", "Cache")
	items, err := scanDir(ctx, path, "yandex_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: yandex cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanEdgeWebViewCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "EdgeWebView", "User", "Default")
	for _, sub := range []string{"Cache", "Code Cache", "GPUCache"} {
		subItems, err := scanDir(ctx, filepath.Join(base, sub), "edge_webview_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: edge webview cache scan failed", "path", filepath.Join(base, sub), "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}
