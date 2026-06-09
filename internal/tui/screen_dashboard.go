package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/style"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
		m.confirmMsg = buildConfirmMessage(m.categories, m.result)
		m.screen = ScreenConfirm
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("m", "q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m model) viewDashboard() string {
	if m.result == nil {
		return m.spinner.View() + " " + i18n.T("scanning") + "\n"
	}

	// Risk summary
	body := m.appTitle(i18n.T("dashboard")) + "\n\n" +
		fmt.Sprintf("  Total found: %s\n", valueStyle.Render(util.FormatSize(m.result.TotalSize))) +
		fmt.Sprintf("  %s Safe:       %s\n", safeStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.SafeSize))) +
		fmt.Sprintf("  %s Review:     %s\n", reviewStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.ReviewSize))) +
		fmt.Sprintf("  %s Danger:     %s\n", dangerStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.DangerSize))) +
		"\n"

	// Category list sorted by size desc
	cats := make([]types.CategorySummary, len(m.result.Categories))
	copy(cats, m.result.Categories)
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Size > cats[j].Size
	})

	visible := m.height - 12
	if visible < 3 {
		visible = 3
	}
	if visible > len(cats) {
		visible = len(cats)
	}

	catW, sizeW, filesW := 30, 10, 8
	head := style.Cyanf("%-*s %*s %*s", catW, i18n.T("category"), sizeW, i18n.T("size"), filesW, i18n.T("files"))
	body += "  " + mutedStyle.Render(head) + "\n"
	body += "  " + mutedStyle.Render(style.Grayf(strings.Repeat("─", catW+sizeW+filesW+2))) + "\n"

	for i := 0; i < visible; i++ {
		c := cats[i]
		name := i18n.CategoryName(c.Category)
		if len(name) > catW {
			name = name[:catW-3] + "..."
		}
		var riskDot string
		switch c.Risk {
		case types.RiskSafe:
			riskDot = safeStyle.Render("●")
		case types.RiskReview:
			riskDot = reviewStyle.Render("●")
		case types.RiskDanger:
			riskDot = dangerStyle.Render("●")
		}
		line := fmt.Sprintf("%-*s %*s %*s %s", catW, name, sizeW, util.FormatSize(c.Size), filesW, style.Cyanf("%d", c.Files), riskDot)
		body += "  " + line + "\n"
	}

	if len(cats) > visible {
		body += "  " + mutedStyle.Render(fmt.Sprintf("... %s (%d)", i18n.T("more_categories"), len(cats)-visible)) + "\n"
	}

	body += "\n" + footer(keyHint("Enter", i18n.T("continue")), keyHint("Esc", i18n.T("back")))
	return body
}
