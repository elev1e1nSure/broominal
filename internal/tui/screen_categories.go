package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/cleaner"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
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
	if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
		if m.selectedIdx < len(m.categories) {
			m.categories[m.selectedIdx].selected = !m.categories[m.selectedIdx].selected
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
			m.detailList = buildDetailList(cat.Items, m.width, m.height)
			m.screen = ScreenDetails
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.detailList = buildDetailList(m.categories[m.detailCat].cat.Items, m.width, m.height)
		m.screen = ScreenDetails
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyDetails(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenCategories
		m.selectedIdx = m.detailCat
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	// list handles its own keys
	var cmd tea.Cmd
	m.detailList, cmd = m.detailList.Update(msg)
	return m, cmd
}

func (m model) handleKeyConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("t"))) {
		m.dryRun = !m.dryRun
		m.confirmMsg = buildConfirmMessage(m.categories, m.result)
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
			res, err := cleaner.Run(context.Background(), selected, m.dryRun, m.result)
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
	body += titleStyle.Render(i18n.T("categories")) + "\n\n"

	// Build aligned header using same widths as rows
	catW, sizeW, filesW, riskW, selW := 26, 10, 7, 12, 5
	head := lipgloss.NewStyle().Width(catW).Render(i18n.T("category")) + " " +
		lipgloss.NewStyle().Width(sizeW).Align(lipgloss.Right).Render(i18n.T("size")) + " " +
		lipgloss.NewStyle().Width(filesW).Align(lipgloss.Right).Render(i18n.T("files")) + " " +
		lipgloss.NewStyle().Width(riskW).Render(i18n.T("risk")) + " " +
		lipgloss.NewStyle().Width(selW).Render(i18n.T("select"))
	body += mutedStyle.Render("  "+head) + "\n"
	body += mutedStyle.Render("  "+strings.Repeat("─", catW+sizeW+filesW+riskW+selW+4)) + "\n"

	for i, c := range m.categories {
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
		line := prefix +
			nameSt.Render(i18n.CategoryName(c.cat.Category)) + " " +
			sizeSt.Render(util.FormatSize(c.cat.Size)) + " " +
			filesSt.Render(fmt.Sprintf("%d", c.cat.Files)) + " " +
			riskSt.Render(riskLabel) + " " +
			marker
		body += line + "\n"
	}
	body += "\n" + footer(
		keyHint("Space", i18n.T("toggle")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("D", i18n.T("details")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewWarnRecycleBin() string {
	cat := m.categories[m.detailCat].cat
	return titleStyle.Render(i18n.T("warning")) + "\n\n" +
		dangerStyle.Render(fmt.Sprintf("  "+i18n.T("recycle_bin_warn"), cat.Files)) + "\n" +
		mutedStyle.Render("  "+i18n.T("hint_recycle_warn")) + "\n\n" +
		footer(
			keyHint("Enter", i18n.T("continue_anyway")),
			keyHint("Esc", i18n.T("back")),
		)
}

func (m model) viewDetails() string {
	head := titleStyle.Render(fmt.Sprintf("%s: %s", i18n.T("details"), m.categories[m.detailCat].cat.Category))
	return head + "\n\n" + m.detailList.View() + "\n" + footer(keyHint("Esc", i18n.T("back")))
}

func (m model) viewConfirm() string {
	head := titleStyle.Render(i18n.T("confirm_cleanup"))
	if m.dryRun {
		head += " " + reviewStyle.Render(i18n.T("dry_run"))
	}
	return head + "\n\n" + m.confirmMsg + "\n\n" + footer(
		keyHint("Enter", i18n.T("proceed")),
		keyHint("T", i18n.T("toggle_dry_run")),
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
	return fmt.Sprintf(
		"  Will free: %s\n  Files:     %d\n  Safe:      %s\n  Review:    %s\n",
		util.FormatSize(safe+review), files, util.FormatSize(safe), util.FormatSize(review),
	)
}

func buildDetailList(items []types.Item, w, h int) list.Model {
	var entries []list.Item
	for _, it := range items {
		entries = append(entries, detailItem{it})
	}
	listHeight := h - 4
	if listHeight < 4 {
		listHeight = 4
	}
	listWidth := w
	if listWidth < 20 {
		listWidth = 20
	}
	l := list.New(entries, list.NewDefaultDelegate(), listWidth, listHeight)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return l
}
