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

// scanPkgManagerCache finds download caches left by Windows package managers:
// Chocolatey, Scoop, winget (source metadata), pip wheels, and Conda/Miniconda/Anaconda.
// All of these are safe to remove — they are rebuilt on the next install or update operation.
func scanPkgManagerCache(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	var items []types.Item

	// Chocolatey stores its download cache under <ChocolateyInstall>\cache.
	// If the env var is not set, fall back to the default ProgramData location.
	chocoRoot := os.Getenv("ChocolateyInstall")
	if chocoRoot == "" {
		chocoRoot = filepath.Join(os.Getenv("ProgramData"), "chocolatey")
	}
	for _, p := range []string{
		filepath.Join(chocoRoot, "cache"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Choco-cache"),
	} {
		sub, err := scanDir(ctx, p, "pkg_manager_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: chocolatey cache scan failed", "path", p, "error", err)
		}
		items = append(items, sub...)
	}

	// Scoop keeps downloaded archives in <SCOOP>\cache (env var) or the
	// default %USERPROFILE%\scoop\cache.
	scoopRoot := os.Getenv("SCOOP")
	if scoopRoot == "" {
		scoopRoot = filepath.Join(os.Getenv("USERPROFILE"), "scoop")
	}
	sub, err := scanDir(ctx, filepath.Join(scoopRoot, "cache"), "pkg_manager_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: scoop cache scan failed", "path", scoopRoot, "error", err)
	}
	items = append(items, sub...)

	// winget does not keep a persistent download cache in the user profile —
	// binaries go through %TEMP% (already covered by Temp category).
	// What it does keep is source metadata index files under the DesktopAppInstaller
	// package. These are rebuilt automatically by `winget source update`.
	wingetPkgs := filepath.Join(
		os.Getenv("LOCALAPPDATA"),
		"Packages",
		"Microsoft.DesktopAppInstaller_8wekyb3d8bbwe",
		"LocalState",
	)
	entries, _ := os.ReadDir(wingetPkgs)
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "Microsoft.Winget.Source_") {
			continue
		}
		p := filepath.Join(wingetPkgs, e.Name())
		sub, err = scanDir(ctx, p, "pkg_manager_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: winget source cache scan failed", "path", p, "error", err)
		}
		items = append(items, sub...)
	}

	// pip's http cache is already included in Dev Cache (scanDevCache scans
	// %LOCALAPPDATA%\pip\cache). Only the wheels sub-directory is unique here:
	// built wheel archives that pip reuses across installs.
	pipWheels := filepath.Join(os.Getenv("LOCALAPPDATA"), "pip", "cache", "wheels")
	sub, err = scanDir(ctx, pipWheels, "pkg_manager_cache", types.RiskSafe, nil, true, cfg)
	if err != nil {
		slog.Warn("scanner: pip wheels cache scan failed", "path", pipWheels, "error", err)
	}
	items = append(items, sub...)

	// Conda, Miniconda and Anaconda store downloaded package tarballs in a
	// central pkgs\ directory inside the installation root. This is equivalent
	// to `conda clean --packages` and is safe to remove.
	for _, condaRoot := range resolveCondaRoots() {
		p := filepath.Join(condaRoot, "pkgs")
		sub, err = scanDir(ctx, p, "pkg_manager_cache", types.RiskSafe, nil, true, cfg)
		if err != nil {
			slog.Warn("scanner: conda pkgs cache scan failed", "path", p, "error", err)
		}
		items = append(items, sub...)
	}

	return items, nil
}

// resolveCondaRoots returns the list of conda/miniconda/anaconda installation
// root directories found on this machine.
//
// Search order:
//  1. CONDA_EXE env var — strips the trailing \Scripts\conda.exe suffix to get
//     the installation root. Most reliable when conda is active in the shell.
//  2. Well-known default install locations for user and system-wide installs.
//
// Duplicates are suppressed so the same root is not scanned twice.
func resolveCondaRoots() []string {
	seen := map[string]struct{}{}
	var roots []string

	add := func(p string) {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		if _, err := os.Stat(p); err != nil {
			return
		}
		seen[p] = struct{}{}
		roots = append(roots, p)
	}

	// CONDA_EXE points to <root>\Scripts\conda.exe; walk two levels up.
	if exe := os.Getenv("CONDA_EXE"); exe != "" {
		add(filepath.Dir(filepath.Dir(exe)))
	}

	for _, p := range []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "miniconda3"),
		filepath.Join(os.Getenv("USERPROFILE"), "miniconda3"),
		filepath.Join(os.Getenv("USERPROFILE"), "anaconda3"),
		filepath.Join(os.Getenv("ProgramFiles"), "miniconda3"),
		filepath.Join(os.Getenv("ProgramData"), "miniconda3"),
		filepath.Join(os.Getenv("ProgramData"), "anaconda3"),
	} {
		add(p)
	}

	return roots
}
