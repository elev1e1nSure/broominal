package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

type installUpdateMsg struct {
	err error
}

func (m model) handleKeyUpdateAvailable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		m.updateAvailableRelease = nil
		m.updateError = nil
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("y", "enter"))) {
		if m.updateAvailableRelease != nil {
			m.screen = ScreenUpdating
			m.updateProgress = i18n.T("downloading_update")
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				path, err := update.DownloadUpdate(m.updateAvailableRelease)
				return downloadUpdateMsg{path, err}
			})
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("n"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		m.updateAvailableRelease = nil
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyUpdating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// No user input allowed during update
	return m, nil
}

func (m model) viewUpdateAvailable() string {
	var body string
	body += titleStyle.Render(i18n.T("update_available")) + "\n\n"

	if m.updateAvailableRelease != nil {
		body += fmt.Sprintf("  %s: %s\n", i18n.T("current_version"), mutedStyle.Render("v"+m.version))
		body += fmt.Sprintf("  %s: %s\n\n", i18n.T("latest_version"), safeStyle.Render(m.updateAvailableRelease.TagName))
		body += mutedStyle.Render("  "+i18n.T("update_prompt")) + "\n\n"
	}

	body += footer(
		keyHint("Y", i18n.T("yes_update")),
		keyHint("N", i18n.T("no")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewUpdating() string {
	var body string
	body += titleStyle.Render(i18n.T("updating")) + "\n\n"
	body += fmt.Sprintf("  %s %s\n", m.spinner.View(), m.updateProgress)
	if m.updateError != nil {
		body += "\n" + dangerStyle.Render(fmt.Sprintf("  [ERROR] %v", m.updateError)) + "\n"
		body += "\n" + footer(keyHint("Esc", i18n.T("back")))
	} else {
		body += "\n" + mutedStyle.Render("  "+i18n.T("please_wait")) + "\n"
	}
	return body
}
