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
	catScanner{"Delivery Optimization", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Network", "Downloader")
		return scanDir(ctx, path, "delivery_optimization", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Windows Error Reports", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "WER")
		return scanDir(ctx, path, "windows_error_reports", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Messenger Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanMessengerCache(ctx, cfg)
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
	catScanner{"Empty Folders", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanEmptyFolders(ctx, cfg)
	}},
	catScanner{"npm Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanNpmCache(ctx, cfg)
	}},
	catScanner{"pip Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		return scanPipCache(ctx, cfg)
	}},
	catScanner{"Spotify Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("APPDATA"), "Spotify", "Data")
		return scanDir(ctx, path, "spotify_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"OneDrive Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "OneDrive", "cache")
		return scanDir(ctx, path, "onedrive_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Visual Studio Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "VisualStudio")
		return scanDir(ctx, path, "vs_cache", types.RiskSafe, nil, true, cfg)
	}},
	catScanner{"Git Cache", types.RiskSafe, func(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Git", "CredentialManager", "cache")
		return scanDir(ctx, path, "git_cache", types.RiskSafe, nil, true, cfg)
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
	catScanner{"Office Cache", types.RiskSafe, scanOfficeCache},
	catScanner{"Adobe Cache", types.RiskReview, scanAdobeCache},
	catScanner{"Docker Cache", types.RiskReview, scanDockerCache},
	catScanner{"JetBrains Cache", types.RiskSafe, scanJetBrainsCache},
	catScanner{"Go Build Cache", types.RiskSafe, scanGoBuildCache},
	catScanner{"Rust Cache", types.RiskSafe, scanRustCache},
	catScanner{"NuGet Cache", types.RiskReview, scanNuGetCache},
	catScanner{"Unity Cache", types.RiskReview, scanUnityCache},
	catScanner{"Epic Games Cache", types.RiskSafe, scanEpicGamesCache},
	catScanner{"Battle.net Cache", types.RiskSafe, scanBattleNetCache},
	catScanner{"Rockstar Cache", types.RiskSafe, scanRockstarCache},
	catScanner{"EA App Cache", types.RiskSafe, scanEAAppCache},
	catScanner{"Ubisoft Cache", types.RiskSafe, scanUbisoftCache},
	catScanner{"GOG Galaxy Cache", types.RiskSafe, scanGOGGalaxyCache},
	catScanner{"OBS Cache", types.RiskSafe, scanOBSCache},
	catScanner{"Windows Defender", types.RiskReview, scanWindowsDefender},
	catScanner{"TeamViewer Logs", types.RiskSafe, scanTeamViewerLogs},
}
