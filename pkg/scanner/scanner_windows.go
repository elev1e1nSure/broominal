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
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanTemp(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	tempPath := os.Getenv("TEMP")
	if tempPath == "" {
		return nil, nil
	}
	return scanDir(ctx, tempPath, "temp", types.RiskSafe, nil, true, cfg)
}

func scanLogs(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	tempPath := os.Getenv("TEMP")
	if tempPath != "" {
		sub, err := scanDir(ctx, tempPath, "logs", types.RiskSafe, []string{".log"}, true, cfg)
		if err != nil {
			slog.Warn("scanner: log scan failed", "path", tempPath, "error", err)
		}
		items = append(items, sub...)
	}
	return items, nil
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
			slog.Warn("scanner: thumbnail stat failed", "path", path, "error", err)
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

func scanWindowsErrorReports(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "WER"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "WER"),
	} {
		sub, err := scanDir(ctx, path, "windows_error_reports", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: WER scan failed", "path", path, "error", err)
		}
		items = append(items, sub...)
	}
	return items, nil
}

func scanDeliveryOptimization(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Network", "Downloader"),
		filepath.Join(os.Getenv("SystemRoot"), "SoftwareDistribution", "DeliveryOptimization"),
	} {
		sub, err := scanDir(ctx, path, "delivery_optimization", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: delivery optimization scan failed", "path", path, "error", err)
		}
		items = append(items, sub...)
	}
	return items, nil
}

func scanWindowsPrefetch(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("SystemRoot"), "Prefetch")
	items, err := scanDir(ctx, path, "windows_prefetch", types.RiskSafe, nil, false, cfg)
	if err != nil {
		slog.Warn("scanner: prefetch scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanIconCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "IconCache.db")
	items, err := scanDir(ctx, path, "icon_cache", types.RiskSafe, nil, false, cfg)
	if err != nil {
		slog.Warn("scanner: icon cache scan failed", "path", path, "error", err)
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

func scanWindowsUpdateCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("SystemRoot"), "SoftwareDistribution", "Download")
	items, err := scanDir(ctx, path, "windows_update_cache", types.RiskReview, nil, false, cfg)
	if err != nil {
		slog.Warn("scanner: windows update cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanWindowsDefender(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("PROGRAMDATA"), "Microsoft", "Windows Defender", "Scans", "History")
	items, err := scanDir(ctx, path, "windows_defender", types.RiskReview, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: defender scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanFontCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	base := filepath.Join(sysRoot, "ServiceProfiles", "LocalService", "AppData", "Local")
	var items []types.Item
	sub, err := scanDir(ctx, filepath.Join(base, "FontCache"), "font_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: font cache dir scan failed", "error", err)
	}
	items = append(items, sub...)
	dat := filepath.Join(base, "FontCache-System.dat")
	if info, err := os.Stat(dat); err == nil && !info.IsDir() {
		items = append(items, types.Item{
			Category: "font_cache",
			Path:     dat,
			Size:     info.Size(),
			Risk:     types.RiskSafe,
		})
	}
	return items, nil
}

func scanWindowsSetupFiles(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("SystemRoot"), "Panther")
	items, err := scanDir(ctx, path, "windows_setup_files", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: windows setup files scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanOldChkdskFiles(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for drive := 'A'; drive <= 'Z'; drive++ {
		root := string(drive) + `:\`
		if _, err := os.Stat(root); err != nil {
			continue
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := strings.ToUpper(e.Name())
			if len(name) == 9 && strings.HasPrefix(name, "FOUND.") {
				sub, err := scanDir(ctx, filepath.Join(root, e.Name()), "old_chkdsk_files", types.RiskReview, nil, true, cfg)
				if err != nil {
					slog.Warn("scanner: chkdsk scan failed", "path", filepath.Join(root, e.Name()), "error", err)
				}
				items = append(items, sub...)
			}
		}
	}
	return items, nil
}

func scanDiagnosticData(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Diagnosis", "ETLLogs")
	items, err := scanDir(ctx, path, "diagnostic_data", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: diagnostic data scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanDownloadedProgramFiles(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	path := filepath.Join(sysRoot, "Downloaded Program Files")
	items, err := scanDir(ctx, path, "downloaded_program_files", types.RiskSafe, nil, false, cfg)
	if err != nil {
		slog.Warn("scanner: downloaded program files scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanFeedbackHubLogs(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Packages", "Microsoft.WindowsFeedbackHub_8wekyb3d8bbwe", "LocalState", "DiagOutputDir")
	items, err := scanDir(ctx, path, "feedback_hub_logs", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: feedback hub logs scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanBranchCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	path := filepath.Join(sysRoot, "ServiceProfiles", "NetworkService", "AppData", "Local", "PeerDistRepub")
	items, err := scanDir(ctx, path, "branch_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: branch cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanRetailDemoContent(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	path := filepath.Join(sysRoot, "ServiceProfiles", "LocalService", "AppData", "Local", "Microsoft", "Windows", "RetailDemo")
	items, err := scanDir(ctx, path, "retail_demo_content", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: retail demo scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanThumbsDb(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return nil, nil
	}
	var items []types.Item
	_ = filepath.WalkDir(userProfile, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || cfg.IsExcluded(path) {
			return nil
		}
		name := strings.ToLower(d.Name())
		if name != "thumbs.db" && name != "ehthumbs.db" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			slog.Warn("scanner: thumbs.db stat failed", "path", path, "error", err)
			return nil
		}
		items = append(items, types.Item{
			Category: "thumbs_db",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskSafe,
		})
		return nil
	})
	return items, nil
}

func scanWindowsOld(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysDrive := os.Getenv("SystemDrive")
	if sysDrive == "" {
		sysDrive = `C:`
	}
	var items []types.Item
	for _, name := range []string{"Windows.old", "$WinREAgent"} {
		path := filepath.Join(sysDrive+`\`, name)
		sub, err := scanDir(ctx, path, "windows_old", types.RiskReview, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: windows.old scan failed", "path", path, "error", err)
		}
		items = append(items, sub...)
	}
	return items, nil
}

func scanCbsDismLogs(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	for _, sub := range []string{"CBS", "DISM"} {
		path := filepath.Join(sysRoot, "Logs", sub)
		subItems, err := scanDir(ctx, path, "cbs_dism_logs", types.RiskReview, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: cbs/dism logs scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanWindowsInstallerPatches(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	path := filepath.Join(sysRoot, "Installer", "$PatchCache$")
	items, err := scanDir(ctx, path, "windows_installer_patches", types.RiskReview, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: installer patches scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanOldInstallersCat(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	downloads := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
	return scanOldInstallers(ctx, downloads, cfg)
}

func scanLargeOldFilesCat(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	profile := os.Getenv("USERPROFILE")
	return scanLargeOldFiles(ctx, profile, cfg)
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
