package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/pathman"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func pathCmd() *cobra.Command {
	var add bool
	var remove bool
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Add or remove broominal from user PATH",
		Run: func(cmd *cobra.Command, args []string) {
			if add && remove {
				fmt.Fprintln(os.Stderr, style.Failf("[FAIL]"), "Use --add or --remove, not both.")
				os.Exit(1)
			}
			if !add && !remove {
				inPath, err := pathman.IsInPath()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s %v\n", style.Failf("[FAIL]"), err)
					os.Exit(1)
				}
				if inPath {
					fmt.Printf("%s %s\n", style.Passf("[PASS]"), style.Grayf("broominal is already in PATH."))
				} else {
					fmt.Printf("%s %s\n", style.Warnf("[WARN]"), style.Grayf("broominal is not in PATH. Use --add to add."))
				}
				return
			}
			if add {
				err := pathman.AddToPath()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s %v\n", style.Failf("[FAIL]"), err)
					os.Exit(1)
				}
				fmt.Printf("%s %s\n", style.Greenf("Added to PATH."), style.Grayf("Restart your terminal to apply."))
				return
			}
			if remove {
				err := pathman.RemoveFromPath()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s %v\n", style.Failf("[FAIL]"), err)
					os.Exit(1)
				}
				fmt.Printf("%s %s\n", style.Greenf("Removed from PATH."), style.Grayf("Restart your terminal to apply."))
				return
			}
		},
	}
	cmd.Flags().BoolVar(&add, "add", false, "Add broominal directory to user PATH")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove broominal directory from user PATH")
	return cmd
}
