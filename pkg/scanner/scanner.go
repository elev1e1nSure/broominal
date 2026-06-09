package scanner

import (
	"fmt"
	"io/fs"
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

func Scan() (*types.ScanResult, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	return ScanWithConfig(cfg)
}

func ScanWithConfig(cfg *config.Config) (*types.ScanResult, error) {
	result := &types.ScanResult{}
	categories := make(map[string]*types.CategorySummary)

	// Temp
	if cfg.IsCategoryEnabled("Temp") {
		tempPath := os.Getenv("TEMP")
		if tempPath != "" {
			items, err := scanDir(tempPath, "temp", types.RiskSafe, nil, true, cfg)
			if err == nil {
				mergeItems(categories, "Temp", types.RiskSafe, items)
			}
		}
	}

	// Downloads
	if cfg.IsCategoryEnabled("Downloads") {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		items, err := scanDir(downloadsPath, "downloads", types.RiskReview, nil, true, cfg)
		if err == nil {
			mergeItems(categories, "Downloads", types.RiskReview, items)
		}
	}

	// Browser Cache
	if cfg.IsCategoryEnabled("Browser Cache") {
		for _, rel := range browserCachePaths {
			path := filepath.Join(os.Getenv("USERPROFILE"), rel)
			if strings.Contains(rel, "Firefox") {
				items, _ := scanFirefoxCache(path, cfg)
				mergeItems(categories, "Browser Cache", types.RiskSafe, items)
			} else {
				items, _ := scanDir(path, "browser_cache", types.RiskSafe, nil, true, cfg)
				mergeItems(categories, "Browser Cache", types.RiskSafe, items)
			}
		}
	}

	// Recycle Bin
	if cfg.IsCategoryEnabled("Recycle Bin") {
		recyclePaths := recycleBinPaths()
		for _, rp := range recyclePaths {
			items, _ := scanDir(rp, "recycle_bin", types.RiskSafe, nil, true, cfg)
			mergeItems(categories, "Recycle Bin", types.RiskSafe, items)
		}
	}

	// Logs
	if cfg.IsCategoryEnabled("Logs") {
		logItems := scanLogs(cfg)
		mergeItems(categories, "Logs", types.RiskSafe, logItems)
	}

	// Old Installers
	if cfg.IsCategoryEnabled("Old Installers") {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		installerItems, _ := scanOldInstallers(downloadsPath, cfg)
		mergeItems(categories, "Old Installers", types.RiskReview, installerItems)
	}

	// Large Old Files
	if cfg.IsCategoryEnabled("Large Old Files") {
		downloadsPath := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
		largeItems, _ := scanLargeOldFiles(downloadsPath, cfg)
		mergeItems(categories, "Large Old Files", types.RiskReview, largeItems)
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

func isExcluded(path string, cfg *config.Config) bool {
	for _, ex := range cfg.Exclusions {
		if strings.Contains(strings.ToLower(path), strings.ToLower(ex)) {
			return true
		}
	}
	return false
}

func scanDir(root, category string, risk types.RiskLevel, matchExt []string, recursive bool, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
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
		if isExcluded(path, cfg) {
			return nil
		}

		if len(matchExt) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			found := false
			for _, me := range matchExt {
				if ext == me {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		items = append(items, types.Item{
			Category: category,
			Path:     path,
			Size:     info.Size(),
			Risk:     risk,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func scanFirefoxCache(root string, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if filepath.Base(path) == "cache2" {
				// scan cache2
				sub, _ := scanDir(path, "browser_cache", types.RiskSafe, nil, true, cfg)
				items = append(items, sub...)
				return filepath.SkipDir
			}
			return nil
		}
		return nil
	})
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

func scanLogs(cfg *config.Config) []types.Item {
	var items []types.Item
	tempPath := os.Getenv("TEMP")
	if tempPath != "" {
		sub, _ := scanDir(tempPath, "logs", types.RiskSafe, []string{".log"}, true, cfg)
		items = append(items, sub...)
	}
	// Windows Event Logs (safe to list, but we can't delete them easily)
	// Skip system event logs for safety
	return items
}

func scanOldInstallers(root string, cfg *config.Config) ([]types.Item, error) {
	months := cfg.OldInstallerMonths
	if months <= 0 {
		months = 6
	}
	cutoff := time.Now().AddDate(0, -months, 0)
	var items []types.Item

	match := func(path string, d fs.DirEntry) bool {
		if d.IsDir() {
			return false
		}
		if isExcluded(path, cfg) {
			return false
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".msi" && ext != ".exe" {
			return false
		}
		info, err := d.Info()
		if err != nil {
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

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if match(path, d) {
			info, _ := d.Info()
			items = append(items, types.Item{
				Category: "old_installers",
				Path:     path,
				Size:     info.Size(),
				Risk:     types.RiskReview,
			})
		}
		return nil
	})
	return items, err
}

func scanLargeOldFiles(root string, cfg *config.Config) ([]types.Item, error) {
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

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if isExcluded(path, cfg) {
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
		items = append(items, types.Item{
			Category: "large_old_files",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskReview,
		})
		return nil
	})
	return items, err
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

// FormatSize форматирует байты в человекочитаемый вид
func FormatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
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
