package main

import (
	"fmt"

	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/update"
	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for updates and install if available",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Checking for updates... ")
			release, err := update.CheckForUpdates(Version)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("check failed: %w", err)
			}
			if release == nil {
				fmt.Println(style.Passf("up to date"))
				return nil
			}
			fmt.Println(style.Cyanf(release.TagName + " available"))

			fmt.Print("Downloading... ")
			path, err := update.DownloadUpdate(release)
			if err != nil {
				fmt.Println()
				return fmt.Errorf("download failed: %w", err)
			}
			fmt.Println(style.Passf("done"))

			fmt.Print("Installing... ")
			if err := update.InstallUpdate(path); err != nil {
				fmt.Println()
				return fmt.Errorf("install failed: %w", err)
			}
			fmt.Println(style.Passf("done"))
			fmt.Printf("Updated to %s — restart to apply.\n", style.Boldf(release.TagName))
			return nil
		},
	}
}
