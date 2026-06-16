package main

import (
	"fmt"

	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

// quarantineCleanupCmd returns a hidden backward-compatibility alias for
// scheduled tasks created before the quarantine subcommand redesign.
// New installs use: broominal quarantine clean --yes --max-age-days N
func quarantineCleanupCmd() *cobra.Command {
	var force bool
	var maxAgeDays int
	cmd := &cobra.Command{
		Use:    "quarantine-cleanup",
		Short:  "Delete old quarantines (deprecated: use 'quarantine clean')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxAgeDays <= 0 {
				maxAgeDays = 30
			}
			if !force {
				return quarantineCleanPreview(maxAgeDays)
			}
			deleted, freed, err := quarantine.Cleanup(maxAgeDays)
			if err != nil {
				if util.IsFileLocked(err) {
					return fmt.Errorf("%s", i18n.T("quarantine_locked"))
				}
				return fmt.Errorf("cleanup failed: %w", err)
			}
			if deleted == 0 {
				fmt.Println("  Nothing to clean.")
				return nil
			}
			fmt.Printf("  Deleted %s, freed %s.\n",
				style.Boldf("%s", pluralize(deleted, "1 backup", fmt.Sprintf("%d backups", deleted))),
				style.Cyanf(util.FormatSize(freed)),
			)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Execute deletion without prompt (deprecated: use --yes)")
	cmd.Flags().BoolVar(&force, "yes", false, "Execute deletion without prompt")
	cmd.Flags().IntVar(&maxAgeDays, "max-age-days", 30, "Delete backups older than this many days")
	return cmd
}
