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

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// walkMatch decides whether a file path should be included in scan results.
type walkMatch func(path string, d fs.DirEntry) bool

func walkDirItems(ctx context.Context, root, category string, risk types.RiskLevel, recursive bool, matchExt []string, match walkMatch, cfg *config.Config) ([]types.Item, error) {
	if root == "" {
		return nil, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan %s: stat %s: %w", category, root, err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	var items []types.Item
	var count int

	var extSet map[string]struct{}
	if len(matchExt) > 0 {
		extSet = make(map[string]struct{}, len(matchExt))
		for _, e := range matchExt {
			extSet[e] = struct{}{}
		}
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if walkErr != nil {
			if errors.Is(walkErr, os.ErrPermission) {
				return filepath.SkipDir
			}
			if errors.Is(walkErr, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("scan %s: walk error at %s: %w", category, path, walkErr)
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
		if match != nil && !match(path, d) {
			return nil
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

func scanPath(ctx context.Context, path, category string, risk types.RiskLevel, recursive bool, cfg *config.Config) ([]types.Item, error) {
	return walkDirItems(ctx, path, category, risk, recursive, nil, nil, cfg)
}

func scanPaths(ctx context.Context, paths []string, category string, risk types.RiskLevel, recursive bool, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range paths {
		sub, err := scanPath(ctx, path, category, risk, recursive, cfg)
		if err != nil {
			return nil, err
		}
		items = append(items, sub...)
	}
	return items, nil
}
