//go:build windows

package main

import "github.com/spf13/cobra"

func init() {
	cobra.MousetrapHelpText = ""
}
