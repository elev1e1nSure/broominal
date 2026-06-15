package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/cleaner"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	var safeOnly bool
	var danger bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean selected items",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cfg, _ := config.Load()
			if cfg == nil {
				cfg = config.Default()
			}
			res, err := scanner.ScanWithConfig(ctx, cfg, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
				os.Exit(1)
			}
			var hasDanger bool
			for _, cat := range res.Categories {
				if cat.Risk == types.RiskDanger && len(cat.Items) > 0 {
					hasDanger = true
					break
				}
			}
			if hasDanger && !danger {
				fmt.Fprintf(os.Stderr, "%s Danger items found. Use %s to confirm.\n", style.Failf("[BLOCKED]"), style.Yellowf("--danger"))
				os.Exit(1)
			}
			var selected []types.Item
			for _, cat := range res.Categories {
				if safeOnly && cat.Risk != types.RiskSafe {
					continue
				}
				for _, it := range cat.Items {
					it.Selected = true
					selected = append(selected, it)
				}
			}
			cleanResult, err := cleaner.Run(ctx, selected, res, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Clean failed: %v\n", err)
				os.Exit(1)
			}
			msg := fmt.Sprintf("%s %s in %s files", style.Greenf("Cleaned"), style.Cyanf(util.FormatSize(cleanResult.Freed)), style.Boldf("%d", cleanResult.Files))
			if cleanResult.Skipped > 0 {
				msg += fmt.Sprintf(" (%s)", style.Yellowf("%d skipped", cleanResult.Skipped))
			}
			fmt.Printf("%s. Restore ID: %s\n", msg, style.Yellowf(cleanResult.RestoreID))
		},
	}
	cmd.Flags().BoolVar(&safeOnly, "safe", false, "Only clean safe items")
	cmd.Flags().BoolVar(&danger, "danger", false, "Allow cleaning danger items")
	return cmd
}
