package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Screen — текущий экран
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenCategories
	ScreenDetails
	ScreenConfirm
	ScreenResult
)

// TUI model
type model struct {
	screen      Screen
	result      *types.ScanResult
	categories  []categoryItem
	selectedIdx int
	detailCat   int
	detailList  list.Model
	confirmMsg  string
	cleanResult *types.CleanResult
	err         error
	width       int
	height      int
}

type categoryItem struct {
	cat      types.CategorySummary
	selected bool
}

func (c categoryItem) FilterValue() string { return c.cat.Category }

// Start запускает TUI
func Start() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialModel() model {
	return model{
		screen: ScreenDashboard,
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		res, err := scanner.Scan()
		if err != nil {
			return errMsg{err}
		}
		return scanDoneMsg{res}
	}
}

type scanDoneMsg struct {
	result *types.ScanResult
}

type errMsg struct {
	err error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.screen == ScreenDetails && m.height > 4 {
			m.detailList.SetWidth(msg.Width)
			m.detailList.SetHeight(msg.Height - 4)
		}
		return m, nil

	case scanDoneMsg:
		m.result = msg.result
		m.categories = make([]categoryItem, 0, len(msg.result.Categories))
		for _, c := range msg.result.Categories {
			// auto-select safe items
			sel := c.Risk == types.RiskSafe
			m.categories = append(m.categories, categoryItem{cat: c, selected: sel})
		}
		m.screen = ScreenCategories
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.screen == ScreenDetails {
		var cmd tea.Cmd
		m.detailList, cmd = m.detailList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenDashboard:
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
			m.screen = ScreenCategories
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
			return m, tea.Quit
		}

	case ScreenCategories:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			return m, tea.Quit
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
				m.detailList = buildDetailList(m.categories[m.selectedIdx].cat.Items, m.width, m.height)
				m.screen = ScreenDetails
			}
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			m.screen = ScreenConfirm
			m.confirmMsg = buildConfirmMessage(m.categories, m.result)
			return m, nil
		}

	case ScreenDetails:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenCategories
			return m, nil
		}
		// list handles its own keys
		var cmd tea.Cmd
		m.detailList, cmd = m.detailList.Update(msg)
		return m, cmd

	case ScreenConfirm:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenCategories
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			// Execute clean
			var selected []types.Item
			for _, c := range m.categories {
				if c.selected {
					for i := range c.cat.Items {
						c.cat.Items[i].Selected = true
						selected = append(selected, c.cat.Items[i])
					}
				}
			}
			id, freed, files, err := quarantine.Move(selected)
			if err != nil {
				m.err = err
				return m, tea.Quit
			}
			m.cleanResult = &types.CleanResult{
				RestoreID: id,
				Freed:     freed,
				Files:     files,
			}
			// Save report
			_, _ = report.Save(m.result, m.cleanResult)
			m.screen = ScreenResult
			return m, nil
		}

	case ScreenResult:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			return m, tea.Quit
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("r"))) {
			if m.cleanResult != nil {
				err := quarantine.Restore(m.cleanResult.RestoreID)
				if err != nil {
					m.err = err
				} else {
					m.cleanResult = nil // restored
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	switch m.screen {
	case ScreenDashboard:
		return m.viewDashboard()
	case ScreenCategories:
		return m.viewCategories()
	case ScreenDetails:
		return m.viewDetails()
	case ScreenConfirm:
		return m.viewConfirm()
	case ScreenResult:
		return m.viewResult()
	}
	return ""
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#60a5fa"))
	safeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	reviewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24"))
	dangerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#374151"))
)

func (m model) viewDashboard() string {
	if m.result == nil {
		return "Scanning...\n"
	}
	return fmt.Sprintf(
		"%s\n\n%s\n%s\n%s\n%s\n\n%s\n",
		titleStyle.Render(" Broominal — Dashboard "),
		fmt.Sprintf("  Total found:  %s", scanner.FormatSize(m.result.TotalSize)),
		fmt.Sprintf("  %s Safe:       %s", safeStyle.Render("●"), scanner.FormatSize(m.result.SafeSize)),
		fmt.Sprintf("  %s Review:     %s", reviewStyle.Render("●"), scanner.FormatSize(m.result.ReviewSize)),
		fmt.Sprintf("  %s Danger:     %s", dangerStyle.Render("●"), scanner.FormatSize(m.result.DangerSize)),
		mutedStyle.Render("  Press Enter or Space to continue"),
	)
}

func (m model) viewCategories() string {
	var s string
	s += titleStyle.Render(" Categories ") + "\n\n"
	s += fmt.Sprintf("  %-20s %10s %8s %8s  %s\n", "Category", "Size", "Files", "Risk", "Select")
	s += "  " + strings.Repeat("-", 60) + "\n"
	for i, c := range m.categories {
		marker := "[ ]"
		if c.selected {
			marker = "[x]"
		}
		style := mutedStyle
		if i == m.selectedIdx {
			style = selectedStyle
		}
		riskStr := string(c.cat.Risk)
		riskColor := "#9ca3af"
		switch c.cat.Risk {
		case types.RiskSafe:
			riskColor = "#4ade80"
		case types.RiskReview:
			riskColor = "#fbbf24"
		case types.RiskDanger:
			riskColor = "#f87171"
		}
		riskRendered := lipgloss.NewStyle().Width(8).Foreground(lipgloss.Color(riskColor)).Render(riskStr)
		line := fmt.Sprintf("  %-20s %10s %8d %s  %s",
			c.cat.Category,
			scanner.FormatSize(c.cat.Size),
			c.cat.Files,
			riskRendered,
			marker,
		)
		s += style.Render(line) + "\n"
	}
	s += "\n" + mutedStyle.Render("  Space: toggle  Enter: confirm  D: details  Q: quit")
	return s
}

func (m model) viewDetails() string {
	return titleStyle.Render(fmt.Sprintf(" Details: %s ", m.categories[m.detailCat].cat.Category)) + "\n\n" +
		m.detailList.View() + "\n" +
		mutedStyle.Render("  Q/Esc: back")
}

func (m model) viewConfirm() string {
	return titleStyle.Render(" Confirm Cleanup ") + "\n\n" +
		m.confirmMsg + "\n\n" +
		mutedStyle.Render("  Enter: proceed  Esc: back")
}

func (m model) viewResult() string {
	if m.cleanResult == nil {
		return titleStyle.Render(" Restored ") + "\n\n" +
			mutedStyle.Render("  Files restored successfully.") + "\n\n" +
			mutedStyle.Render("  Q: quit")
	}
	return titleStyle.Render(" Cleanup Complete ") + "\n\n" +
		fmt.Sprintf("  Freed:     %s\n", scanner.FormatSize(m.cleanResult.Freed)) +
		fmt.Sprintf("  Files:     %d\n", m.cleanResult.Files) +
		fmt.Sprintf("  Restore:   %s\n\n", m.cleanResult.RestoreID) +
		mutedStyle.Render("  R: restore last  Q: quit")
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

type detailItem struct {
	item types.Item
}

func (d detailItem) FilterValue() string { return d.item.Path }

func (d detailItem) Title() string       { return d.item.Path }
func (d detailItem) Description() string {
	return fmt.Sprintf("%s  %s", scanner.FormatSize(d.item.Size), d.item.Risk)
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
		scanner.FormatSize(safe+review), files, scanner.FormatSize(safe), scanner.FormatSize(review),
	)
}
