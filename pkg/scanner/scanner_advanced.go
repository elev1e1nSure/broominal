package scanner

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanRecycleBin(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, rp := range recycleBinPaths() {
		sub, err := scanDir(ctx, rp, "recycle_bin", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: recycle bin scan failed", "path", rp, "error", err)
		}
		items = append(items, sub...)
	}
	return items, nil
}

func scanStartupLeftovers(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	paths := []string{
		filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup"),
		filepath.Join(os.Getenv("PROGRAMDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "StartUp"),
	}
	for _, path := range paths {
		subItems, err := scanDir(ctx, path, "startup_leftover", types.RiskReview, []string{".lnk", ".url"}, false, cfg)
		if err != nil {
			slog.Warn("scanner: startup leftovers scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanScheduledTasksLeftovers(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	root := filepath.Join(os.Getenv("SystemRoot"), "System32", "Tasks")
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
			return nil
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			top := strings.SplitN(rel, string(filepath.Separator), 2)[0]
			if strings.EqualFold(top, "Microsoft") {
				return filepath.SkipDir
			}
			return nil
		}
		if cfg.IsExcluded(path) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		cmd := extractTaskCommand(data)
		if cmd == "" {
			return nil
		}
		expanded := os.ExpandEnv(cmd)
		if _, statErr := os.Stat(expanded); statErr == nil {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		count++
		if count > maxScanFiles {
			slog.Info("scan: max file limit reached, truncating", "category", "scheduled_tasks_leftover", "limit", maxScanFiles)
			return errScanLimit
		}
		items = append(items, types.Item{
			Category: "scheduled_tasks_leftover",
			Path:     path,
			Size:     info.Size(),
			Risk:     types.RiskReview,
		})
		return nil
	})
	return items, nil
}

func scanDuplicateFiles(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	const minSize = 1 * 1024 * 1024 // 1 MB
	roots := []string{
		filepath.Join(os.Getenv("USERPROFILE"), "Downloads"),
		filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
		filepath.Join(os.Getenv("USERPROFILE"), "Documents"),
	}
	type entry struct {
		path string
		size int64
	}
	bySize := make(map[int64][]entry)
	for _, root := range roots {
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
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if cfg.IsExcluded(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil || info.Size() < minSize {
				return nil
			}
			bySize[info.Size()] = append(bySize[info.Size()], entry{path: path, size: info.Size()})
			return nil
		})
	}
	byHash := make(map[string][]entry)
	for _, files := range bySize {
		if len(files) < 2 {
			continue
		}
		for _, f := range files {
			if ctx.Err() != nil {
				break
			}
			h, err := hashFileMD5(f.path)
			if err != nil {
				slog.Warn("scanner: duplicate hash failed", "path", f.path, "error", err)
				continue
			}
			byHash[h] = append(byHash[h], f)
		}
	}
	var items []types.Item
	for _, entries := range byHash {
		if len(entries) < 2 {
			continue
		}
		for _, e := range entries[1:] {
			items = append(items, types.Item{
				Category: "duplicate_files",
				Path:     e.path,
				Size:     e.size,
				Risk:     types.RiskReview,
			})
		}
	}
	return items, nil
}

func scanRecentDocuments(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Recent")
	items, err := scanDir(ctx, path, "recent_documents", types.RiskReview, []string{".lnk"}, false, cfg)
	if err != nil {
		slog.Warn("scanner: recent documents scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanCrashMemoryDumps(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	cutoff := time.Now().AddDate(0, 0, -7)
	var items []types.Item
	paths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "CrashDumps"),
		filepath.Join(os.Getenv("SystemRoot"), "Minidump"),
	}
	for _, path := range paths {
		subItems, err := scanDir(ctx, path, "crash_dumps", types.RiskReview, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: crash dumps scan failed", "path", path, "error", err)
		}
		for _, it := range subItems {
			info, err := os.Stat(it.Path)
			if err != nil || info.ModTime().After(cutoff) {
				continue
			}
			items = append(items, it)
		}
	}
	memDmp := filepath.Join(os.Getenv("SystemRoot"), "MEMORY.DMP")
	if info, err := os.Stat(memDmp); err == nil && !info.IsDir() && info.ModTime().Before(cutoff) {
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
			slog.Warn("scanner: nvidia installer leftovers scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}
