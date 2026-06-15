package scanner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func scanDirectXShaderCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	path := filepath.Join(os.Getenv("LOCALAPPDATA"), "D3DSCache")
	items, err := scanDir(ctx, path, "directx_shader_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: directx shader cache scan failed", "path", path, "error", err)
	}
	return items, nil
}

func scanAMDGPUCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "AMD", "DxCache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "AMD", "CLCache"),
	} {
		subItems, err := scanDir(ctx, path, "amd_gpu_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: amd gpu cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanNvidiaShaderCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "NVIDIA", "DXCache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "NVIDIA", "GLCache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "NVIDIA Corporation", "NV_Cache"),
	} {
		subItems, err := scanDir(ctx, path, "nvidia_shader_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: nvidia shader cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}

func scanIntelGpuCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item
	for _, path := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Intel", "ShaderCache"),
		filepath.Join(os.Getenv("ProgramData"), "Intel", "ShaderCache"),
	} {
		subItems, err := scanDir(ctx, path, "intel_gpu_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: intel gpu cache scan failed", "path", path, "error", err)
		}
		items = append(items, subItems...)
	}
	return items, nil
}
