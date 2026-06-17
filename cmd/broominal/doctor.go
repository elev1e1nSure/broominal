package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	var fix bool
	var yes bool
	cmd := &cobra.Command{
		Use:    "doctor",
		Short:  "Run health checks",
		Hidden: true, // Temporarily hidden for user as requested
		Long: `Run health checks and report the status of all system components.

Without flags, prints a status table and exits with code 1 if any check fails.

With --fix, also runs automatic fixes for any issues that support them.
Add --yes to skip the confirmation prompt.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := doctor.Run()
			var hasFail bool
			for _, c := range checks {
				var marker string
				switch c.Status {
				case doctor.StatusPass:
					marker = style.Passf(i18n.T("status_pass"))
				case doctor.StatusWarn:
					marker = style.Warnf(i18n.T("status_warn"))
				case doctor.StatusFail:
					marker = style.Failf(i18n.T("status_fail"))
					hasFail = true
				}
				fmt.Printf("%-24s %s  %s\n", style.Boldf(c.Name), marker, style.Grayf(c.Detail))
				if c.Status != doctor.StatusPass {
					if c.FixKey != "" && !fix {
						fmt.Printf("  → %s\n", style.Grayf("run 'broominal doctor --fix --yes' to repair automatically"))
					} else if c.FixKey == "" && c.Suggestion != "" {
						fmt.Printf("  → %s\n", style.Grayf(c.Suggestion))
					}
				}
			}

			if !fix {
				if hasFail {
					os.Exit(1)
				}
				return nil
			}

			// Collect fixable checks.
			var fixable []doctor.Check
			for _, c := range checks {
				if c.FixKey != "" {
					fixable = append(fixable, c)
				}
			}
			if len(fixable) == 0 {
				fmt.Println("\n  Nothing to fix.")
				return nil
			}

			fmt.Println()
			for _, c := range fixable {
				fmt.Printf("  %s  %s\n", style.Warnf("→"), style.Boldf(c.Name))
			}

			if !yes {
				fmt.Printf("\n  Run with --yes to apply %s.\n",
					pluralize(len(fixable), "this fix", fmt.Sprintf("these %d fixes", len(fixable))),
				)
				return nil
			}

			fmt.Println()
			for _, c := range fixable {
				result, err := doctor.Fix(c.FixKey)
				if err != nil {
					fmt.Printf("  %s  %s: %s\n", style.Failf("[FAIL]"), style.Boldf(c.Name), err)
				} else {
					fmt.Printf("  %s  %s: %s\n", style.Passf("[OK]"), style.Boldf(c.Name), result)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Run automatic fixes for issues that support them")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation when used with --fix")
	return cmd
}
