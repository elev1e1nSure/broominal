package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
)

func (m model) handleKeyMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.selectedIdx < 4 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		switch m.selectedIdx {
		case 0: // Scan & Clean
			m.screen = ScreenDashboard
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				cfg, _ := config.Load()
				if cfg == nil {
					cfg = config.Default()
				}
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				res, err := scanner.ScanWithConfig(ctx, cfg)
				if err != nil {
					return errMsg{err}
				}
				return scanDoneMsg{res}
			})
		case 1: // Restore
			ids, err := quarantine.List()
			if err != nil {
				m.err = err
				m.screen = ScreenError
				return m, nil
			}
			var entries []restoreEntry
			for _, id := range ids {
				mf, _ := quarantine.GetManifest(id)
				if mf == nil {
					continue
				}
				entries = append(entries, restoreEntry{
					id:        mf.ID,
					createdAt: mf.CreatedAt,
					totalSize: mf.TotalSize,
					files:     mf.Files,
					label:     mf.Label,
				})
			}
			m.restoreEntries = entries
			m.restoreIdx = 0
			m.screen = ScreenRestore
			return m, nil
		case 2: // Doctor
			m.doctorChecks = doctor.Run()
			m.doctorQuarantineStats = doctor.QuarantineStats()
			m.screen = ScreenDoctor
			return m, nil
		case 3: // Settings
			cfg, err := config.Load()
			if err != nil {
				m.err = err
				m.screen = ScreenError
				return m, nil
			}
			m.configCfg = cfg
			m.selectedIdx = 0
			m.screen = ScreenConfig
			return m, nil
		}
	}
	return m, nil
}

func (m model) viewMainMenu() string {
	items := []string{
		i18n.T("menu_scan_clean"),
		i18n.T("menu_restore"),
		i18n.T("menu_doctor"),
		i18n.T("menu_settings"),
	}
	var body string
	body += titleStyle.Render(fmt.Sprintf("Broominal [v%s] %s", m.version, i18n.T("main_menu"))) + "\n\n"
	for i, item := range items {
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %s", item)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %s", item)) + "\n"
		}
	}
	body += "\n" + footer(keyHint("Enter", i18n.T("select")), keyHint("Q", i18n.T("quit")))
	return body
}
