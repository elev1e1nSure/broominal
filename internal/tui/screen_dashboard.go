package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
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
		if m.scanTotal > 0 {
			return m.appTitle(i18n.T("scanning")) + "\n\n  " +
				m.spinner.View() + " " +
				fmt.Sprintf("%s  %d / %d", i18n.T("scanning"), m.scanCompleted, m.scanTotal) + "\n"
		}
		return m.spinner.View() + " " + i18n.T("scanning") + "\n"
	}

	var totalFiles int
	for _, c := range m.result.Categories {
		totalFiles += c.Files
	}

	cats := make([]types.CategorySummary, len(m.result.Categories))
	copy(cats, m.result.Categories)
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Size > cats[j].Size
	})

	body := m.appTitle(i18n.T("dashboard")) + "\n\n"
	body += fmt.Sprintf("  %s  ·  %d %s  ·  %d %s\n",
		valueStyle.Render(util.FormatSize(m.result.TotalSize)),
		totalFiles,
		i18n.T("files"),
		len(m.result.Categories),
		i18n.T("categories"),
	)
	body += "  " + mutedStyle.Render(strings.Repeat("─", 50)) + "\n\n"

	const barWidth = 20
	maxSize := int64(1)
	for _, c := range cats {
		if c.Size > maxSize {
			maxSize = c.Size
		}
	}

	for _, c := range cats {
		if c.Size == 0 {
			continue
		}
		name := i18n.CategoryName(c.Category)
		filled := int(float64(c.Size) / float64(maxSize) * barWidth)
		if filled < 1 {
			filled = 1
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		body += fmt.Sprintf("  %-22s %s %s\n", name, mutedStyle.Render(bar), util.FormatSize(c.Size))
	}

	body += "\n" + footer(keyHint("D", i18n.T("select_categories")), keyHint("Enter", i18n.T("confirm")), keyHint("Esc", i18n.T("back")))
	return body
}
