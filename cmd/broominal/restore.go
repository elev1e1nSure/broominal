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
		Use:   "restore [id|date-prefix|last]",
		Short: "Restore a backup by ID, date prefix (e.g. 2025-06-09), or 'last'",
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
			} else {
				ids, err := quarantine.List()
				if err != nil {
					fmt.Fprintf(os.Stderr, "List failed: %v\n", err)
					os.Exit(1)
				}
				var matches []string
				for _, rid := range ids {
					if strings.HasPrefix(rid, id) || rid == id {
						matches = append(matches, rid)
					}
				}
				if len(matches) == 0 {
					fmt.Fprintf(os.Stderr, "No backup found matching %q\n", id)
					os.Exit(1)
				}
				if len(matches) > 1 {
					fmt.Printf("%s\n", style.Yellowf("Multiple backups found, please be more specific:"))
					for _, m := range matches {
						mf, _ := quarantine.GetManifest(m)
						if mf != nil {
							fmt.Printf("  %s  %s  %s, %d files\n", m, mf.CreatedAt.Format("2006-01-02 15:04"), util.FormatSize(mf.TotalSize), mf.Files)
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
