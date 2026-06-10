package scanner

import (
	"context"
	"log/slog"
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
	}},
	catScanner{"Recycle Bin", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		var items []types.Item
		for _, rp := range recycleBinPaths() {
			sub, err := scanDir(ctx, rp, "recycle_bin", types.RiskSafe, nil, true, cfg)
			if err != nil {
				slog.Warn("scanner: recycle bin scan failed", "path", rp, "error", err)
			}
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
	catScanner{"Delivery Optimization", types.RiskSafe, scanDeliveryOptimization},
	catScanner{"Windows Error Reports", types.RiskSafe, scanWindowsErrorReports},
	catScanner{"Messenger Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanMessengerCache(ctx, cfg)
	}},
	catScanner{"Game Launcher Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanGameLauncherCache(ctx, cfg)
	}},
	catScanner{"Service Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanServiceCache(ctx, cfg)
	}},
	catScanner{"Dev Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanDevCache(ctx, cfg)
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
	catScanner{"Empty Folders", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanEmptyFolders(ctx, cfg)
	}},
	catScanner{"Windows Prefetch", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("SystemRoot"), "Prefetch")
		return scanDir(ctx, path, "windows_prefetch", types.RiskSafe, nil, false, cfg)
	}},
	catScanner{"Icon Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "IconCache.db")
		items, err := scanDir(ctx, path, "icon_cache", types.RiskSafe, nil, false, cfg)
		if err != nil {
			slog.Warn("scanner: icon cache scan failed", "path", path, "error", err)
		}
		return items, nil
	}},
	catScanner{"Opera Cache", types.RiskSafe, scanOperaCache},
	catScanner{"Brave Cache", types.RiskSafe, scanBraveCache},
	catScanner{"Vivaldi Cache", types.RiskSafe, scanVivaldiCache},
	catScanner{"Yandex Cache", types.RiskSafe, scanYandexCache},
	catScanner{"Windows Defender", types.RiskReview, scanWindowsDefender},
	catScanner{"AMD GPU Cache", types.RiskSafe, scanAMDGPUCache},
	catScanner{"Zoom Cache", types.RiskSafe, scanZoomCache},
	catScanner{"Startup Leftovers", types.RiskReview, scanStartupLeftovers},
	catScanner{"Scheduled Tasks", types.RiskReview, scanScheduledTasksLeftovers},
	catScanner{"Duplicate Files", types.RiskReview, scanDuplicateFiles},
	catScanner{"Edge WebView2 Cache", types.RiskSafe, scanEdgeWebViewCache},
	catScanner{"Epic Games Cache", types.RiskSafe, scanEpicGamesCache},
	catScanner{"Adobe Cache", types.RiskSafe, scanAdobeCache},
	catScanner{"JetBrains Cache", types.RiskSafe, scanJetBrainsCache},
	catScanner{"Office Cache", types.RiskSafe, scanOfficeCache},
	catScanner{"Java Cache", types.RiskSafe, scanJavaCache},
	catScanner{"Recent Documents", types.RiskReview, scanRecentDocuments},
	catScanner{"Font Cache", types.RiskSafe, scanFontCache},
	catScanner{"Windows Setup Files", types.RiskSafe, scanWindowsSetupFiles},
	catScanner{"Old Chkdsk Files", types.RiskReview, scanOldChkdskFiles},
	catScanner{"Diagnostic Data", types.RiskSafe, scanDiagnosticData},
	catScanner{"Downloaded Program Files", types.RiskSafe, scanDownloadedProgramFiles},
	catScanner{"Feedback Hub Logs", types.RiskSafe, scanFeedbackHubLogs},
	catScanner{"BranchCache", types.RiskSafe, scanBranchCache},
	catScanner{"RetailDemo Content", types.RiskSafe, scanRetailDemoContent},
	catScanner{"Thumbs.db", types.RiskSafe, scanThumbsDb},
	catScanner{"Windows.old", types.RiskReview, scanWindowsOld},
}
