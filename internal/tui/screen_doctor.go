package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyDoctor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m model) viewDoctor() string {
	var body string
	body += titleStyle.Render(i18n.T("doctor")) + "\n\n"
	for _, c := range m.doctorChecks {
		var marker string
		switch c.Status {
		case doctor.StatusPass:
			marker = safeStyle.Render("[PASS]")
		case doctor.StatusWarn:
			marker = reviewStyle.Render("[WARN]")
		case doctor.StatusFail:
			marker = dangerStyle.Render("[FAIL]")
		}
		body += fmt.Sprintf("  %-28s %s  %s\n", c.Name, marker, mutedStyle.Render(c.Detail))
	}
	body += "\n" + footer(
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
