package scanner

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

const maxScanFiles = 50000

const scanParallelism = 8

var errScanLimit = errors.New("scan file limit reached")

// EnabledScannerCount returns how many scanners are enabled in the given config.
func EnabledScannerCount(cfg *config.Config) int {
	n := 0
	for _, sc := range allScanners {
		if cfg.IsCategoryEnabled(sc.Name()) {
			n++
		}
	}
	return n
}

func ScanWithConfig(ctx context.Context, cfg *config.Config, progress func(done int)) (*types.ScanResult, error) {
	var enabled []CategoryScanner
	for _, sc := range allScanners {
		if cfg.IsCategoryEnabled(sc.Name()) {
			enabled = append(enabled, sc)
		}
	}

	categories := make(map[string]*types.CategorySummary)
	var mu sync.Mutex
	var done int

	sem := make(chan struct{}, scanParallelism)
	var wg sync.WaitGroup

	for _, sc := range enabled {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sc := sc
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			items, err := sc.Scan(ctx, cfg)
			mu.Lock()
			done++
			if err == nil {
				mergeItems(categories, sc.Name(), sc.Risk(), items)
			} else {
				slog.Warn("scanner: category scan failed", "category", sc.Name(), "error", err)
			}
			if progress != nil {
				progress(done)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	result := &types.ScanResult{}
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

// scanDir is the generic helper for walking a directory and collecting file Items.
// It handles permission errors, exclusions, extension filters, and file limits.
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

func hashFileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// scanFirefoxCache walks Firefox profile dirs looking for cache2 subdirs.
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
				if err != nil {
					slog.Warn("scanner: firefox cache2 scan failed", "path", path, "error", err)
				}
				items = append(items, sub...)
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

// extractTaskCommand parses the <Command> element from a Windows scheduled task XML file.
// Handles UTF-16LE encoded files.
func extractTaskCommand(data []byte) string {
	s := string(data)
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		var b strings.Builder
		for i := 2; i+1 < len(data); i += 2 {
			b.WriteRune(rune(uint16(data[i]) | uint16(data[i+1])<<8))
		}
		s = b.String()
	}
	lower := strings.ToLower(s)
	const open, close = "<command>", "</command>"
	start := strings.Index(lower, open)
	if start < 0 {
		return ""
	}
	start += len(open)
	end := strings.Index(lower[start:], close)
	if end < 0 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(s[start:start+end]), `"'`)
}

// scanDirWithAge walks a directory and collects items older than the given cutoff.
func scanDirWithAge(ctx context.Context, root, category string, risk types.RiskLevel, cutoff time.Time, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	var count int
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
		if d.IsDir() || cfg.IsExcluded(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.ModTime().After(cutoff) {
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
