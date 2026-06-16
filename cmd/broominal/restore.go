package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func restoreCmd() *cobra.Command {
	var forceOverwrite bool
	cmd := &cobra.Command{
		Use:   "restore <id|prefix|last>",
		Short: "Restore a quarantined backup",
		Long: `Restore files from a quarantine backup to their original locations.

The backup is identified by:
  last       — most recent backup
  <id>       — exact quarantine ID
  <prefix>   — any unique prefix of an ID (e.g. "2026-06-16")

If a prefix matches more than one backup, the candidates are listed and
nothing is restored. Narrow the prefix and try again.

Files already present at the destination are skipped unless --force-overwrite
is set.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if id == "last" {
				lastID, err := quarantine.GetLast()
				if err != nil {
					return fmt.Errorf("no backups found: %w", err)
				}
				id = lastID
			} else {
				ids, err := quarantine.List()
				if err != nil {
					return fmt.Errorf("list failed: %w", err)
				}
				var matches []string
				for _, rid := range ids {
					if rid == id || strings.HasPrefix(rid, id) {
						matches = append(matches, rid)
					}
				}
				if len(matches) == 0 {
					fmt.Fprintf(os.Stderr, "No backup found matching %q\n", id)
					os.Exit(1)
				}
				if len(matches) > 1 {
					fmt.Printf("  Multiple backups match %q — be more specific:\n\n", id)
					for _, m := range matches {
						mf, _ := quarantine.GetManifest(m)
						if mf != nil {
							cats := formatCLICategories(mf.Categories)
							fmt.Printf("  %s  %9s  %s\n",
								cliDate(mf.CreatedAt),
								util.FormatSize(mf.TotalSize),
								cats,
							)
						} else {
							fmt.Printf("  %s\n", m)
						}
					}
					os.Exit(1)
				}
				id = matches[0]
			}
			restored, skipped, err := quarantine.Restore(id, forceOverwrite)
			if err != nil {
				return fmt.Errorf("restore failed: %w", err)
			}
			msg := fmt.Sprintf("Restored %s from %s",
				style.Boldf("%s", pluralize(restored, "1 file", fmt.Sprintf("%d files", restored))),
				style.Yellowf(id),
			)
			if skipped > 0 {
				msg += fmt.Sprintf("  (%s skipped)", style.Yellowf("%d", skipped))
			}
			fmt.Println("  " + msg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&forceOverwrite, "force-overwrite", false, "Overwrite files that already exist at the destination")
	return cmd
}
