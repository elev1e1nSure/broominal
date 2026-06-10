package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/pathman"
	"github.com/elev1e1nSure/broominal/pkg/style"
)

func (m model) handleKeyPathConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.pathConfirmIdx > 0 {
			m.pathConfirmIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.pathConfirmIdx < 1 {
			m.pathConfirmIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.pathConfirmIdx == 1 {
			// No — back to config
			m.screen = ScreenConfig
			m.selectedIdx = m.lastConfigIdx
			return m, nil
		}
		// Yes — execute
		var err error
		if m.pathOperation == "add" {
			err = pathman.AddToPath()
		} else {
			err = pathman.RemoveFromPath()
		}
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		if m.pathOperation == "add" {
			m.pathResultMsg = i18n.T("path_result_success_add")
		} else {
			m.pathResultMsg = i18n.T("path_result_success_remove")
		}
		m.screen = ScreenPathResult
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyPathResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "enter"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	return m, nil
}

func (m model) viewPathConfirm() string {
	exe, _ := os.Executable()
	exe = filepath.Clean(exe)

	var titleKey, descKey string
	if m.pathOperation == "add" {
		titleKey = "path_confirm_add"
		descKey = "path_confirm_desc_add"
	} else {
		titleKey = "path_confirm_remove"
		descKey = "path_confirm_desc_remove"
	}

	var body string
	body += m.appTitle(i18n.T(titleKey)) + "\n\n"
	body += "  " + mutedStyle.Render(fmt.Sprintf(i18n.T(descKey), exe)) + "\n\n"

	options := []string{
		style.Cyanf("[%s]", i18n.T("yes")),
		style.Grayf("[%s]", i18n.T("no")),
	}
	if m.pathConfirmIdx == 1 {
		options[0] = style.Grayf("[%s]", i18n.T("yes"))
		options[1] = style.Cyanf("[%s]", i18n.T("no"))
	}

	body += "  " + options[0] + "\n"
	body += "  " + options[1] + "\n\n"
	body += footer(
		keyHint("↑↓", i18n.T("toggle")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewPathResult() string {
	var body string
	body += m.appTitle(m.pathResultMsg) + "\n\n"
	body += "  " + mutedStyle.Render(i18n.T("path_result_restart_terminal")) + "\n\n"
	body += footer(
		keyHint("Enter/Esc", i18n.T("back")),
	)
	return body
}
