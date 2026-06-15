package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/internal/tui"
	"github.com/spf13/cobra"
)

func uiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch interactive TUI",
		Run: func(cmd *cobra.Command, args []string) {
			restart, err := tui.Start(Version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
				os.Exit(1)
			}
			if restart {
				relaunch()
			}
		},
	}
}
