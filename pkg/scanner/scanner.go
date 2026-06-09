package scanner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

var browserCachePaths = []string{
	`AppData\Local\Google\Chrome\User Data\Default\Cache`,
	`AppData\Local\Google\Chrome\User Data\Default\Code Cache`,
	`AppData\Local\Microsoft\Edge\User Data\Default\Cache`,
	`AppData\Local\Microsoft\Edge\User Data\Default\Code Cache`,
	`AppData\Local\Mozilla\Firefox\Profiles`,
}

var logPatterns = []string{
	`*.log`,
	`AppData\Local\Temp\*.log`,
}

const maxScanFiles = 50000

func ScanWithConfig(ctx context.Context, cfg *config.Config) (*types.ScanResult, error) {
	result := &types.ScanResult{}
	categories := make(map[string]*types.CategorySummary)

	for _, sc := range allScanners {
		if !cfg.IsCategoryEnabled(sc.Name()) {
			continue
		}
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		items, err := sc.Scan(ctx, cfg)
		if err != nil {
			continue
		}
		mergeItems(categories, sc.Name(), sc.Risk(), items)
	}

	for _, cat := range categories {
		result.Categories = append(result.Categories, *cat)
		result.TotalSize += cat.Size
		switch cat.Risk {
		case types.RiskSafe:
			result.SafeSize += cat.Size
		case types.RiskReview:
			result.ReviewSize += cat.Size
		case types.RiskDanger:
			result.DangerSize += cat.Size
		}
	}

	return result, nil
}

var errScanLimit = errors.New("scan file limit reached")

func scanDir(ctx context.Context, root, category string, risk types.RiskLevel, matchExt []string, recursive bool, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	var count int

	var extSet map[string]struct{}
	if len(matchExt) > 0 {
		extSet = make(map[string]struct{}, len(matchExt))
		for _, e := range matchExt {
			extSet[e] = struct{}{}
		}
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return filepath.SkipDir
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("scan %s: walk error at %s: %w", category, path, err)
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			if !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if cfg.IsExcluded(path) {
			return nil
		}

		if extSet != nil {
			if _, ok := extSet[strings.ToLower(filepath.Ext(path))]; !ok {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		count++
		if count > maxScanFiles {
			slog.Info("scan: max file limit reached, truncating", "category", category, "limit", maxScanFiles)
			return errScanLimit
		}

		items = append(items, types.Item{
			Category: category,
			Path:     path,
			Size:     info.Size(),
			Risk:     risk,
		})
		return nil
	})
	if err != nil && !errors.Is(err, errScanLimit) {
		return nil, err
	}
	return items, nil
}

func scanFirefoxCache(ctx context.Context, root string, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return filepath.SkipDir
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("scan browser_cache: walk error at %s: %w", path, err)
		}
		if d.IsDir() {
			if filepath.Base(path) == "cache2" {
				sub, err := scanDir(ctx, path, "browser_cache", types.RiskSafe, nil, true, cfg)
				if err == nil {
					items = append(items, sub...)
				}
				return filepath.SkipDir
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func recycleBinPaths() []string {
	drive := os.Getenv("SYSTEMDRIVE")
	if drive == "" {
		drive = "C:"
	}
	binPath := filepath.Join(drive, "$Recycle.Bin")
	entries, err := os.ReadDir(binPath)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) > 2 {
			paths = append(paths, filepath.Join(binPath, e.Name()))
		}
	}
	return paths
}

func scanLogs(ctx context.Context, cfg *config.Config) []types.Item {
	var items []types.Item
	tempPath := os.Getenv("TEMP")
	if tempPath != "" {
		sub, _ := scanDir(ctx, tempPath, "logs", types.RiskSafe, []string{".log"}, true, cfg)
		items = append(items, sub...)
	}
	// Windows Event Logs (safe to list, but we can't delete them easily)
	// Skip system event logs for safety
	return items
}

func scanOldInstallers(ctx context.Context, root string, cfg *config.Config) ([]types.Item, error) {
	months := cfg.OldInstallerMonths
	if months <= 0 {
		months = 6
	}
	cutoff := time.Now().AddDate(0, -months, 0)
	var items []types.Item
	var count int

	match := func(path string, d fs.DirEntry, info fs.FileInfo) bool {
		if d.IsDir() {
			return false
		}
		if cfg.IsExcluded(path) {
			return false
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".msi" && ext != ".exe" {
			return false
		}
		if info.ModTime().After(cutoff) {
			return false
		}
		// Skip if in system-ish paths
		lp := strings.ToLower(path)
		if strings.Contains(lp, "system32") || strings.Contains(lp, "syswow64") || strings.Contains(lp, "windows") {
			return false
		}
		return true
	}

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return filepath.SkipDir
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("scan old_installers: walk error at %s: %w", path, err)
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if match(path, d, info) {
			count++
			if count > maxScanFiles {
				slog.Info("scan: max file limit reached, truncating", "category", "old_installers", "limit", maxScanFiles)
				return errScanLimit
			}
			items = append(items, types.Item{
				Category: "old_installers",
				Path:     path,
				Size:     info.Size(),
				Risk:     types.RiskReview,
			})
		}
		return nil
	})
	return items, nil
}

func scanLargeOldFiles(ctx context.Context, root string, cfg *config.Config) ([]types.Item, error) {
	months := cfg.LargeFileMonths
	if months <= 0 {
		months = 6
	}
	cutoff := time.Now().AddDate(0, -months, 0)
	minSizeMB := cfg.LargeFileMinSizeMB
	if minSizeMB <= 0 {
		minSizeMB = 100
	}
	minSize := int64(minSizeMB) * 1024 * 1024
	var items []types.Item
	var count int

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return filepath.SkipDir
			}
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("scan large_old_files: walk error at %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		if cfg.IsExcluded(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() < minSize {
			return nil
		}
		if info.ModTime().After(cutoff) {
			return nil
		}
		lp := strings.ToLower(path)
		if strings.Contains(lp, "system32") || strings.Contains(lp, "syswow64") || strings.Contains(lp, "windows") {
			return nil
		}
		count++
		if count > maxScanFiles {
			slog.Info("scan: max file limit reached, truncating", "category", "large_old_files", "limit", maxScanFiles)
			return errScanLimit
		}
		items = append(items, types.Item{
			Category: "large_old_files",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskReview,
		})
		return nil
	})
	return items, nil
}

func mergeItems(cats map[string]*types.CategorySummary, name string, risk types.RiskLevel, items []types.Item) {
	if len(items) == 0 {
		return
	}
	cat, ok := cats[name]
	if !ok {
		cat = &types.CategorySummary{
			Category: name,
			Risk:     risk,
		}
		cats[name] = cat
	}
	for _, it := range items {
		cat.Size += it.Size
		cat.Files++
		cat.Items = append(cat.Items, it)
	}
}

// IsSystemFile проверяет, не является ли файл системным или скрытым
func IsSystemFile(path string) bool {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return false
	}
	return attrs&syscall.FILE_ATTRIBUTE_SYSTEM != 0
}

func scanThumbnails(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	root := filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Explorer")
	matches, err := filepath.Glob(filepath.Join(root, "thumbcache_*.db"))
	if err != nil {
		return nil, err
	}
	var items []types.Item
	for _, path := range matches {
		if cfg.IsExcluded(path) {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		items = append(items, types.Item{
			Category: "thumbnails_cache",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskSafe,
		})
	}
	return items, nil
}

func scanMessengerCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	// Discord
	discordRoot := filepath.Join(os.Getenv("APPDATA"), "discord")
	for _, sub := range []string{"Cache", "Code Cache"} {
		path := filepath.Join(discordRoot, sub)
		subItems, _ := scanDir(ctx, path, "messenger_cache", types.RiskSafe, nil, true, cfg)
		items = append(items, subItems...)
	}
	// Telegram Desktop
	telePath := filepath.Join(os.Getenv("APPDATA"), "Telegram Desktop", "tdata", "user_data")
	subItems, _ := scanDir(ctx, telePath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	items = append(items, subItems...)
	// Slack
	slackPath := filepath.Join(os.Getenv("APPDATA"), "Slack", "storage", "slack-settings")
	subItems, _ = scanDir(ctx, slackPath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	items = append(items, subItems...)
	// Teams
	teamsPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Teams", "Cache")
	subItems, _ = scanDir(ctx, teamsPath, "messenger_cache", types.RiskSafe, nil, true, cfg)
	items = append(items, subItems...)
	return items, nil
}

func scanSteamCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	root := os.Getenv("ProgramFiles(x86)")
	if root == "" {
		root = os.Getenv("ProgramFiles")
	}
	root = filepath.Join(root, "Steam")
	var items []types.Item
	for _, sub := range []string{"appcache", "htmlcache"} {
		path := filepath.Join(root, sub)
		subItems, err := scanDir(ctx, path, "steam_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanCrashMemoryDumps(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	paths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "CrashDumps"),
		filepath.Join(os.Getenv("SystemRoot"), "Minidump"),
	}
	for _, path := range paths {
		subItems, err := scanDir(ctx, path, "crash_dumps", types.RiskReview, nil, true, cfg)
		if err != nil {
		}
		items = append(items, subItems...)
	}
	// MEMORY.DMP
	memDmp := filepath.Join(os.Getenv("SystemRoot"), "MEMORY.DMP")
	if info, err := os.Stat(memDmp); err == nil && !info.IsDir() {
		items = append(items, types.Item{
			Category: "memory_dumps",
			Path:     memDmp,
			Size:     info.Size(),
			Risk:     types.RiskReview,
		})
	}
	return items, nil
}

func scanNvidiaInstallerLeftovers(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	paths := []string{
		`C:\NVIDIA\DisplayDriver`,
		filepath.Join(os.Getenv("ProgramData"), "NVIDIA Corporation", "Downloader"),
	}
	for _, path := range paths {
		subItems, err := scanDir(ctx, path, "nvidia_installer_leftovers", types.RiskReview, nil, true, cfg)
		if err != nil {
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanVSCodeCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	root := filepath.Join(os.Getenv("APPDATA"), "Code")
	for _, sub := range []string{"Cache", "Code Cache"} {
		path := filepath.Join(root, sub)
		subItems, err := scanDir(ctx, path, "vscode_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
		}
		items = append(items, subItems...)
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
		subItems, _ := scanDir(ctx, path, "firefox_cache2", types.RiskSafe, nil, true, cfg)
		items = append(items, subItems...)
	}
	return items, nil
}

func scanOldTempFiles(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	days := cfg.OldTempDays
	if days <= 0 {
		days = 7
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var items []types.Item
	var count int
	paths := []string{
		os.Getenv("TEMP"),
		filepath.Join(os.Getenv("WINDIR"), "Temp"),
	}
	for _, root := range paths {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					return filepath.SkipDir
				}
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}
				return fmt.Errorf("scan old_temp_files: walk error at %s: %w", path, err)
			}
			if d.IsDir() {
				return nil
			}
			if cfg.IsExcluded(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			if info.ModTime().Before(cutoff) {
				count++
				if count > maxScanFiles {
					slog.Info("scan: max file limit reached, truncating", "category", "old_temp_files", "limit", maxScanFiles)
					return errScanLimit
				}
				items = append(items, types.Item{
					Category: "old_temp_files",
					Path:     path,
					Size:     info.Size(),
					Risk:     types.RiskSafe,
				})
			}
			return nil
		})
	}
	return items, nil
}

func scanOldExtensions(ctx context.Context, ext string, cfg *config.Config) ([]types.Item, error) {
	days := cfg.OldExtensionDays
	if days <= 0 {
		days = 30
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var items []types.Item
	var count int
	paths := []string{
		os.Getenv("TEMP"),
		filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
		os.Getenv("USERPROFILE"),
	}
	for _, root := range paths {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					return filepath.SkipDir
				}
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}
				return fmt.Errorf("scan old_extension_files: walk error at %s: %w", path, err)
			}
			if d.IsDir() {
				return nil
			}
			if cfg.IsExcluded(path) {
				return nil
			}
			if !strings.EqualFold(filepath.Ext(path), ext) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			if info.ModTime().Before(cutoff) {
				count++
				if count > maxScanFiles {
					slog.Info("scan: max file limit reached, truncating", "category", "old_extension_files", "limit", maxScanFiles)
					return errScanLimit
				}
				items = append(items, types.Item{
					Category: "old_" + strings.TrimPrefix(ext, ".") + "_files",
					Path:     path,
					Size:     info.Size(),
					Risk:     types.RiskReview,
				})
			}
			return nil
		})
	}
	return items, nil
}

func scanEmptyFolders(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	paths := []string{
		os.Getenv("TEMP"),
		filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
	}
	for _, root := range paths {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					return filepath.SkipDir
				}
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}
				return fmt.Errorf("scan empty_folders: walk error at %s: %w", path, err)
			}
			if !d.IsDir() || path == root {
				return nil
			}
			if cfg.IsExcluded(path) {
				return nil
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}
			if len(entries) == 0 {
				items = append(items, types.Item{
					Category: "empty_folders",
					Path:     path,
					Size:     0,
					Risk:     types.RiskSafe,
				})
			}
			return nil
		})
	}
	return items, nil
}

func scanNpmCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "npm-cache")
	return scanDir(ctx, path, "npm_cache", types.RiskSafe, nil, true, cfg)
}

func scanPipCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "pip", "cache")
	return scanDir(ctx, path, "pip_cache", types.RiskSafe, nil, true, cfg)
}
