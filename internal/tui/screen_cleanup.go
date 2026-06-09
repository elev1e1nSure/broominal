package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyQuarantineCleanup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("t"))) {
		m.dryRun = !m.dryRun
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("a"))) {
		m.quarantineCleanupAll = !m.quarantineCleanupAll
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			var deleted int
			var freed int64
			var err error
			if m.quarantineCleanupAll {
				deleted, freed, err = quarantine.CleanupAll(m.dryRun)
			} else {
				cfg, _ := config.Load()
				maxAge := 30
				if cfg != nil && cfg.QuarantineMaxAgeDays > 0 {
					maxAge = cfg.QuarantineMaxAgeDays
				}
				deleted, freed, err = quarantine.Cleanup(maxAge, m.dryRun)
			}
			if err != nil {
				return errMsg{err}
			}
			label := "Removed"
			if m.dryRun {
				label = "Would remove"
			}
			modeLabel := i18n.T("old_only")
			if m.quarantineCleanupAll {
				modeLabel = i18n.T("all_quarantines")
			}
			return cleanDoneMsg{&types.CleanResult{
				RestoreID: fmt.Sprintf("%s %d quarantines (%s), freed %s", label, deleted, modeLabel, util.FormatSize(freed)),
				Freed:     freed,
				Files:     deleted,
			}, nil}
		})
	}
	return m, nil
}

func (m model) viewQuarantineCleanup() string {
	head := titleStyle.Render(i18n.T("quarantine_cleanup"))
	if m.dryRun {
		head += " " + reviewStyle.Render(i18n.T("dry_run"))
	}
	var modeLine string
	if m.quarantineCleanupAll {
		modeLine = safeStyle.Render(i18n.T("all_quarantines")) + " " + mutedStyle.Render("/ "+i18n.T("old_only"))
	} else {
		modeLine = mutedStyle.Render(i18n.T("all_quarantines") + " /") + " " + safeStyle.Render(i18n.T("old_only"))
	}
	return head + "\n\n" +
		mutedStyle.Render("  "+i18n.T("cleanup_desc")) + "\n\n" +
		"  " + i18n.T("mode") + ": " + modeLine + "\n\n" +
		footer(
			keyHint("A", i18n.T("toggle_mode")),
			keyHint("T", i18n.T("toggle_dry_run")),
			keyHint("Enter", i18n.T("proceed")),
			keyHint("Esc", i18n.T("back")),
		)
}
