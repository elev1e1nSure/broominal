package tui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
		m.screen = ScreenCategories
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
	return titleStyle.Render(i18n.T("dashboard")) + "\n\n" +
		fmt.Sprintf("  Total found: %s\n", valueStyle.Render(util.FormatSize(m.result.TotalSize))) +
		fmt.Sprintf("  %s Safe:       %s\n", safeStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.SafeSize))) +
		fmt.Sprintf("  %s Review:     %s\n", reviewStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.ReviewSize))) +
		fmt.Sprintf("  %s Danger:     %s\n\n", dangerStyle.Render("●"), valueStyle.Render(util.FormatSize(m.result.DangerSize))) +
		mutedStyle.Render("v"+m.version) + "\n" +
		footer(keyHint("Enter", i18n.T("continue")), keyHint("Esc", i18n.T("back")))
}
