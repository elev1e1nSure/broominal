package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/pathman"
)

func (m model) handleKeyPathSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
		inPath, _ := pathman.IsInPath()
		var err error
		if inPath {
			err = pathman.RemoveFromPath()
		} else {
			err = pathman.AddToPath()
		}
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

func (m model) viewPathSettings() string {
	inPath, _ := pathman.IsInPath()

	status := "[OFF]"
	statusStyle := mutedStyle
	if inPath {
		status = "[ON]"
		statusStyle = safeStyle
	}

	content := i18n.T("path_user_status") + " " + statusStyle.Render(status) + "\n"

	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("Enter", i18n.T("toggle")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("config_path"), content, foot)
}
