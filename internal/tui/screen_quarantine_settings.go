package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/taskscheduler"
)

var autoCleanupCycles = []int{0, 7, 14, 30}

func (m model) handleKeyQuarantineSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		m.quarantineSettingsMsg = ""
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.configCfg != nil && m.configCfg.QuarantineEnabled && m.selectedIdx < 1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.configCfg == nil {
			return m, nil
		}
		switch m.selectedIdx {
		case 0: // Toggle quarantine on/off
			m.configCfg.QuarantineEnabled = !m.configCfg.QuarantineEnabled
			_ = config.Save(m.configCfg)
			if !m.configCfg.QuarantineEnabled {
				m.selectedIdx = 0
			}
			m.quarantineSettingsMsg = ""
		case 1: // Cycle auto-cleanup schedule
			if !m.configCfg.QuarantineEnabled {
				return m, nil
			}
			next := nextAutoCleanupDays(m.configCfg.QuarantineAutoCleanupDays)
			m.configCfg.QuarantineAutoCleanupDays = next
			_ = config.Save(m.configCfg)
			if next == 0 {
				if err := taskscheduler.Delete(); err != nil {
					m.quarantineSettingsMsg = i18n.T("quarantine_task_fail")
				} else {
					m.quarantineSettingsMsg = i18n.T("quarantine_task_removed")
				}
			} else {
				if err := taskscheduler.Set(next); err != nil {
					m.quarantineSettingsMsg = i18n.T("quarantine_task_fail")
				} else {
					m.quarantineSettingsMsg = i18n.T("quarantine_task_set")
				}
			}
		}
	}
	return m, nil
}

func nextAutoCleanupDays(current int) int {
	for i, v := range autoCleanupCycles {
		if v == current {
			return autoCleanupCycles[(i+1)%len(autoCleanupCycles)]
		}
	}
	return autoCleanupCycles[0]
}

func autoCleanupLabel(days int) string {
	switch days {
	case 7:
		return "[" + i18n.T("quarantine_cleanup_7d") + "]"
	case 14:
		return "[" + i18n.T("quarantine_cleanup_14d") + "]"
	case 30:
		return "[" + i18n.T("quarantine_cleanup_30d") + "]"
	default:
		return "[" + i18n.T("quarantine_cleanup_off") + "]"
	}
}

func (m model) viewQuarantineSettings() string {
	quarantineStatus := safeStyle.Render("[ON]")
	quarantineText := i18n.T("quarantine_on")
	quarantineWarn := ""
	if m.configCfg != nil && !m.configCfg.QuarantineEnabled {
		quarantineStatus = dangerStyle.Render("[OFF]")
		quarantineText = i18n.T("quarantine_off")
		quarantineWarn = dangerStyle.Render("  ! " + i18n.T("quarantine_disabled_warn"))
	}
	quarantineLabel := quarantineText + "  " + quarantineStatus

	autoCleanupDays := 0
	quarantineEnabled := m.configCfg == nil || m.configCfg.QuarantineEnabled
	if m.configCfg != nil {
		autoCleanupDays = m.configCfg.QuarantineAutoCleanupDays
	}
	autoLabel := i18n.T("quarantine_auto_cleanup") + "  " + mutedStyle.Render(autoCleanupLabel(autoCleanupDays))
	if quarantineEnabled && autoCleanupDays > 0 {
		autoLabel = i18n.T("quarantine_auto_cleanup") + "  " + safeStyle.Render(autoCleanupLabel(autoCleanupDays))
	}
	if !quarantineEnabled {
		autoLabel = mutedStyle.Render(i18n.T("quarantine_auto_cleanup") + "  " + autoCleanupLabel(autoCleanupDays))
	}

	items := []string{quarantineLabel, autoLabel}

	var body string
	body += m.appTitle(i18n.T("quarantine_settings")) + "\n\n"
	for i, item := range items {
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %s", item)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %s", item)) + "\n"
		}
		if i == 0 && quarantineWarn != "" {
			body += quarantineWarn + "\n"
		}
	}
	if m.quarantineSettingsMsg != "" {
		msgStyle := safeStyle
		if m.quarantineSettingsMsg == i18n.T("quarantine_task_fail") {
			msgStyle = dangerStyle
		}
		body += "\n" + msgStyle.Render("  "+m.quarantineSettingsMsg) + "\n"
	}
	body += "\n" + footer()
	return body
}
