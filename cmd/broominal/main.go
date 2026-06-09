package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/elev1e1nSure/broominal/internal/tui"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
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
			id, freed, files, err := quarantine.Move(selected)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Clean failed: %v\n", err)
				os.Exit(1)
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
	return cmd
}

func restoreCmd() *cobra.Command {
	return &cobra.Command{
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
			if err := quarantine.Restore(id); err != nil {
				fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Restored cleanup %s\n", id)
		},
	}
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
