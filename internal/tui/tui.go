package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/report"
	"github.com/elev1e1nSure/broominal/pkg/scanner"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Screen — текущий экран
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenDashboard
	ScreenCategories
	ScreenWarnRecycleBin
	ScreenDetails
	ScreenConfirm
	ScreenCleaning
	ScreenResult
	ScreenRestoreConflict
	ScreenError
	ScreenRestore
	ScreenDoctor
	ScreenConfig
	ScreenQuarantineCleanup
)

// TUI model
type model struct {
	screen                Screen
	result                *types.ScanResult
	categories            []categoryItem
	selectedIdx           int
	detailCat             int
	detailList            list.Model
	confirmMsg            string
	cleanResult           *types.CleanResult
	spinner               spinner.Model
	err                   error
	width                 int
	height                int
	dryRun                bool
	conflicts             []string
	restoreForceOverwrite bool
	// Restore screen
	restoreIDs   []string
	restoreIdx   int
	restoreResult string
	// Doctor screen
	doctorChecks []doctor.Check
	// Config screen
	configView   string
	// Cleanup screen
	cleanupResult string
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
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	return model{
		screen:  ScreenMainMenu,
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type scanDoneMsg struct {
	result *types.ScanResult
}

type errMsg struct {
	err error
}

type cleanDoneMsg struct {
	result *types.CleanResult
	err    error
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
		m.screen = ScreenError
		return m, nil

	case cleanDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.screen = ScreenError
			return m, nil
		}
		m.cleanResult = msg.result
		m.screen = ScreenResult
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

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
	case ScreenMainMenu:
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
			if m.selectedIdx < 4 {
				m.selectedIdx++
			}
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			switch m.selectedIdx {
			case 0: // Scan & Clean
				m.screen = ScreenDashboard
				return m, func() tea.Msg {
					res, err := scanner.Scan()
					if err != nil {
						return errMsg{err}
					}
					return scanDoneMsg{res}
				}
			case 1: // Restore
				ids, err := quarantine.List()
				if err != nil {
					m.err = err
					m.screen = ScreenError
					return m, nil
				}
				m.restoreIDs = ids
				m.restoreIdx = 0
				m.screen = ScreenRestore
				return m, nil
			case 2: // Doctor
				m.doctorChecks = doctor.Run()
				m.screen = ScreenDoctor
				return m, nil
			case 3: // Config
				cfg, err := config.Load()
				if err != nil {
					m.err = err
					m.screen = ScreenError
					return m, nil
				}
				data, _ := json.MarshalIndent(cfg, "", "  ")
				m.configView = string(data)
				m.screen = ScreenConfig
				return m, nil
			case 4: // Quarantine Cleanup
				m.screen = ScreenQuarantineCleanup
				return m, nil
			}
		}

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

	case ScreenWarnRecycleBin:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenCategories
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			m.detailList = buildDetailList(m.categories[m.detailCat].cat.Items, m.width, m.height)
			m.screen = ScreenDetails
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
				id, freed, files, err := quarantine.Move(selected, m.dryRun)
				if err != nil {
					return cleanDoneMsg{nil, err}
				}
				res := &types.CleanResult{
					RestoreID: id,
					Freed:     freed,
					Files:     files,
				}
				if !m.dryRun {
					_, _ = report.Save(m.result, res)
				}
				return cleanDoneMsg{res, nil}
			})
		}

	case ScreenResult:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			return m, tea.Quit
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("r"))) {
			if m.dryRun || m.cleanResult == nil {
				return m, nil
			}
			conflicts, err := quarantine.CheckRestoreConflicts(m.cleanResult.RestoreID)
			if err != nil {
				m.err = err
				return m, tea.Quit
			}
			if len(conflicts) > 0 {
				m.conflicts = conflicts
				m.restoreForceOverwrite = false
				m.screen = ScreenRestoreConflict
				return m, nil
			}
			_, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, false)
			if err != nil {
				m.err = err
				return m, tea.Quit
			}
			if skipped == 0 {
				m.cleanResult = nil // restored
			}
			return m, nil
		}

	case ScreenRestoreConflict:
		if key.Matches(msg, key.NewBinding(key.WithKeys("o"))) {
			if m.cleanResult != nil {
				restored, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, true)
				if err != nil {
					m.err = err
					m.screen = ScreenError
					return m, nil
				}
				_ = restored
				_ = skipped
				m.cleanResult = nil // restored
			}
			m.screen = ScreenResult
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("s"))) {
			if m.cleanResult != nil {
				restored, skipped, err := quarantine.Restore(m.cleanResult.RestoreID, false)
				if err != nil {
					m.err = err
					m.screen = ScreenError
					return m, nil
				}
				_ = restored
				if skipped == 0 {
					m.cleanResult = nil
				}
			}
			m.screen = ScreenResult
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "c"))) {
			m.screen = ScreenResult
			return m, nil
		}

	case ScreenRestore:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
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
			if m.restoreIdx < len(m.restoreIDs)-1 {
				m.restoreIdx++
			}
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			if len(m.restoreIDs) == 0 {
				return m, nil
			}
			id := m.restoreIDs[m.restoreIdx]
			restored, skipped, err := quarantine.Restore(id, false)
			if err != nil {
				m.err = err
				m.screen = ScreenError
				return m, nil
			}
			m.restoreResult = fmt.Sprintf("Restored %d files (%d skipped)", restored, skipped)
			// Refresh list
			m.restoreIDs, _ = quarantine.List()
			if m.restoreIdx >= len(m.restoreIDs) {
				m.restoreIdx = len(m.restoreIDs) - 1
				if m.restoreIdx < 0 {
					m.restoreIdx = 0
				}
			}
			return m, nil
		}

	case ScreenDoctor:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenMainMenu
			m.selectedIdx = 0
			return m, nil
		}

	case ScreenConfig:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenMainMenu
			m.selectedIdx = 0
			return m, nil
		}

	case ScreenQuarantineCleanup:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			m.screen = ScreenMainMenu
			m.selectedIdx = 0
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("t"))) {
			m.dryRun = !m.dryRun
			return m, nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				cfg, _ := config.Load()
				maxAge := 30
				if cfg != nil && cfg.QuarantineMaxAgeDays > 0 {
					maxAge = cfg.QuarantineMaxAgeDays
				}
				deleted, freed, err := quarantine.Cleanup(maxAge, m.dryRun)
				if err != nil {
					return errMsg{err}
				}
				label := "Removed"
				if m.dryRun {
					label = "Would remove"
				}
				return cleanDoneMsg{&types.CleanResult{
					RestoreID: fmt.Sprintf("%s %d quarantines, freed %s", label, deleted, scanner.FormatSize(freed)),
					Freed:     freed,
					Files:     deleted,
				}, nil}
			})
		}

	case ScreenError:
		if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case ScreenMainMenu:
		return m.viewMainMenu()
	case ScreenDashboard:
		return m.viewDashboard()
	case ScreenCategories:
		return m.viewCategories()
	case ScreenWarnRecycleBin:
		return m.viewWarnRecycleBin()
	case ScreenDetails:
		return m.viewDetails()
	case ScreenConfirm:
		return m.viewConfirm()
	case ScreenCleaning:
		return m.viewCleaning()
	case ScreenResult:
		return m.viewResult()
	case ScreenRestoreConflict:
		return m.viewRestoreConflict()
	case ScreenError:
		return m.viewError()
	case ScreenRestore:
		return m.viewRestore()
	case ScreenDoctor:
		return m.viewDoctor()
	case ScreenConfig:
		return m.viewConfig()
	case ScreenQuarantineCleanup:
		return m.viewQuarantineCleanup()
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
	dryLabel := ""
	if m.dryRun {
		dryLabel = reviewStyle.Render(" [DRY-RUN] ")
	}
	return titleStyle.Render(" Confirm Cleanup ") + dryLabel + "\n\n" +
		m.confirmMsg + "\n\n" +
		mutedStyle.Render("  Enter: proceed  T: toggle dry-run  Esc: back")
}

func (m model) viewResult() string {
	if m.cleanResult == nil {
		return titleStyle.Render(" Restored ") + "\n\n" +
			mutedStyle.Render("  Files restored successfully.") + "\n\n" +
			mutedStyle.Render("  Q: quit")
	}
	if m.dryRun {
		return titleStyle.Render(" Dry-Run Complete ") + "\n\n" +
			fmt.Sprintf("  Would free: %s\n", scanner.FormatSize(m.cleanResult.Freed)) +
			fmt.Sprintf("  Files:      %d\n\n", m.cleanResult.Files) +
			mutedStyle.Render("  Q: quit")
	}
	return titleStyle.Render(" Cleanup Complete ") + "\n\n" +
		fmt.Sprintf("  Freed:     %s\n", scanner.FormatSize(m.cleanResult.Freed)) +
		fmt.Sprintf("  Files:     %d\n", m.cleanResult.Files) +
		fmt.Sprintf("  Restore:   %s\n\n", m.cleanResult.RestoreID) +
		mutedStyle.Render("  R: restore last  Q: quit")
}

func (m model) viewRestoreConflict() string {
	var s string
	s += titleStyle.Render(" Restore Conflicts ") + "\n\n"
	s += dangerStyle.Render(fmt.Sprintf("  %d file(s) already exist at original paths:", len(m.conflicts))) + "\n"
	for _, p := range m.conflicts {
		s += mutedStyle.Render(fmt.Sprintf("    %s", p)) + "\n"
	}
	s += "\n" + mutedStyle.Render("  O: overwrite all  S: skip all  C/Esc: cancel")
	return s
}

func (m model) viewWarnRecycleBin() string {
	cat := m.categories[m.detailCat].cat
	return titleStyle.Render(" Warning ") + "\n\n" +
		dangerStyle.Render(fmt.Sprintf("  Recycle Bin contains %d files.", cat.Files)) + "\n" +
		mutedStyle.Render("  Opening details may be very slow.") + "\n\n" +
		mutedStyle.Render("  Enter: continue anyway  Esc: back")
}

func (m model) viewCleaning() string {
	return titleStyle.Render(" Cleaning... ") + "\n\n" +
		fmt.Sprintf("  %s Moving files to quarantine...\n", m.spinner.View()) +
		mutedStyle.Render("  Please wait, this may take a while.")
}

func (m model) viewError() string {
	return titleStyle.Render(" Error ") + "\n\n" +
		dangerStyle.Render(fmt.Sprintf("  %v", m.err)) + "\n\n" +
		mutedStyle.Render("  Press Q or Esc to quit")
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

func (m model) viewMainMenu() string {
	items := []string{
		"Scan & Clean",
		"Restore from Quarantine",
		"Doctor (Health Checks)",
		"View Config",
		"Quarantine Cleanup",
	}
	var s string
	s += titleStyle.Render(" Broominal — Main Menu ") + "\n\n"
	for i, item := range items {
		style := mutedStyle
		prefix := "  "
		if i == m.selectedIdx {
			style = selectedStyle
			prefix = "> "
		}
		s += style.Render(fmt.Sprintf("%s%s", prefix, item)) + "\n"
	}
	s += "\n" + mutedStyle.Render("  Enter: select  Q: quit")
	return s
}

func (m model) viewRestore() string {
	var s string
	s += titleStyle.Render(" Restore from Quarantine ") + "\n\n"
	if len(m.restoreIDs) == 0 {
		s += mutedStyle.Render("  No quarantines available.") + "\n"
	} else {
		for i, id := range m.restoreIDs {
			style := mutedStyle
			prefix := "  "
			if i == m.restoreIdx {
				style = selectedStyle
				prefix = "> "
			}
			s += style.Render(fmt.Sprintf("%s%s", prefix, id)) + "\n"
		}
	}
	if m.restoreResult != "" {
		s += "\n" + safeStyle.Render("  "+m.restoreResult) + "\n"
	}
	s += "\n" + mutedStyle.Render("  Enter: restore  Q/Esc: back")
	return s
}

func (m model) viewDoctor() string {
	var s string
	s += titleStyle.Render(" Doctor — Health Checks ") + "\n\n"
	for _, c := range m.doctorChecks {
		var marker string
		switch c.Status {
		case doctor.StatusPass:
			marker = safeStyle.Render("[PASS]")
		case doctor.StatusWarn:
			marker = reviewStyle.Render("[WARN]")
		case doctor.StatusFail:
			marker = dangerStyle.Render("[FAIL]")
		}
		s += fmt.Sprintf("  %-28s %s  %s\n", c.Name, marker, mutedStyle.Render(c.Detail))
	}
	s += "\n" + mutedStyle.Render("  Q/Esc: back")
	return s
}

func (m model) viewConfig() string {
	var s string
	s += titleStyle.Render(" Current Config ") + "\n\n"
	for _, line := range strings.Split(m.configView, "\n") {
		s += mutedStyle.Render("  "+line) + "\n"
	}
	s += "\n" + mutedStyle.Render("  Q/Esc: back")
	return s
}

func (m model) viewQuarantineCleanup() string {
	var s string
	dryLabel := ""
	if m.dryRun {
		dryLabel = reviewStyle.Render(" [DRY-RUN] ")
	}
	s += titleStyle.Render(" Quarantine Cleanup ") + dryLabel + "\n\n"
	s += mutedStyle.Render("  Deletes quarantines older than max age days.") + "\n\n"
	s += mutedStyle.Render("  T: toggle dry-run  Enter: proceed  Q/Esc: back")
	return s
}
