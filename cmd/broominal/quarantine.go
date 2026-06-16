package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/util"
	"github.com/spf13/cobra"
)

func quarantineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quarantine",
		Short: "Manage quarantine backups",
		Long: `List, clean up, or delete quarantine backups.

Quarantine backups are created by broominal clean. Files are never
permanently deleted until you explicitly run quarantine clean or
quarantine delete.`,
	}
	cmd.AddCommand(quarantineListCmd())
	cmd.AddCommand(quarantineCleanCmd())
	cmd.AddCommand(quarantineDeleteCmd())
	return cmd
}

func quarantineListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all quarantine backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := quarantine.List()
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}
			if len(ids) == 0 {
				fmt.Println("  No quarantine backups.")
				return nil
			}
			fmt.Printf("  %-16s  %9s  %s\n", "Date", "Size", "Categories")
			fmt.Printf("  %-16s  %9s  %s\n", strings.Repeat("─", 16), strings.Repeat("─", 9), strings.Repeat("─", 30))
			for _, id := range ids {
				mf, err := quarantine.GetManifest(id)
				if err != nil || mf == nil {
					continue
				}
				cats := formatCLICategories(mf.Categories)
				fmt.Printf("  %-16s  %9s  %s\n",
					cliDate(mf.CreatedAt),
					util.FormatSize(mf.TotalSize),
					cats,
				)
			}
			return nil
		},
	}
}

func quarantineCleanCmd() *cobra.Command {
	var yes bool
	var maxAgeDays int
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Delete quarantine backups older than N days",
		Long: `Delete quarantine backups whose age exceeds --max-age-days.

Without --yes, a preview is printed and nothing is deleted.
Add --yes to permanently delete the listed backups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxAgeDays <= 0 {
				maxAgeDays = 30
			}

			if !yes {
				return quarantineCleanPreview(maxAgeDays)
			}
			deleted, freed, err := quarantine.Cleanup(maxAgeDays)
			if err != nil {
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
	cmd.Flags().BoolVar(&yes, "yes", false, "Execute the deletion (default: dry-run preview)")
	cmd.Flags().IntVar(&maxAgeDays, "max-age-days", 30, "Delete backups older than this many days")
	return cmd
}

func quarantineCleanPreview(maxAgeDays int) error {
	ids, err := quarantine.List()
	if err != nil {
		return fmt.Errorf("list failed: %w", err)
	}
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	var count int
	var freed int64
	for _, id := range ids {
		mf, err := quarantine.GetManifest(id)
		if err != nil || mf == nil {
			continue
		}
		if mf.CreatedAt.IsZero() || mf.CreatedAt.Before(cutoff) {
			count++
			freed += mf.TotalSize
			cats := formatCLICategories(mf.Categories)
			fmt.Printf("  %s  %9s  %s\n", cliDate(mf.CreatedAt), util.FormatSize(mf.TotalSize), cats)
		}
	}
	if count == 0 {
		fmt.Printf("  Nothing to clean (max-age-days: %d).\n", maxAgeDays)
		return nil
	}
	fmt.Printf("\n  %s would be deleted, freeing %s. Run with %s to proceed.\n",
		pluralize(count, "1 backup", fmt.Sprintf("%d backups", count)),
		util.FormatSize(freed),
		style.Yellowf("--yes"),
	)
	return nil
}

func quarantineDeleteCmd() *cobra.Command {
	var all bool
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Permanently delete a specific quarantine backup",
		Long: `Permanently delete a quarantine backup by ID.

Use --all --yes to delete all backups at once.

Without --yes, a preview is printed and nothing is deleted.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				if !yes {
					return quarantineDeleteAllPreview()
				}
				count, freed, err := quarantine.CleanupAll()
				if err != nil {
					return fmt.Errorf("delete all failed: %w", err)
				}
				if count == 0 {
					fmt.Println("  No backups to delete.")
					return nil
				}
				fmt.Printf("  Deleted %s, freed %s.\n",
					style.Boldf("%s", pluralize(count, "1 backup", fmt.Sprintf("%d backups", count))),
					style.Cyanf(util.FormatSize(freed)),
				)
				return nil
			}
			if len(args) == 0 {
				return fmt.Errorf("provide a backup ID or use --all")
			}
			id := args[0]
			mf, err := quarantine.GetManifest(id)
			if err != nil {
				return fmt.Errorf("backup %q not found: %w", id, err)
			}
			cats := formatCLICategories(mf.Categories)
			if !yes {
				fmt.Printf("  Would delete: %s  %s  %s\n",
					cliDate(mf.CreatedAt),
					util.FormatSize(mf.TotalSize),
					cats,
				)
				fmt.Printf("\n  Run with %s to proceed.\n", style.Yellowf("--yes"))
				return nil
			}
			freed, err := quarantine.Delete(id)
			if err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}
			fmt.Printf("  Deleted backup %s, freed %s.\n",
				style.Yellowf(id),
				style.Cyanf(util.FormatSize(freed)),
			)
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Delete all backups")
	cmd.Flags().BoolVar(&yes, "yes", false, "Execute the deletion (default: dry-run preview)")
	return cmd
}

func quarantineDeleteAllPreview() error {
	ids, err := quarantine.List()
	if err != nil {
		return fmt.Errorf("list failed: %w", err)
	}
	if len(ids) == 0 {
		fmt.Println("  No backups to delete.")
		return nil
	}
	var totalFreed int64
	for _, id := range ids {
		mf, err := quarantine.GetManifest(id)
		if err != nil || mf == nil {
			continue
		}
		totalFreed += mf.TotalSize
		cats := formatCLICategories(mf.Categories)
		fmt.Printf("  %s  %9s  %s\n", cliDate(mf.CreatedAt), util.FormatSize(mf.TotalSize), cats)
	}
	fmt.Printf("\n  All %s would be deleted, freeing %s. Run with %s to proceed.\n",
		pluralize(len(ids), "1 backup", fmt.Sprintf("%d backups", len(ids))),
		util.FormatSize(totalFreed),
		style.Yellowf("--yes"),
	)
	return nil
}
