package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
	var preset string
	var output string
	var stdout bool
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Scan and save a JSON report",
		Long: `Scan configured safe zones and save a JSON report.

By default the report is written to %LOCALAPPDATA%\broominal\reports\.
Use --output to choose a specific file path, or --stdout to print JSON
to standard output instead of saving to disk.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load()
			if cfg == nil {
				cfg = config.Default()
			}
			if err := applyPresetFlag(cfg, preset); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			res, err := scanner.ScanWithConfig(ctx, cfg, nil)
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if stdout {
				data := types.ReportData{
					Timestamp: time.Now(),
					Result:    *res,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}

			var path string
			if output != "" {
				if err := report.SaveTo(output, res, nil); err != nil {
					return fmt.Errorf("save report: %w", err)
				}
				path = output
			} else {
				var err error
				path, err = report.Save(res, nil)
				if err != nil {
					return fmt.Errorf("save report: %w", err)
				}
			}
			fmt.Printf("  %s %s\n", style.Passf("[OK]"), style.Cyanf(path))
			report.PrintSummary(res, nil)
			return nil
		},
	}
	cmd.Flags().StringVar(&preset, "preset", "", "Scan preset: quick, standard, deep")
	cmd.Flags().StringVar(&output, "output", "", "Write report to this file path")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "Print JSON to stdout instead of saving to disk")
	return cmd
}
