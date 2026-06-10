package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("d"))) {
		m.selectedIdx = 0
		m.screen = ScreenCategories
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
		m.confirmMsg = buildConfirmMessage(m.categories, m.result)
		m.screen = ScreenConfirm
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	return m, nil
}

func (m model) viewDashboard() string {
	if m.result == nil {
		return m.spinner.View() + " " + i18n.T("scanning") + "\n"
	}

	var totalFiles int
	for _, c := range m.result.Categories {
		totalFiles += c.Files
	}

	barWidth := 32
	total := m.result.TotalSize

	bar := func(size int64, st lipgloss.Style) string {
		if total == 0 {
			return strings.Repeat(" ", barWidth)
		}
		blocks := int(size * int64(barWidth) / total)
		if blocks > barWidth {
			blocks = barWidth
		}
		if blocks == 0 && size > 0 {
			blocks = 1
		}
		empty := barWidth - blocks
		return st.Render(strings.Repeat("█", blocks)) + strings.Repeat(" ", empty)
	}

	body := m.appTitle(i18n.T("dashboard")) + "\n\n"
	body += fmt.Sprintf("  %s  ·  %d %s  ·  %d %s\n",
		valueStyle.Render(util.FormatSize(m.result.TotalSize)),
		totalFiles,
		i18n.T("files"),
		len(m.result.Categories),
		i18n.T("categories"),
	)
	body += "  " + mutedStyle.Render(strings.Repeat("─", 50)) + "\n\n"

	labels := []struct {
		name string
		size int64
		st   lipgloss.Style
	}{
		{i18n.T("safe"), m.result.SafeSize, safeStyle},
		{i18n.T("review"), m.result.ReviewSize, reviewStyle},
		{i18n.T("danger"), m.result.DangerSize, dangerStyle},
	}

	for _, l := range labels {
		pct := 0
		if total > 0 {
			pct = int(l.size * 100 / total)
		}
		body += fmt.Sprintf("  %-10s %s  %s (%d%%)\n", l.name, bar(l.size, l.st), util.FormatSize(l.size), pct)
	}

	body += "\n" + footer(keyHint("D", i18n.T("select_categories")), keyHint("Enter", i18n.T("confirm")), keyHint("Esc", i18n.T("back")))
	return body
}
