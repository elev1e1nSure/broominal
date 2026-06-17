package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/cleaner"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	var yes bool
	var preset string
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Scan and move junk to quarantine",
		Long: `Scan configured safe zones and move all found items to quarantine.

Without --yes, a dry-run preview is printed and nothing is changed.
Add --yes to execute the cleanup.

Files are never deleted — they are moved to a quarantine directory and can
be restored at any time with: broominal restore last

Use --preset to control which categories are included:
  quick     — low-risk only, fastest
  standard  — balanced (default)
  deep      — includes browser caches and other heavy items`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cfg, _ := config.Load()
			if cfg == nil {
				cfg = config.Default()
			}
			if err := applyPresetFlag(cfg, preset); err != nil {
				return err
			}
			res, err := scanner.ScanWithConfig(ctx, cfg, nil)
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if !yes {
				printCleanPreview(res)
				fmt.Printf("\n  Run with %s to proceed.\n", style.Yellowf("--yes"))
				return nil
			}

			var selected []types.Item
			for _, cat := range res.Categories {
				for i := range cat.Items {
					cat.Items[i].Selected = true
					selected = append(selected, cat.Items[i])
				}
			}
			cleanResult, err := cleaner.Run(ctx, selected, res, cfg)
			if err != nil {
				return fmt.Errorf("clean failed: %w", err)
			}
			msg := fmt.Sprintf("Freed %s in %s",
				style.Cyanf(util.FormatSize(cleanResult.Freed)),
				style.Boldf("%s", pluralize(cleanResult.Files, "1 file", fmt.Sprintf("%d files", cleanResult.Files))),
			)
			if cleanResult.Skipped > 0 {
				msg += fmt.Sprintf("  (%s skipped)", style.Yellowf("%d", cleanResult.Skipped))
			}
			fmt.Printf("  %s\n  Restore ID: %s\n", msg, style.Yellowf(cleanResult.RestoreID))
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Execute the cleanup (default: dry-run preview)")
	cmd.Flags().StringVar(&preset, "preset", "", "Scan preset: quick, standard, deep")
	return cmd
}

func printCleanPreview(res *types.ScanResult) {
	fmt.Printf("  %-32s  %10s  %6s  %s\n", "Category", "Size", "Files", "Risk")
	fmt.Printf("  %-32s  %10s  %6s  %s\n", "────────────────────────────────", "──────────", "──────", "────────")
	for _, c := range res.Categories {
		label, colorFn := riskDisplay(c.Risk)
		name := i18n.CategoryName(c.Category)
		fmt.Printf("  %-32s  %10s  %6d  %s\n",
			truncateCLI(name, 32),
			util.FormatSize(c.Size),
			c.Files,
			colorFn("%s", label),
		)
	}
	fmt.Println()
	fmt.Printf("  Total %s  ·  Safe %s  ·  Review %s  ·  Danger %s\n",
		util.FormatSize(res.TotalSize),
		util.FormatSize(res.SafeSize),
		util.FormatSize(res.ReviewSize),
		util.FormatSize(res.DangerSize),
	)
}
