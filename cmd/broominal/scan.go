package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func scanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan safe zones and show summary",
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
			fmt.Println(style.Boldf("Scan complete:"))
			for _, c := range res.Categories {
				var riskCol string
				switch c.Risk {
				case types.RiskSafe:
					riskCol = style.Greenf(string(c.Risk))
				case types.RiskReview:
					riskCol = style.Yellowf(string(c.Risk))
				case types.RiskDanger:
					riskCol = style.Redf(string(c.Risk))
				}
				fmt.Printf("  %-20s %10s  %s\n", c.Category, util.FormatSize(c.Size), riskCol)
			}
			fmt.Println()
			fmt.Printf("Total: %s | Safe: %s | Review: %s | Danger: %s\n",
				style.Cyanf(util.FormatSize(res.TotalSize)),
				style.Greenf(util.FormatSize(res.SafeSize)),
				style.Yellowf(util.FormatSize(res.ReviewSize)),
				style.Redf(util.FormatSize(res.DangerSize)),
			)
		},
	}
}
