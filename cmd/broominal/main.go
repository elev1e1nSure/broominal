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
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func init() {
	cobra.AddTemplateFunc("bold", func(s string) string { return style.Bold + s + style.Reset })
	cobra.AddTemplateFunc("green", func(s string) string { return style.Green + s + style.Reset })
	cobra.AddTemplateFunc("yellow", func(s string) string { return style.Yellow + s + style.Reset })
	cobra.AddTemplateFunc("red", func(s string) string { return style.Red + s + style.Reset })
	cobra.AddTemplateFunc("cyan", func(s string) string { return style.Cyan + s + style.Reset })
	cobra.AddTemplateFunc("gray", func(s string) string { return style.Gray + s + style.Reset })
}

const helpTemplate = `{{with (or .Long .Short)}}{{bold .}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

const usageTemplate = `{{bold "Usage:"}}
{{if .Runnable}}  {{cyan .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{cyan .CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{bold "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{bold "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{bold "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{cyan (rpad .Name .NamePadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{bold "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{bold "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{bold "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{gray "Use"}} {{cyan .CommandPath}} [command] --help {{gray "for more information about a command."}}{{end}}
`

func main() {
	var rootCmd = &cobra.Command{
		Use:   "broominal",
		Short: "Safe Windows cleanup with undo",
		Long:  "A safe, transparent, undoable Windows cleanup tool.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := tui.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

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
	var danger bool
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
			id, freed, files, err := quarantine.Move(selected, dryRun)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Clean failed: %v\n", err)
				os.Exit(1)
			}
			if dryRun {
				fmt.Printf("%s Would free %s in %d files\n", style.Yellowf("[dry-run]"), style.Cyanf(util.FormatSize(freed)), files)
				return
			}
			_, _ = report.Save(res, &types.CleanResult{
				RestoreID: id,
				Freed:     freed,
				Files:     files,
			})
			fmt.Printf("%s %s in %s files. Restore ID: %s\n", style.Greenf("Cleaned"), style.Cyanf(util.FormatSize(freed)), style.Boldf("%d", files), style.Yellowf(id))
		},
	}
	cmd.Flags().BoolVar(&safeOnly, "safe", false, "Only clean safe items")
	cmd.Flags().BoolVar(&danger, "danger", false, "Allow cleaning danger items")
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
			fmt.Printf("%s %s files from %s", style.Greenf("Restored"), style.Boldf("%d", restored), style.Yellowf(id))
			if skipped > 0 {
				fmt.Printf(" (%s)", style.Yellowf("%d skipped", skipped))
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
			fmt.Printf("%s %s\n", style.Greenf("Report saved to:"), style.Cyanf(path))
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
			fmt.Printf("%s %s\n\n%s\n", style.Boldf("Config path:"), style.Cyanf(config.Path()), string(data))
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
				var marker string
				switch c.Status {
				case doctor.StatusPass:
					marker = style.Passf("[PASS]")
				case doctor.StatusWarn:
					marker = style.Warnf("[WARN]")
				case doctor.StatusFail:
					marker = style.Failf("[FAIL]")
					fail = true
				}
				fmt.Printf("%-24s %s  %s\n", style.Boldf(c.Name), marker, style.Grayf(c.Detail))
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
				fmt.Println(style.Greenf("No old quarantines to remove."))
				return
			}
			fmt.Printf("Will remove %s quarantine(s) (%s)\n", style.Boldf("%d", deleted), style.Cyanf(util.FormatSize(freed)))
			if !force && !dryRun {
				fmt.Printf("Use %s to proceed.\n", style.Yellowf("--force"))
				return
			}
			if dryRun {
				fmt.Println(style.Yellowf("[dry-run] Nothing removed."))
				return
			}
			deleted, freed, err = quarantine.Cleanup(maxAgeDays, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s %s quarantine(s), freed %s\n", style.Greenf("Removed"), style.Boldf("%d", deleted), style.Cyanf(util.FormatSize(freed)))
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted")
	cmd.Flags().IntVar(&maxAgeDays, "max-age-days", 0, "Override max age (default from config or 30)")
	return cmd
}
