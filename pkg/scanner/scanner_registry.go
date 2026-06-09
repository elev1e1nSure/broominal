package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// CategoryScanner scans a single cleanup category.
type CategoryScanner interface {
	Name() string
	Risk() types.RiskLevel
	Scan(ctx context.Context, cfg *config.Config) ([]types.Item, error)
}

type catScanner struct {
	name string
	risk types.RiskLevel
	scan func(context.Context, *config.Config) ([]types.Item, error)
}

func (c catScanner) Name() string          { return c.name }
func (c catScanner) Risk() types.RiskLevel { return c.risk }
func (c catScanner) Scan(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return c.scan(ctx, cfg)
}

// allScanners registers every supported cleanup category.
var allScanners = []CategoryScanner{
	catScanner{"Temp", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		tempPath := os.Getenv("TEMP")
		if tempPath == "" {
			return nil, nil
		}
		return scanDir(ctx, tempPath, "temp", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Downloads", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		return scanDir(ctx, downloadsPath, "downloads", types.RiskReview, nil, true, cfg)
	}},
	catScanner{"Browser Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		var items []types.Item
		for _, rel := range browserCachePaths {
			path := filepath.Join(os.Getenv("USERPROFILE"), rel)
			if strings.Contains(rel, "Firefox") {
				sub, _ := scanFirefoxCache(ctx, path, cfg)
				items = append(items, sub...)
			} else {
				sub, _ := scanDir(ctx, path, "browser_cache", types.RiskSafe, nil, true, cfg)
				items = append(items, sub...)
			}
		}
		return items, nil
	}},
	catScanner{"Recycle Bin", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		var items []types.Item
		for _, rp := range recycleBinPaths() {
			sub, _ := scanDir(ctx, rp, "recycle_bin", types.RiskSafe, nil, true, cfg)
			items = append(items, sub...)
		}
		return items, nil
	}},
	catScanner{"Logs", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanLogs(ctx, cfg), nil
	}},
	catScanner{"Old Installers", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		return scanOldInstallers(ctx, downloadsPath, cfg)
	}},
	catScanner{"Large Old Files", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		return scanLargeOldFiles(ctx, downloadsPath, cfg)
	}},
	catScanner{"Thumbnails Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanThumbnails(ctx, cfg)
	}},
	catScanner{"DirectX Shader Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "D3DSCache")
		return scanDir(ctx, path, "directx_shader_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Delivery Optimization", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Network", "Downloader")
		return scanDir(ctx, path, "delivery_optimization", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Windows Error Reports", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "WER")
		return scanDir(ctx, path, "windows_error_reports", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Discord Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanDiscordCache(ctx, cfg)
	}},
	catScanner{"Steam Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanSteamCache(ctx, cfg)
	}},
	catScanner{"Windows Update Cache", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("SystemRoot"), "SoftwareDistribution", "Download")
		return scanDir(ctx, path, "windows_update_cache", types.RiskReview, nil, false, cfg)
	}},
	catScanner{"Crash & Memory Dumps", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanCrashMemoryDumps(ctx, cfg)
	}},
	catScanner{"Nvidia Installer Leftovers", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanNvidiaInstallerLeftovers(ctx, cfg)
	}},
	catScanner{"Telegram Desktop Cache", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("APPDATA"), "Telegram Desktop", "tdata", "user_data")
		return scanDir(ctx, path, "telegram_desktop_cache", types.RiskReview, nil, true, cfg)
	}},
	catScanner{"VSCode Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanVSCodeCache(ctx, cfg)
	}},
	catScanner{"Edge Code Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data", "Default", "Code Cache")
		return scanDir(ctx, path, "edge_code_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Chrome Code Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default", "Code Cache")
		return scanDir(ctx, path, "chrome_code_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Firefox Cache2", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanFirefoxCache2(ctx, cfg)
	}},
	catScanner{"Old Temp Files", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanOldTempFiles(ctx, cfg)
	}},
	catScanner{"Old .tmp Files", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanOldExtensions(ctx, ".tmp", cfg)
	}},
	catScanner{"Old .log Files", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanOldExtensions(ctx, ".log", cfg)
	}},
	catScanner{"Old .bak Files", types.RiskReview, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanOldExtensions(ctx, ".bak", cfg)
	}},
	catScanner{"Empty Folders", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanEmptyFolders(ctx, cfg)
	}},
	catScanner{"npm Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanNpmCache(ctx, cfg)
	}},
	catScanner{"pip Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanPipCache(ctx, cfg)
	}},
}
