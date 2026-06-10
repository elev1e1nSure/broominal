package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		m.cleanResult = nil
		m.restoreResult = ""
		m.result = nil
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("r"))) {
		if m.cleanResult == nil {
			return m, nil
		}
		conflicts, err := quarantine.CheckRestoreConflicts(m.cleanResult.RestoreID)
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		if len(conflicts) > 0 {
			m.conflicts = conflicts
			m.restoreForceOverwrite = false
			m.screen = ScreenRestoreConflict
			return m, nil
		}
		_, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, false)
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		if skipped == 0 {
			m.cleanResult = nil // restored
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyRestoreConflict(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("o"))) {
		if m.cleanResult != nil {
			restored, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, true)
			if err != nil {
				m.err = err
				m.screen = ScreenError
				return m, nil
			}
			_ = restored
			_ = skipped
			m.cleanResult = nil // restored
		}
		m.screen = ScreenResult
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("s"))) {
		if m.cleanResult != nil {
			restored, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, false)
			if err != nil {
				m.err = err
				m.screen = ScreenError
				return m, nil
			}
			_ = restored
			if skipped == 0 {
				m.cleanResult = nil
			}
		}
		m.screen = ScreenResult
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenResult
		return m, nil
	}
	return m, nil
}

func (m model) viewCleaning() string {
	return m.appTitle(i18n.T("cleaning")) + "\n\n" +
		fmt.Sprintf("  %s %s\n", m.spinner.View(), i18n.T("moving_files")) +
		mutedStyle.Render("  "+i18n.T("please_wait"))
}

func (m model) viewResult() string {
	if m.cleanResult == nil {
		return m.appTitle(i18n.T("restored")) + "\n\n" +
			safeStyle.Render("  [OK] "+i18n.T("hint_restored")) + "\n\n" +
			footer(keyHint("Esc", i18n.T("back")))
	}
	body := m.appTitle(i18n.T("cleanup_complete")) + "\n\n" +
		fmt.Sprintf("  Freed:      %s\n", safeStyle.Render(util.FormatSize(m.cleanResult.Freed))) +
		fmt.Sprintf("  Files:      %s\n", valueStyle.Render(fmt.Sprintf("%d", m.cleanResult.Files)))
	if m.cleanResult.Skipped > 0 {
		body += fmt.Sprintf("  Skipped:    %s\n", reviewStyle.Render(fmt.Sprintf("%d", m.cleanResult.Skipped)))
	}
	body += fmt.Sprintf("  Restore ID: %s\n", mutedStyle.Render(m.cleanResult.RestoreID))
	body += "\n" + footer(
		keyHint("R", i18n.T("restore_last")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewRestoreConflict() string {
	var body string
	body += m.appTitle(i18n.T("restore_conflicts")) + "\n\n"
	body += dangerStyle.Render(fmt.Sprintf("  %d %s:", len(m.conflicts), i18n.T("files_already_exist"))) + "\n"
	for _, p := range m.conflicts {
		body += mutedStyle.Render("    "+p) + "\n"
	}
	body += "\n" + footer(
		keyHint("O", i18n.T("overwrite_all")),
		keyHint("S", i18n.T("skip_all")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
