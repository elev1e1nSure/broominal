package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func quarantineCleanupCmd() *cobra.Command {
	var force bool
	var maxAgeDays int
	cmd := &cobra.Command{
		Use:   "quarantine-cleanup",
		Short: "Delete old quarantines",
		Run: func(cmd *cobra.Command, args []string) {
			if maxAgeDays <= 0 {
				maxAgeDays = 30
			}
			deleted, freed, err := quarantine.Cleanup(maxAgeDays)
			if err != nil {
				if util.IsFileLocked(err) {
					fmt.Fprintf(os.Stderr, "%s\n", i18n.T("quarantine_locked"))
				} else {
					fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
				}
				os.Exit(1)
			}
			if deleted == 0 {
				fmt.Println(style.Greenf("No old quarantines to remove."))
				return
			}
			if !force {
				fmt.Printf("Will remove %s quarantine(s) (%s)\n", style.Boldf("%d", deleted), style.Cyanf(util.FormatSize(freed)))
				fmt.Printf("Use %s to proceed.\n", style.Yellowf("--force"))
				return
			}
			fmt.Printf("%s %s quarantine(s), freed %s\n", style.Greenf("Removed"), style.Boldf("%d", deleted), style.Cyanf(util.FormatSize(freed)))
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion without prompt")
	cmd.Flags().IntVar(&maxAgeDays, "max-age-days", 0, "Override max age (default from config or 30)")
	return cmd
}
