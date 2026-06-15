package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyRestore(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.restoreIdx > 0 {
			m.restoreIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.restoreIdx < len(m.restoreEntries)-1 {
			m.restoreIdx++
		}
		return m, nil
	}
	// Delete selected — go to confirmation screen
	if key.Matches(msg, key.NewBinding(key.WithKeys("x"))) {
		if len(m.restoreEntries) == 0 || m.restoreIdx >= len(m.restoreEntries) {
			return m, nil
		}
		m.screen = ScreenConfirmDeleteQuarantine
		return m, nil
	}
	// Delete all — go to confirmation screen
	if key.Matches(msg, key.NewBinding(key.WithKeys("a"))) {
		if len(m.restoreEntries) == 0 {
			return m, nil
		}
		m.screen = ScreenConfirmDeleteAllQuarantine
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if len(m.restoreEntries) == 0 {
			return m, nil
		}
		id := m.restoreEntries[m.restoreIdx].id
		restored, skipped, err := quarantine.Restore(id, false)
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		m.restoreResult = fmt.Sprintf(i18n.T("restored_n_skipped"), restored, skipped)
		m.restoreEntries = reloadEntries()
		if m.restoreIdx >= len(m.restoreEntries) {
			m.restoreIdx = len(m.restoreEntries) - 1
			if m.restoreIdx < 0 {
				m.restoreIdx = 0
			}
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyConfirmDeleteQuarantine(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenRestore
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		id := m.restoreEntries[m.restoreIdx].id
		m.deleteAllQuarantine = false
		m.screen = ScreenDeletingQuarantine
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			freed, err := quarantine.Delete(id)
			if err != nil && util.IsFileLocked(err) {
				err = fmt.Errorf("%s", i18n.T("quarantine_locked"))
			}
			return quarantineDeleteDoneMsg{count: 1, freed: freed, err: err}
		})
	}
	return m, nil
}

func (m model) handleKeyConfirmDeleteAllQuarantine(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenRestore
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.deleteAllQuarantine = true
		m.screen = ScreenDeletingQuarantine
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			count, freed, err := quarantine.CleanupAll()
			if err != nil && util.IsFileLocked(err) {
				err = fmt.Errorf("%s", i18n.T("quarantine_locked"))
			}
			return quarantineDeleteDoneMsg{count: count, freed: freed, err: err}
		})
	}
	return m, nil
}

func (m model) handleKeyDeletingQuarantine(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) viewDeletingQuarantine() string {
	label := i18n.T("cleaning_quarantines")
	if !m.deleteAllQuarantine && m.restoreIdx < len(m.restoreEntries) {
		label = fmt.Sprintf("%s %s", i18n.T("cleaning_quarantines"), m.restoreEntries[m.restoreIdx].id)
	}
	return m.appTitle(i18n.T("cleaning_quarantines")) + "\n\n" +
		fmt.Sprintf("  %s %s\n", m.spinner.View(), label) +
		mutedStyle.Render("  "+i18n.T("please_wait")) + "\n"
}

func (m model) viewRestore() string {
	var body string
	body += m.appTitle(i18n.T("restore")) + "\n\n"
	if len(m.restoreEntries) == 0 {
		body += mutedStyle.Render("  "+i18n.T("no_quarantines")) + "\n"
	} else {
		visible := m.height - 8
		if visible < 5 {
			visible = 5
		}
		start, end := clampWindow(m.restoreIdx, len(m.restoreEntries), visible)
		for i := start; i < end; i++ {
			e := m.restoreEntries[i]
			dateStr := e.createdAt.Format("2006-01-02 15:04")
			line := fmt.Sprintf("%s  %s  %s, %d files", e.id, dateStr, util.FormatSize(e.totalSize), e.files)
			if i == m.restoreIdx {
				body += selectedStyle.Render(fmt.Sprintf("> %s", line)) + "\n"
			} else {
				body += mutedStyle.Render(fmt.Sprintf("  %s", line)) + "\n"
			}
		}
	}
	if m.restoreResult != "" {
		body += "\n" + safeStyle.Render("  [OK] "+m.restoreResult) + "\n"
	}
	body += "\n" + footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("Enter", i18n.T("restore")),
		keyHint("X", i18n.T("delete")),
		keyHint("A", i18n.T("delete_all")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewConfirmDeleteQuarantine() string {
	var entry string
	if m.restoreIdx < len(m.restoreEntries) {
		e := m.restoreEntries[m.restoreIdx]
		entry = fmt.Sprintf("%s  %s  %s", e.id, e.createdAt.Format("2006-01-02 15:04"), util.FormatSize(e.totalSize))
	}
	return m.appTitle(i18n.T("warning")) + "\n\n" +
		reviewStyle.Render("  "+i18n.T("confirm_delete_one")) + "\n" +
		mutedStyle.Render("  "+entry) + "\n\n" +
		footer(
			keyHint("Q", i18n.T("quit")),
			keyHint("Enter", i18n.T("confirm")),
			keyHint("Esc", i18n.T("back")),
		)
}

func (m model) viewConfirmDeleteAllQuarantine() string {
	msg := fmt.Sprintf(i18n.T("confirm_delete_all"), len(m.restoreEntries))
	return m.appTitle(i18n.T("warning")) + "\n\n" +
		dangerStyle.Render("  "+msg) + "\n\n" +
		footer(
			keyHint("Q", i18n.T("quit")),
			keyHint("Enter", i18n.T("confirm")),
			keyHint("Esc", i18n.T("back")),
		)
}

// reloadEntries returns an up-to-date list of quarantine entries.
func reloadEntries() []restoreEntry {
	ids, _ := quarantine.List()
	var entries []restoreEntry
	for _, rid := range ids {
		mf, _ := quarantine.GetManifest(rid)
		if mf == nil {
			continue
		}
		entries = append(entries, restoreEntry{
			id:        mf.ID,
			createdAt: mf.CreatedAt,
			totalSize: mf.TotalSize,
			files:     mf.Files,
		})
	}
	return entries
}
