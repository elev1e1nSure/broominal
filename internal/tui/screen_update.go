package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

type installUpdateMsg struct {
	err error
}

func (m model) handleKeyUpdateAvailable(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "esc" {
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
	if s == "q" {
		return m, tea.Quit
	}
	if s == "enter" && m.updateAvailableRelease != nil {
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
	s := msg.String()
	if s == "esc" || s == "q" {
		m.updateCancelled = true
		if m.updateFromConfig {
			m.screen = ScreenConfig
			m.selectedIdx = m.lastConfigIdx
		} else {
			m.screen = ScreenMainMenu
			m.selectedIdx = m.lastMainMenuIdx
		}
		m.updateProgress = ""
		m.updateFromConfig = false
		if m.updateError != nil {
			m.updateError = nil
		}
		return m, nil
	}
	return m, nil
}

func (m model) viewUpdateAvailable() string {
	var content string
	if m.updateAvailableRelease != nil {
		content += fmt.Sprintf("  %s: %s\n", i18n.T("current_version"), mutedStyle.Render(m.version))
		content += fmt.Sprintf("  %s: %s\n\n", i18n.T("latest_version"), safeStyle.Render(m.updateAvailableRelease.TagName))
		content += mutedStyle.Render("  "+i18n.T("update_prompt")) + "\n"
	}

	var foot string
	if m.updateFromConfig {
		foot = footer(
			keyHint("Q", i18n.T("quit")),
			keyHint("Enter", i18n.T("yes_update")),
			keyHint("Esc", i18n.T("back")),
		)
	} else {
		foot = footer(
			keyHint("Q", i18n.T("quit")),
			keyHint("Enter", i18n.T("yes_update")),
			keyHint("Esc", i18n.T("skip")),
		)
	}
	return m.appFrame(i18n.T("update_available"), content, foot)
}

func (m model) handleKeyNoUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "enter" || s == "esc" {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if s == "q" {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) viewNoUpdate() string {
	content := "  " + safeStyle.Render(i18n.T("no_update_desc")) + "\n"
	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("Enter", i18n.T("back")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("no_update"), content, foot)
}

func (m model) viewUpdating() string {
	var content string
	var foot string
	if m.updateError != nil {
		content += fmt.Sprintf("  %s\n", m.updateProgress)
		content += "\n" + dangerStyle.Render(fmt.Sprintf("  [ERROR] %v", m.updateError)) + "\n"
		foot = footer(keyHint("Q", i18n.T("quit")), keyHint("Esc", i18n.T("back")))
	} else {
		width := 40
		pos := m.updateTick % (width * 2)
		if pos >= width {
			pos = width*2 - 1 - pos
		}
		fraction := float64(pos) / float64(width-1)
		bar := m.progress.ViewAs(fraction)
		statusLine := lipgloss.NewStyle().Width(m.progress.Width).Align(lipgloss.Left).Render(mutedStyle.Render(m.updateProgress))
		content += fmt.Sprintf("\n%s\n\n%s\n", bar, statusLine)
		foot = footer(keyHint("Esc", i18n.T("cancel")))
	}
	return m.appFrame(i18n.T("updating"), content, foot)
}
