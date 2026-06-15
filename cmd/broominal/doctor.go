package main

import (
	"fmt"
	"os"

	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks",
		Run: func(cmd *cobra.Command, args []string) {
			checks := doctor.Run()
			var fail bool
			for _, c := range checks {
				var marker string
				switch c.Status {
				case doctor.StatusPass:
					marker = style.Passf(i18n.T("status_pass"))
				case doctor.StatusWarn:
					marker = style.Warnf(i18n.T("status_warn"))
				case doctor.StatusFail:
					marker = style.Failf(i18n.T("status_fail"))
					fail = true
				}
				fmt.Printf("%-24s %s  %s\n", style.Boldf(c.Name), marker, style.Grayf(c.Detail))
			}
			if fail {
				os.Exit(1)
			}
		},
	}
}
