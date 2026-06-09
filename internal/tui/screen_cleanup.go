package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyQuarantineCleanup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			deleted, freed, err := quarantine.CleanupAll()
			if err != nil {
				return errMsg{err}
			}
			return cleanDoneMsg{&types.CleanResult{
				RestoreID: fmt.Sprintf("Removed %d quarantines, freed %s", deleted, util.FormatSize(freed)),
				Freed:     freed,
				Files:     deleted,
			}, nil}
		})
	}
	return m, nil
}

func (m model) viewQuarantineCleanup() string {
	head := titleStyle.Render(i18n.T("quarantine_cleanup"))
	return head + "\n\n" +
		mutedStyle.Render("  "+i18n.T("cleanup_desc")) + "\n\n" +
		footer(
			keyHint("Enter", i18n.T("proceed")),
			keyHint("Esc", i18n.T("back")),
		)
}
