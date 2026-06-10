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
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenQuarantineCleaning
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			deleted, freed, err := quarantine.CleanupAll()
			if err != nil {
				if util.IsFileLocked(err) {
					err = fmt.Errorf("%s", i18n.T("quarantine_locked"))
				}
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
	head := m.appTitle(i18n.T("quarantine_cleanup"))
	return head + "\n\n" +
		mutedStyle.Render("  "+i18n.T("cleanup_desc")) + "\n\n" +
		footer()
}

func (m model) viewQuarantineCleaning() string {
	return m.appTitle(i18n.T("quarantine_cleanup")) + "\n\n" +
		fmt.Sprintf("  %s %s\n", m.spinner.View(), i18n.T("cleaning_quarantines")) +
		mutedStyle.Render("  "+i18n.T("please_wait"))
}
