package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/internal/tui"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/spf13/cobra"
)

var Version = "v1.6.0"

func init() {
	cobra.AddTemplateFunc("bold", func(s string) string { return style.Bold + s + style.Reset })
	cobra.AddTemplateFunc("green", func(s string) string { return style.Green + s + style.Reset })
	cobra.AddTemplateFunc("yellow", func(s string) string { return style.Yellow + s + style.Reset })
	cobra.AddTemplateFunc("red", func(s string) string { return style.Red + s + style.Reset })
	cobra.AddTemplateFunc("cyan", func(s string) string { return style.Cyan + s + style.Reset })
	cobra.AddTemplateFunc("gray", func(s string) string { return style.Gray + s + style.Reset })
}

const helpTemplate = `{{with (or .Long .Short)}}{{bold .}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

const usageTemplate = `{{bold "Usage:"}}
{{if .Runnable}}  {{cyan .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{cyan .CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{bold "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{bold "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{bold "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{cyan (rpad .Name .NamePadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{bold "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{bold "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{bold "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{gray "Use"}} {{cyan .CommandPath}} [command] --help {{gray "for more information about a command."}}{{end}}
`

func setupFileLogging() {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		return
	}
	logDir := filepath.Join(appData, "broominal")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	logFile := filepath.Join(logDir, "broominal.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	setupFileLogging()

	var jsonLogs bool
	var rootCmd = &cobra.Command{
		Use:   "broominal",
		Short: "Safe Windows cleanup with undo",
		Long:  "A safe, transparent, undoable Windows cleanup tool.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if jsonLogs {
				slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
			}
		},
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

	rootCmd.PersistentFlags().BoolVar(&jsonLogs, "json-logs", false, "Output structured JSON logs to stderr")
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(uiCmd())
	rootCmd.AddCommand(cleanCmd())
	rootCmd.AddCommand(restoreCmd())
	rootCmd.AddCommand(reportCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(quarantineCleanupCmd())
	rootCmd.AddCommand(pathCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func relaunch() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	time.Sleep(200 * time.Millisecond)
	_ = cmd.Run()
	if strings.HasSuffix(exe, ".old") {
		scheduleOldCleanup(exe)
	}
}

func scheduleOldCleanup(exePath string) {
	scriptPath := exePath + ".bat"
	script := fmt.Sprintf(
		"@ping -n 2 127.0.0.1 >nul\r\n@del /f /q \"%s\"\r\n@del /f /q \"%%~f0\"\r\n",
		exePath,
	)
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return
	}
	_ = exec.Command("cmd", "/c", "start", "/b", "", scriptPath).Start()
}
