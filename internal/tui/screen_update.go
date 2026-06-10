package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

type installUpdateMsg struct {
	err error
}

func (m model) handleKeyUpdateAvailable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "esc" || s == "q" {
		if m.updateFromConfig {
			m.screen = ScreenConfig
			m.selectedIdx = m.lastConfigIdx
		} else {
			m.screen = ScreenMainMenu
			m.selectedIdx = m.lastMainMenuIdx
		}
		m.updateAvailableRelease = nil
		m.updateError = nil
		m.updateFromConfig = false
		return m, nil
	}
	if (s == "y" || s == "enter") && m.updateAvailableRelease != nil {
		m.screen = ScreenUpdating
		m.updateProgress = i18n.T("downloading_update")
		m.updateFromConfig = false
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			path, err := update.DownloadUpdate(m.updateAvailableRelease)
			return downloadUpdateMsg{path, err}
		})
	}
	return m, nil
}

func (m model) handleKeyUpdating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.updateError != nil {
		s := msg.String()
		if s == "esc" || s == "q" {
			if m.updateFromConfig {
				m.screen = ScreenConfig
				m.selectedIdx = m.lastConfigIdx
			} else {
				m.screen = ScreenMainMenu
				m.selectedIdx = m.lastMainMenuIdx
			}
			m.updateError = nil
			m.updateProgress = ""
			m.updateFromConfig = false
			return m, nil
		}
	}
	return m, nil
}

func (m model) viewUpdateAvailable() string {
	var body string
	body += m.appTitle(i18n.T("update_available")) + "\n\n"

	if m.updateAvailableRelease != nil {
		body += fmt.Sprintf("  %s: %s\n", i18n.T("current_version"), mutedStyle.Render(m.version))
		body += fmt.Sprintf("  %s: %s\n\n", i18n.T("latest_version"), safeStyle.Render(m.updateAvailableRelease.TagName))
		body += mutedStyle.Render("  "+i18n.T("update_prompt")) + "\n\n"
	}

	if m.updateFromConfig {
		body += footer(
			keyHint("Enter", i18n.T("yes_update")),
			keyHint("Esc", i18n.T("back")),
		)
	} else {
		body += footer(
			keyHint("Enter", i18n.T("yes_update")),
			keyHint("Esc", i18n.T("skip")),
		)
	}
	return body
}

func (m model) handleKeyNoUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "enter" || s == "esc" {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if s == "q" {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	return m, nil
}

func (m model) viewNoUpdate() string {
	var body string
	body += m.appTitle(i18n.T("no_update")) + "\n\n"
	body += "  " + safeStyle.Render(i18n.T("no_update_desc")) + "\n\n"
	body += footer(keyHint("Enter", i18n.T("back")))
	return body
}

func (m model) viewUpdating() string {
	var body string
	body += m.appTitle(i18n.T("updating")) + "\n\n"
	if m.updateError != nil {
		body += fmt.Sprintf("  %s\n", m.updateProgress)
		body += "\n" + dangerStyle.Render(fmt.Sprintf("  [ERROR] %v", m.updateError)) + "\n"
		body += "\n" + footer(keyHint("Esc", i18n.T("back")))
	} else {
		body += fmt.Sprintf("  %s %s\n", m.spinner.View(), m.updateProgress)
		body += "\n" + mutedStyle.Render("  "+i18n.T("please_wait")) + "\n"
	}
	return body
}
