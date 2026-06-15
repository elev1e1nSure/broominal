package scanner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanMessengerCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Discord
	discordRoot := filepath.Join(os.Getenv("APPDATA"), "discord")
	for _, sub := range []string{"Cache", "Code Cache"} {
		path := filepath.Join(discordRoot, sub)
		subItems, err := scanDir(ctx, path, "messenger_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: discord cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// Telegram Desktop
	telePath := filepath.Join(os.Getenv("APPDATA"), "Telegram Desktop", "tdata", "user_data")
	subItems, err := scanDir(ctx, telePath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: telegram cache scan failed", "path", telePath, "error", err)
	}
	items = append(items, subItems...)
	// Slack
	slackRoot := filepath.Join(os.Getenv("APPDATA"), "Slack")
	for _, sub := range []string{"Cache", "Code Cache", "GPUCache"} {
		path := filepath.Join(slackRoot, sub)
		subItems, err = scanDir(ctx, path, "messenger_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: slack cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// Teams (classic)
	teamsPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Teams", "Cache")
	subItems, err = scanDir(ctx, teamsPath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: teams cache scan failed", "path", teamsPath, "error", err)
	}
	items = append(items, subItems...)
	// New Microsoft Teams (Windows 11)
	newTeamsPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Packages", "MSTeams_8wekyb3d8bbwe", "LocalCache", "Microsoft", "MSTeams")
	subItems, err = scanDir(ctx, newTeamsPath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: new teams cache scan failed", "path", newTeamsPath, "error", err)
	}
	items = append(items, subItems...)
	return items, nil
}

func scanGameLauncherCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Steam
	root := os.Getenv("ProgramFiles(x86)")
	if root == "" {
		root = os.Getenv("ProgramFiles")
	}
	for _, sub := range []string{"appcache", "htmlcache"} {
		path := filepath.Join(root, "Steam", sub)
		subItems, err := scanDir(ctx, path, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: steam cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// Epic Games
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "EpicGamesLauncher", "Saved", "webcache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "EpicGamesLauncher", "Saved", "webcache_4147"),
	} {
		subItems, err := scanDir(ctx, path, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: epic game cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// Battle.net
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Battle.net", "Cache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Battle.net", "BrowserCache"),
	} {
		subItems, err := scanDir(ctx, path, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: battlenet cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// Rockstar
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Rockstar Games", "Launcher", "Cache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Rockstar Games", "Social Club", "Renderer", "Cache"),
	} {
		subItems, err := scanDir(ctx, path, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: rockstar cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// EA App
	eaPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Electronic Arts", "EA Desktop", "Cache")
	subItems, err := scanDir(ctx, eaPath, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: EA cache scan failed", "path", eaPath, "error", err)
	}
	items = append(items, subItems...)
	// Ubisoft
	ubiPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Ubisoft Game Launcher", "cache")
	subItems, err = scanDir(ctx, ubiPath, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: ubisoft cache scan failed", "path", ubiPath, "error", err)
	}
	items = append(items, subItems...)
	// GOG Galaxy
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "GOG.com", "Galaxy", "webcache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "GOG.com", "Galaxy", "webcache_2"),
	} {
		subItems, err := scanDir(ctx, path, "game_launcher_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: GOG cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanServiceCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Spotify
	spotifyPath := filepath.Join(os.Getenv("APPDATA"), "Spotify", "Data")
	subItems, err := scanDir(ctx, spotifyPath, "service_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: spotify cache scan failed", "path", spotifyPath, "error", err)
	}
	items = append(items, subItems...)
	// OneDrive
	onedrivePath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "OneDrive", "cache")
	subItems, err = scanDir(ctx, onedrivePath, "service_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: onedrive cache scan failed", "path", onedrivePath, "error", err)
	}
	items = append(items, subItems...)
	// Office
	officePath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Office", "16.0", "OfficeFileCache")
	subItems, err = scanDir(ctx, officePath, "service_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: office service cache scan failed", "path", officePath, "error", err)
	}
	items = append(items, subItems...)
	// Adobe
	for _, path := range []string{
		filepath.Join(os.Getenv("APPDATA"), "Adobe", "Common", "Media Cache"),
		filepath.Join(os.Getenv("APPDATA"), "Adobe", "Common", "Media Cache Files"),
		filepath.Join(os.Getenv("APPDATA"), "Adobe", "Common", "Peak Files"),
	} {
		subItems, err := scanDir(ctx, path, "service_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: adobe service cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// OBS
	for _, path := range []string{
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "plugin_config"),
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "logs"),
	} {
		subItems, err := scanDir(ctx, path, "service_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: OBS cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	// TeamViewer
	tvPath := filepath.Join(os.Getenv("PROGRAMDATA"), "TeamViewer")
	subItems, err = scanDir(ctx, tvPath, "service_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: teamviewer cache scan failed", "path", tvPath, "error", err)
	}
	items = append(items, subItems...)
	return items, nil
}

func scanZoomCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("APPDATA"), "Zoom", "data"),
		filepath.Join(os.Getenv("APPDATA"), "Zoom", "logs"),
	} {
		subItems, err := scanDir(ctx, path, "zoom_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: zoom cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanEpicGamesCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "EpicGamesLauncher", "Saved")
	for _, sub := range []string{"webcache", "Logs", "crashes"} {
		subItems, err := scanDir(ctx, filepath.Join(base, sub), "epic_games_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: epic games cache scan failed", "path", filepath.Join(base, sub), "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanAdobeCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("APPDATA"), "Adobe", "Common", "Media Cache Files"),
		filepath.Join(os.Getenv("APPDATA"), "Adobe", "Common", "Media Cache"),
	} {
		subItems, err := scanDir(ctx, path, "adobe_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: adobe cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanJetBrainsCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "JetBrains")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, nil
	}
	var items []types.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subItems, err := scanDir(ctx, filepath.Join(base, e.Name(), "caches"), "jetbrains_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: jetbrains cache scan failed", "path", filepath.Join(base, e.Name()), "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanOfficeCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Office")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, nil
	}
	var items []types.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subItems, err := scanDir(ctx, filepath.Join(base, e.Name(), "OfficeFileCache"), "office_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: office cache scan failed", "path", filepath.Join(base, e.Name()), "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanJavaCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("APPDATA"), "LocalLow", "Sun", "Java", "Deployment", "cache")
	items, err := scanDir(ctx, path, "java_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: java cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanOneDriveLogs(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "OneDrive", "logs"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "OneDrive", "setup", "logs"),
		filepath.Join(os.Getenv("USERPROFILE"), "OneDrive", "Logs"),
	} {
		subItems, err := scanDir(ctx, path, "onedrive_logs", types.RiskSafe, []string{".log", ".etl", ".txt"}, true, cfg)
		if err != nil {
			slog.Warn("scanner: onedrive logs scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanRdpCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Terminal Server Client", "Cache")
	items, err := scanDir(ctx, path, "rdp_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: rdp cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanPrinterSpooler(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("SystemRoot"), "System32", "spool", "PRINTERS")
	cutoff := time.Now().Add(-1 * time.Hour)
	items, err := scanDirWithAge(ctx, path, "printer_spooler", types.RiskReview, cutoff, cfg)
	if err != nil {
		slog.Warn("scanner: printer spooler scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanMsStoreCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	packagesRoot := filepath.Join(os.Getenv("LOCALAPPDATA"), "Packages")
	entries, err := os.ReadDir(packagesRoot)
	if err != nil {
		return nil, nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		for _, sub := range []string{"AC", "LocalCache", "TempState"} {
			path := filepath.Join(packagesRoot, e.Name(), sub)
			subItems, err := scanDir(ctx, path, "ms_store_cache", types.RiskReview, nil, true, cfg)
			if err != nil {
				slog.Warn("scanner: ms store cache scan failed", "path", path, "error", err)
			}
			items = append(items, subItems...)
		}
	}
	return items, nil
}

func scanPostmanCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("APPDATA"), "Postman", "Cache"),
		filepath.Join(os.Getenv("APPDATA"), "Postman", "Code Cache"),
		filepath.Join(os.Getenv("APPDATA"), "Postman", "GPUCache"),
		filepath.Join(os.Getenv("APPDATA"), "Postman", "blob_storage"),
		filepath.Join(os.Getenv("APPDATA"), "Postman", "DawnGraphiteCache"),
		filepath.Join(os.Getenv("APPDATA"), "Postman", "DawnWebGPUCache"),
	} {
		subItems, err := scanDir(ctx, path, "postman_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: postman cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}
