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
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
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
	// Delete selected backup
	if key.Matches(msg, key.NewBinding(key.WithKeys("d"))) {
		if len(m.restoreEntries) == 0 || m.restoreIdx >= len(m.restoreEntries) {
			return m, nil
		}
		id := m.restoreEntries[m.restoreIdx].id
		_, err := quarantine.Delete(id)
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		// Reload list
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
				label:     mf.Label,
			})
		}
		m.restoreEntries = entries
		if m.restoreIdx >= len(m.restoreEntries) && len(m.restoreEntries) > 0 {
			m.restoreIdx = len(m.restoreEntries) - 1
		}
		return m, nil
	}
	// Delete all backups
	if key.Matches(msg, key.NewBinding(key.WithKeys("a"))) {
		if len(m.restoreEntries) == 0 {
			return m, nil
		}
		_, _, err := quarantine.CleanupAll()
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
		}
		m.restoreEntries = nil
		m.restoreIdx = 0
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
				label:     mf.Label,
			})
		}
		m.restoreEntries = entries
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

func (m model) viewRestore() string {
	var body string
	body += m.appTitle(i18n.T("restore")) + "\n\n"
	if len(m.restoreEntries) == 0 {
		body += mutedStyle.Render("  "+i18n.T("no_quarantines")) + "\n"
	} else {
		for i, e := range m.restoreEntries {
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
		keyHint("Enter", i18n.T("restore")),
		keyHint("D", i18n.T("delete")),
		keyHint("A", i18n.T("delete_all")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
