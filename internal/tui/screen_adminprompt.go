package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/style"
)

func (m model) handleKeyAdminPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.adminPromptIdx > 0 {
			m.adminPromptIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.adminPromptIdx < 1 {
			m.adminPromptIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.adminPromptIdx == 0 {
			// Restart as administrator
			_, _ = doctor.Fix("admin")
			return m, tea.Quit
		}
		// Exit
		return m, tea.Quit
	}
	return m, nil
}

func (m model) viewAdminPrompt() string {
	var body string
	body += m.appTitle(i18n.T("admin_required")) + "\n\n"
	body += "  " + mutedStyle.Render(i18n.T("admin_required_desc")) + "\n\n"

	options := []string{
		style.Cyanf("[%s]", i18n.T("restart_as_admin")),
		style.Grayf("[%s]", i18n.T("exit")),
	}
	if m.adminPromptIdx == 1 {
		options[0] = style.Grayf("[%s]", i18n.T("restart_as_admin"))
		options[1] = style.Cyanf("[%s]", i18n.T("exit"))
	}

	body += "  " + options[0] + "\n"
	body += "  " + options[1] + "\n\n"
	body += footer(
		keyHint("↑↓", i18n.T("toggle")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("Esc", i18n.T("exit")),
	)
	return body
}
