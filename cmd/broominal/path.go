package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/pathman"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func pathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Manage broominal in the system PATH",
		Long: `Add, remove, or check whether broominal is in the user PATH.

Changes take effect after restarting the terminal.`,
	}
	cmd.AddCommand(pathStatusCmd())
	cmd.AddCommand(pathAddCmd())
	cmd.AddCommand(pathRemoveCmd())
	return cmd
}

func pathStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether broominal is in PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			inPath, err := pathman.IsInPath()
			if err != nil {
				return fmt.Errorf("check PATH: %w", err)
			}
			if inPath {
				fmt.Printf("  %s broominal is in PATH.\n", style.Passf("[OK]"))
			} else {
				fmt.Printf("  %s broominal is not in PATH. Run: broominal path add\n", style.Warnf("[WARN]"))
			}
			return nil
		},
	}
}

func pathAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add broominal directory to user PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pathman.AddToPath(); err != nil {
				fmt.Fprintf(os.Stderr, "  %s %v\n", style.Failf("[FAIL]"), err)
				os.Exit(1)
			}
			fmt.Printf("  %s Added to PATH. Restart your terminal to apply.\n", style.Passf("[OK]"))
			return nil
		},
	}
}

func pathRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove broominal directory from user PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pathman.RemoveFromPath(); err != nil {
				fmt.Fprintf(os.Stderr, "  %s %v\n", style.Failf("[FAIL]"), err)
				os.Exit(1)
			}
			fmt.Printf("  %s Removed from PATH. Restart your terminal to apply.\n", style.Passf("[OK]"))
			return nil
		},
	}
}
