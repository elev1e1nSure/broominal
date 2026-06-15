package main

import (
	"context"
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate a cleanup report",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			if cfg == nil {
				cfg = config.Default()
			}
			res, err := scanner.ScanWithConfig(context.Background(), cfg, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
				os.Exit(1)
			}
			path, err := report.Save(res, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Report failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s %s\n", style.Greenf("Report saved to:"), style.Cyanf(path))
		},
	}
}
