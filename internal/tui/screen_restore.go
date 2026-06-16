package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
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
		if len(m.restoreEntries) == 0 || m.restoreIdx >= len(m.restoreEntries) {
			return m, nil
		}
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

		const timeW = 16
		const sizeW = 9
		for i := start; i < end; i++ {
			e := m.restoreEntries[i]

			rowStyle := mutedStyle
			prefix := "  "
			if i == m.restoreIdx {
				rowStyle = selectedStyle
				prefix = selectedStyle.Render("> ")
			}

			timeSt := lipgloss.NewStyle().Width(timeW).Inherit(rowStyle)
			sizeSt := lipgloss.NewStyle().Width(sizeW).Align(lipgloss.Right).Inherit(rowStyle)

			catMaxW := m.width - 2 - timeW - 2 - sizeW - 2
			if catMaxW < 20 {
				catMaxW = 20
			}

			line := timeSt.Render(quarantineDate(e.createdAt)) +
				"  " +
				sizeSt.Render(util.FormatSize(e.totalSize)) +
				"  " +
				rowStyle.Render(formatQuarantineCategories(e.categories))

			body += prefix + line + "\n"
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
		cats := formatQuarantineCategories(e.categories)
		entry = quarantineDate(e.createdAt) + "   " + util.FormatSize(e.totalSize)
		if cats != "" {
			entry += "   " + cats
		}
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
			id:         mf.ID,
			createdAt:  mf.CreatedAt,
			totalSize:  mf.TotalSize,
			files:      mf.Files,
			categories: mf.Categories,
		})
	}
	return entries
}

// quarantineDate formats a quarantine timestamp as a human-readable relative date.
func quarantineDate(t time.Time) string {
	now := time.Now()
	ty, tm, td := t.Date()
	ny, nm, nd := now.Date()
	if ty == ny && tm == nm && td == nd {
		return i18n.T("today") + " " + t.Format("15:04")
	}
	py, pm, pd := now.AddDate(0, 0, -1).Date()
	if ty == py && tm == pm && td == pd {
		return i18n.T("yesterday") + " " + t.Format("15:04")
	}
	if t.Year() == now.Year() {
		return t.Format("02.01, 15:04")
	}
	return t.Format("02.01.06, 15:04")
}

// formatQuarantineCategories returns translated category names joined by " · ",
// capped at 3 with "+N" suffix for the rest.
func formatQuarantineCategories(cats []string) string {
	if len(cats) == 0 {
		return ""
	}
	const maxShow = 3
	shown := cats
	extra := 0
	if len(cats) > maxShow {
		shown = cats[:maxShow]
		extra = len(cats) - maxShow
	}
	names := make([]string, len(shown))
	for i, c := range shown {
		names[i] = i18n.CategoryName(c)
	}
	result := strings.Join(names, " · ")
	if extra > 0 {
		result += fmt.Sprintf(" +%d", extra)
	}
	return result
}
