package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
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
		m.restoreResult = fmt.Sprintf("Restored %d files (%d skipped)", restored, skipped)
		// Refresh list
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
	body += titleStyle.Render(i18n.T("restore")) + "\n\n"
	if len(m.restoreEntries) == 0 {
		body += mutedStyle.Render("  " + i18n.T("no_quarantines")) + "\n"
	} else {
		for i, entry := range m.restoreEntries {
			label := entry.id
			if entry.label != "" {
				label = fmt.Sprintf("%s (%s, %d files)", entry.id, util.FormatSize(entry.totalSize), entry.files)
			}
			if i == m.restoreIdx {
				body += selectedStyle.Render(fmt.Sprintf("> %s", label)) + "\n"
			} else {
				body += mutedStyle.Render(fmt.Sprintf("  %s", label)) + "\n"
			}
		}
	}
	if m.restoreResult != "" {
		body += "\n" + safeStyle.Render("  [OK] " + m.restoreResult) + "\n"
	}
	body += "\n" + footer(
		keyHint("Enter", i18n.T("restore")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}