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
	if key.Matches(msg, key.NewBinding(key.WithKeys("left", "h", "right", "l"))) {
		m.cleanupOldOnly = !m.cleanupOldOnly
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenQuarantineCleaning
		oldOnly := m.cleanupOldOnly
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			var deleted int
			var freed int64
			var err error
			if oldOnly {
				deleted, freed, err = quarantine.Cleanup(30)
			} else {
				deleted, freed, err = quarantine.CleanupAll()
			}
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
	modeLabel := i18n.T("all_quarantines")
	if m.cleanupOldOnly {
		modeLabel = i18n.T("old_only")
	}
	head := m.appTitle(i18n.T("quarantine_cleanup"))
	return head + "\n\n" +
		mutedStyle.Render("  "+i18n.T("cleanup_desc")) + "\n\n" +
		fmt.Sprintf("  %s: %s\n", i18n.T("mode"), modeLabel) + "\n" +
		footer(
			keyHint("←/→", i18n.T("toggle")),
			keyHint("Enter", i18n.T("confirm")),
			keyHint("Esc", i18n.T("back")),
		)
}

func (m model) viewQuarantineCleaning() string {
	return m.appTitle(i18n.T("quarantine_cleanup")) + "\n\n" +
		fmt.Sprintf("  %s %s\n", m.spinner.View(), i18n.T("cleaning_quarantines")) +
		mutedStyle.Render("  "+i18n.T("please_wait"))
}
