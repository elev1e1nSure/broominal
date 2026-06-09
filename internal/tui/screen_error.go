package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyError(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		m.err = nil
		return m, nil
	}
	return m, nil
}

func (m model) viewError() string {
	return m.appTitle(i18n.T("error")) + "\n\n" +
		dangerStyle.Render(fmt.Sprintf("  %v", m.err)) + "\n\n" +
		footer(
			keyHint("Esc", i18n.T("back")),
		)
}
