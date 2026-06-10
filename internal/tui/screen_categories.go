package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/cleaner"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenDashboard
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.selectedIdx < len(m.categories)-1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("d"))) {
		if m.selectedIdx < len(m.categories) {
			m.detailCat = m.selectedIdx
			cat := m.categories[m.selectedIdx].cat
			if cat.Category == "Recycle Bin" && cat.Files > 10000 {
				m.screen = ScreenWarnRecycleBin
				return m, nil
			}
			m.screen = ScreenCategoryInfo
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
		if m.selectedIdx < len(m.categories) {
			cat := m.categories[m.selectedIdx]
			if cat.cat.Category == "Duplicate Files" && !cat.selected {
				m.detailCat = m.selectedIdx
				m.screen = ScreenWarnDuplicates
				return m, nil
			}
			m.categories[m.selectedIdx].selected = !cat.selected
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenConfirm
		m.confirmMsg = buildConfirmMessage(m.categories, m.result)
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyWarnRecycleBin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenCategories
		m.selectedIdx = m.detailCat
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenCategoryInfo
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyWarnDuplicates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenCategories
		m.selectedIdx = m.detailCat
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.categories[m.detailCat].selected = true
		m.screen = ScreenCategories
		m.selectedIdx = m.detailCat
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyCategoryInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenCategories
		m.selectedIdx = m.detailCat
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenCategories
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenCleaning
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			var selected []types.Item
			for _, c := range m.categories {
				if c.selected {
					for i := range c.cat.Items {
						c.cat.Items[i].Selected = true
						selected = append(selected, c.cat.Items[i])
					}
				}
			}
			cfg, _ := config.Load()
			res, err := cleaner.Run(context.Background(), selected, m.result, cfg)
			if err != nil {
				return cleanDoneMsg{nil, err}
			}
			return cleanDoneMsg{res, nil}
		})
	}
	return m, nil
}

func (m model) viewCategories() string {
	var body string
	body += m.appTitle(i18n.T("categories")) + "\n\n"

	// Build aligned header using same widths as rows
	catW, sizeW, filesW, riskW, selW := 28, 12, 8, 10, 8
	head := lipgloss.NewStyle().Width(catW).Render(i18n.T("category")) + " " +
		lipgloss.NewStyle().Width(sizeW).Align(lipgloss.Right).Render(i18n.T("size")) + " " +
		lipgloss.NewStyle().Width(filesW).Align(lipgloss.Right).Render(i18n.T("files")) + " " +
		lipgloss.NewStyle().Width(riskW).Render(i18n.T("risk")) + " " +
		lipgloss.NewStyle().Width(selW).Render(i18n.T("select"))
	body += mutedStyle.Render("  "+head) + "\n"
	body += mutedStyle.Render("  "+strings.Repeat("─", catW+sizeW+filesW+riskW+selW+4)) + "\n"

	visible := m.height - 9
	if visible < 5 {
		visible = 5
	}
	start, end := clampWindow(m.selectedIdx, len(m.categories), visible)
	for i := start; i < end; i++ {
		c := m.categories[i]
		marker := "[ ]"
		if c.selected {
			marker = safeStyle.Render("[x]")
		}
		prefix := "  "
		nameSt := lipgloss.NewStyle().Width(catW)
		sizeSt := lipgloss.NewStyle().Width(sizeW).Align(lipgloss.Right)
		filesSt := lipgloss.NewStyle().Width(filesW).Align(lipgloss.Right)
		riskSt := lipgloss.NewStyle().Width(riskW)
		if i == m.selectedIdx {
			prefix = selectedStyle.Render("> ")
			nameSt = nameSt.Inherit(selectedStyle)
			sizeSt = sizeSt.Inherit(selectedStyle)
			filesSt = filesSt.Inherit(selectedStyle)
			riskSt = riskSt.Inherit(selectedStyle)
		}
		riskCol := lipgloss.Color("#9ca3af")
		switch c.cat.Risk {
		case types.RiskSafe:
			riskCol = lipgloss.Color("#4ade80")
		case types.RiskReview:
			riskCol = lipgloss.Color("#fbbf24")
		case types.RiskDanger:
			riskCol = lipgloss.Color("#f87171")
		}
		riskSt = riskSt.Foreground(riskCol).Bold(true)
		riskLabel := i18n.T("risk_" + strings.ToLower(string(c.cat.Risk)))
		name := truncateDisplay(i18n.CategoryName(c.cat.Category), catW)
		line := prefix +
			nameSt.Render(name) + " " +
			sizeSt.Render(util.FormatSize(c.cat.Size)) + " " +
			filesSt.Render(fmt.Sprintf("%d", c.cat.Files)) + " " +
			riskSt.Render(riskLabel) + " " +
			marker
		body += line + "\n"
	}
	body += "\n" + footer(
		keyHint("Space", i18n.T("toggle")),
		keyHint("D", i18n.T("details")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewWarnRecycleBin() string {
	cat := m.categories[m.detailCat].cat
	return m.appTitle(i18n.T("warning")) + "\n\n" +
		dangerStyle.Render(fmt.Sprintf("  "+i18n.T("recycle_bin_warn"), cat.Files)) + "\n" +
		mutedStyle.Render("  "+i18n.T("hint_recycle_warn")) + "\n\n" +
		footer(
			keyHint("Enter", i18n.T("proceed")),
			keyHint("Esc", i18n.T("back")),
		)
}

func (m model) viewWarnDuplicates() string {
	return m.appTitle(i18n.T("warning")) + "\n\n" +
		reviewStyle.Render("  "+i18n.T("duplicate_files_warn")) + "\n\n" +
		footer(
			keyHint("Enter", i18n.T("proceed")),
			keyHint("Esc", i18n.T("back")),
		)
}

func (m model) viewCategoryInfo() string {
	if m.detailCat >= len(m.categories) {
		return ""
	}
	cat := m.categories[m.detailCat].cat
	desc := i18n.CategoryDescription(cat.Category)

	var body string
	body += m.appTitle(i18n.T("details")) + "\n\n"

	// Category name and basic stats
	body += "  " + selectedStyle.Render(i18n.CategoryName(cat.Category)) + "\n"
	body += "  " + mutedStyle.Render(fmt.Sprintf("  %s: %s  |  %s: %d  |  %s: %s",
		i18n.T("size"), util.FormatSize(cat.Size),
		i18n.T("files"), cat.Files,
		i18n.T("risk"), i18n.T("risk_"+strings.ToLower(string(cat.Risk))))) + "\n\n"

	// Description box
	if desc != "" {
		body += "  " + mutedStyle.Render(desc) + "\n\n"
	}

	body += footer(keyHint("Esc", i18n.T("back")))
	return body
}

func (m model) viewConfirm() string {
	head := m.appTitle(i18n.T("confirm_cleanup"))
	return head + "\n\n" + m.confirmMsg +
		"  " + mutedStyle.Render(strings.Repeat("─", 42)) + "\n\n" +
		footer(
			keyHint("Enter", i18n.T("confirm")),
			keyHint("Esc", i18n.T("back")),
		)
}

func buildConfirmMessage(cats []categoryItem, result *types.ScanResult) string {
	var safe, review int64
	var files int
	for _, c := range cats {
		if c.selected {
			if c.cat.Risk == types.RiskSafe {
				safe += c.cat.Size
			} else {
				review += c.cat.Size
			}
			files += c.cat.Files
		}
	}
	total := safe + review

	const labelWidth = 22
	row := func(label string, val string, vs lipgloss.Style) string {
		labelCell := lipgloss.NewStyle().Width(labelWidth).Render(mutedStyle.Render(label))
		return "  " + labelCell + vs.Render(val) + "\n"
	}

	body := row(i18n.T("will_free"), util.FormatSize(total), valueStyle)
	body += row(i18n.T("files"), fmt.Sprintf("%d", files), valueStyle)
	body += "\n"
	body += row(i18n.T("risk_safe"), util.FormatSize(safe), safeStyle)
	if review > 0 {
		body += row(i18n.T("risk_review"), util.FormatSize(review), reviewStyle)
		body += "\n" + reviewStyle.Render("  ⚠  "+i18n.T("contains_review_files")) + "\n\n"
	} else {
		body += "\n"
	}
	return body
}
