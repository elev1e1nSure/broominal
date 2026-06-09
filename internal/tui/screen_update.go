package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
		} else {
			m.screen = ScreenMainMenu
		}
		m.selectedIdx = 0
		m.updateAvailableRelease = nil
		m.updateError = nil
		return m, nil
	}
	if s == "q" || s == "m" {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		m.updateAvailableRelease = nil
		m.updateError = nil
		return m, nil
	}
	if (s == "y" || s == "enter") && m.updateAvailableRelease != nil {
		m.screen = ScreenUpdating
		m.updateProgress = i18n.T("downloading_update")
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			path, err := update.DownloadUpdate(m.updateAvailableRelease)
			return downloadUpdateMsg{path, err}
		})
	}
	if s == "n" {
		if m.updateFromConfig {
			m.screen = ScreenConfig
		} else {
			m.screen = ScreenMainMenu
		}
		m.selectedIdx = 0
		m.updateAvailableRelease = nil
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyUpdating(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if m.updateProgress == i18n.T("update_complete_restart") && (s == "enter" || s == "y") {
		exePath, _ := os.Executable()
		if strings.HasSuffix(exePath, ".old") {
			exePath = strings.TrimSuffix(exePath, ".old")
		}
		_ = exec.Command("cmd", "/c", "start", "", exePath, "ui").Start()
		return m, tea.Quit
	}
	if m.updateError != nil {
		if s == "esc" {
			if m.updateFromConfig {
				m.screen = ScreenConfig
			} else {
				m.screen = ScreenMainMenu
			}
			m.selectedIdx = 0
			m.updateError = nil
			m.updateProgress = ""
			return m, nil
		}
		if s == "q" || s == "m" {
			m.screen = ScreenMainMenu
			m.selectedIdx = 0
			m.updateError = nil
			m.updateProgress = ""
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

	body += footer(
		keyHint("Y", i18n.T("yes_update")),
		keyHint("N", i18n.T("no")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) handleKeyNoUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "enter" || s == "esc" {
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	if s == "q" || s == "m" {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
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
	if m.updateProgress == i18n.T("update_complete_restart") {
		body += fmt.Sprintf("  %s\n", m.updateProgress)
		body += "\n" + footer(keyHint("Enter", i18n.T("restart")))
	} else {
		body += fmt.Sprintf("  %s %s\n", m.spinner.View(), m.updateProgress)
		if m.updateError != nil {
			body += "\n" + dangerStyle.Render(fmt.Sprintf("  [ERROR] %v", m.updateError)) + "\n"
			body += "\n" + footer(keyHint("Esc", i18n.T("back")))
		} else {
			body += "\n" + mutedStyle.Render("  "+i18n.T("please_wait")) + "\n"
		}
	}
	return body
}
