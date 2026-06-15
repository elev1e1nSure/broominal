package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
				os.Exit(1)
			}
			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("%s %s\n\n%s\n", style.Boldf("Config path:"), style.Cyanf(config.Path()), string(data))
		},
	}
}
