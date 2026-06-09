package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/elev1e1nSure/broominal/internal/tui"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "broominal",
		Short: "Safe Windows cleanup with undo",
		Long:  `A safe, transparent, undoable Windows cleanup tool.`,
	}

	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(uiCmd())
	rootCmd.AddCommand(cleanCmd())
	rootCmd.AddCommand(restoreCmd())
	rootCmd.AddCommand(reportCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(quarantineCleanupCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func scanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan safe zones and show summary",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := scanner.Scan()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Scan complete:")
			for _, c := range res.Categories {
				fmt.Printf("  %-20s %10s  %s\n", c.Category, scanner.FormatSize(c.Size), c.Risk)
			}
			fmt.Println()
			fmt.Printf("Total: %s | Safe: %s | Review: %s | Danger: %s\n",
				scanner.FormatSize(res.TotalSize),
				scanner.FormatSize(res.SafeSize),
				scanner.FormatSize(res.ReviewSize),
				scanner.FormatSize(res.DangerSize),
			)
		},
	}
}

func uiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch interactive TUI",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func cleanCmd() *cobra.Command {
	var safeOnly bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean selected items",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := scanner.Scan()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
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
			id, freed, files, err := quarantine.Move(selected, dryRun)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Clean failed: %v\n", err)
				os.Exit(1)
			}
			if dryRun {
				fmt.Printf("[dry-run] Would free %s in %d files\n", scanner.FormatSize(freed), files)
				return
			}
			_, _ = report.Save(res, &types.CleanResult{
				RestoreID: id,
				Freed:     freed,
				Files:     files,
			})
			fmt.Printf("Cleaned %s in %d files. Restore ID: %s\n", scanner.FormatSize(freed), files, id)
		},
	}
	cmd.Flags().BoolVar(&safeOnly, "safe", false, "Only clean safe items")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate cleaning without moving files")
	return cmd
}

func restoreCmd() *cobra.Command {
	var forceOverwrite bool
	cmd := &cobra.Command{
		Use:   "restore [id]",
		Short: "Restore a cleanup (use 'last' for most recent)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			if id == "last" {
				lastID, err := quarantine.GetLast()
				if err != nil {
					fmt.Fprintf(os.Stderr, "No cleanups found: %v\n", err)
					os.Exit(1)
				}
				id = lastID
			}
			restored, skipped, err := quarantine.Restore(id, forceOverwrite)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Restored %d files from %s", restored, id)
			if skipped > 0 {
				fmt.Printf(" (%d skipped due to conflicts)", skipped)
			}
			fmt.Println()
		},
	}
	cmd.Flags().BoolVar(&forceOverwrite, "force-overwrite", false, "Overwrite existing files")
	return cmd
}

func reportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate a cleanup report",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := scanner.Scan()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
				os.Exit(1)
			}
			path, err := report.Save(res, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Report failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Report saved to: %s\n", path)
		},
	}
}

func configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("Config path: %s\n\n%s\n", config.Path(), string(data))
		},
	}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks",
		Run: func(cmd *cobra.Command, args []string) {
			checks := doctor.Run()
			var fail bool
			for _, c := range checks {
				marker := "  "
				switch c.Status {
				case doctor.StatusPass:
					marker = "[PASS]"
				case doctor.StatusWarn:
					marker = "[WARN]"
				case doctor.StatusFail:
					marker = "[FAIL]"
					fail = true
				}
				fmt.Printf("%-8s %-25s %s\n", marker, c.Name, c.Detail)
			}
			if fail {
				os.Exit(1)
			}
		},
	}
}

func quarantineCleanupCmd() *cobra.Command {
	var force bool
	var dryRun bool
	var maxAgeDays int
	cmd := &cobra.Command{
		Use:   "quarantine-cleanup",
		Short: "Delete old quarantines",
		Run: func(cmd *cobra.Command, args []string) {
			if maxAgeDays <= 0 {
				cfg, _ := config.Load()
				if cfg != nil && cfg.QuarantineMaxAgeDays > 0 {
					maxAgeDays = cfg.QuarantineMaxAgeDays
				} else {
					maxAgeDays = 30
				}
			}
			deleted, freed, err := quarantine.Cleanup(maxAgeDays, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cleanup check failed: %v\n", err)
				os.Exit(1)
			}
			if deleted == 0 {
				fmt.Println("No old quarantines to remove.")
				return
			}
			fmt.Printf("Will remove %d quarantine(s) (%s)\n", deleted, scanner.FormatSize(freed))
			if !force && !dryRun {
				fmt.Println("Use --force to proceed.")
				return
			}
			if dryRun {
				fmt.Println("[dry-run] Nothing removed.")
				return
			}
			deleted, freed, err = quarantine.Cleanup(maxAgeDays, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Removed %d quarantine(s), freed %s\n", deleted, scanner.FormatSize(freed))
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted")
	cmd.Flags().IntVar(&maxAgeDays, "max-age-days", 0, "Override max age (default from config or 30)")
	return cmd
}
