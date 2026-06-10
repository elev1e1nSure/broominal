package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.confirmMsg = buildConfirmMessage(m.categories, m.result)
		m.screen = ScreenConfirm
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	return m, nil
}

func riskStyle(risk types.RiskLevel) lipgloss.Style {
	switch risk {
	case types.RiskReview:
		return reviewStyle
	case types.RiskDanger:
		return dangerStyle
	default:
		return safeStyle
	}
}

func riskColor(risk types.RiskLevel) lipgloss.Color {
	switch risk {
	case types.RiskReview:
		return lipgloss.Color("#fbbf24")
	case types.RiskDanger:
		return lipgloss.Color("#f87171")
	default:
		return lipgloss.Color("#4ade80")
	}
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

	// ── Header ──────────────────────────────────────────────────────────────
	body := m.appTitle(i18n.T("dashboard")) + "\n\n"

	// Stats row: total size · files · categories
	secondary := fmt.Sprintf("%d %s  ·  %d %s",
		totalFiles, i18n.T("files"),
		len(m.result.Categories), i18n.T("stat_categories"),
	)
	body += fmt.Sprintf("  %s  %s\n",
		valueStyle.Render(util.FormatSize(m.result.TotalSize)),
		mutedStyle.Render("·  "+secondary),
	)

	// Risk breakdown row
	riskDot := func(label string, size int64, rs lipgloss.Style) string {
		sizeStr := "—"
		if size > 0 {
			sizeStr = util.FormatSize(size)
		}
		return rs.Render("■") + " " + mutedStyle.Render(label) + " " + rs.Render(sizeStr)
	}
	body += fmt.Sprintf("  %s    %s    %s\n",
		riskDot(i18n.T("risk_safe"), m.result.SafeSize, safeStyle),
		riskDot(i18n.T("risk_review"), m.result.ReviewSize, reviewStyle),
		riskDot(i18n.T("risk_danger"), m.result.DangerSize, dangerStyle),
	)

	body += "  " + mutedStyle.Render(strings.Repeat("─", 52)) + "\n\n"

	// ── Category bars ────────────────────────────────────────────────────────
	const nameWidth = 30
	const barWidth = 18
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
		rs := riskStyle(c.Risk)
		name := truncateDisplay(i18n.CategoryName(c.Category), nameWidth)
		filled := int(float64(c.Size) / float64(maxSize) * barWidth)
		if filled < 1 {
			filled = 1
		}
		bar := barFillStyle(riskColor(c.Risk)).Render(strings.Repeat(" ", filled)) + barTrackStyle.Render(strings.Repeat(" ", barWidth-filled))
		sizeStr := fmt.Sprintf("%9s", util.FormatSize(c.Size))
		nameCell := lipgloss.NewStyle().Width(nameWidth).Render(name)
		body += fmt.Sprintf("  %s %s  %s\n", nameCell, bar, rs.Render(sizeStr))
	}

	body += "\n" + footer(
		keyHint("D", i18n.T("select_categories")),
	)
	return body
}
