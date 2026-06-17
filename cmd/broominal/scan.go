package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func scanCmd() *cobra.Command {
	var preset string
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan safe zones and print a summary",
		Long: `Scan all configured safe zones and print a category-by-category summary.

Each row shows the category name, total size, file count, and risk level:
  safe    — temp files, caches, logs; safe to delete
  review  — user-facing or recoverable files; inspect before cleaning

The scan does not modify any files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
			fmt.Printf("  %-32s  %10s  %6s  %s\n", "Category", "Size", "Files", "Risk")
			fmt.Printf("  %-32s  %10s  %6s  %s\n", strings.Repeat("─", 32), strings.Repeat("─", 10), strings.Repeat("─", 6), strings.Repeat("─", 8))
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
			return nil
		},
	}
	cmd.Flags().StringVar(&preset, "preset", "", "Scan preset: quick, standard, deep")
	return cmd
}
