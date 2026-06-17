package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/quarantine"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

func (m model) handleKeyCleaning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))) {
		if m.cleanCtxCancel != nil {
			m.cleanCtxCancel()
			m.cleanCtxCancel = nil
		}
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		m.cleanResult = nil
		m.restoreResult = ""
		m.result = nil
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("r"))) {
		if m.cleanResult == nil {
			return m, nil
		}
		conflicts, err := quarantine.CheckRestoreConflicts(m.cleanResult.RestoreID)
		if err != nil {
			m.err = err
			m.screen = ScreenError
			return m, nil
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
			m.screen = ScreenError
			return m, nil
		}
		if skipped == 0 {
			m.cleanResult = nil // restored
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyRestoreConflict(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenResult
		return m, nil
	}
	return m, nil
}

func (m model) viewCleaning() string {
	var content string

	if m.cleanProgress != nil && m.cleanProgress.Total > 0 {
		fraction := float64(m.cleanProgress.Processed) / float64(m.cleanProgress.Total)
		if fraction > 1.0 {
			fraction = 1.0
		}
		bar := m.progress.ViewAs(fraction)
		barLine := bar

		statusParts := []string{
			fmt.Sprintf("%d%%", m.cleanProgress.Percent()),
		}
		if tp := m.formatSmoothedThroughput(); tp != "" {
			statusParts = append(statusParts, tp)
		}
		if eta := m.formatSmoothedETA(); eta != "" {
			statusParts = append(statusParts, "ETA "+eta)
		}
		statusParts = append(statusParts, fmt.Sprintf("%d/%d", m.cleanProgress.Processed, m.cleanProgress.Total))

		statusText := strings.Join(statusParts, " | ")
		statusLine := lipgloss.NewStyle().Width(m.progress.Width).Align(lipgloss.Left).Render(mutedStyle.Render(statusText))

		content = fmt.Sprintf("\n%s\n\n%s\n", barLine, statusLine)
	} else {
		fraction := 0.0
		if m.scanTotal > 0 {
			fraction = float64(m.scanCompleted) / float64(m.scanTotal)
		}
		bar := m.progress.ViewAs(fraction)
		statusText := fmt.Sprintf("%d%% | %s %d/%d", int(fraction*100), i18n.T("cleaning_in_progress"), m.scanCompleted, m.scanTotal)
		statusLine := lipgloss.NewStyle().Width(m.progress.Width).Align(lipgloss.Left).Render(mutedStyle.Render(statusText))

		content = fmt.Sprintf("\n%s\n\n%s\n", bar, statusLine)
	}

	foot := footer(keyHint("Esc", i18n.T("cancel")))
	return m.appFrame(i18n.T("cleaning"), content, foot)
}

func (m model) viewResult() string {
	if m.cleanResult == nil {
		content := "\n  " + safeStyle.Render("[OK] "+i18n.T("hint_restored")) + "\n"
		foot := footer(keyHint("Q", i18n.T("quit")), keyHint("Esc", i18n.T("back")))
		return m.appFrame(i18n.T("restored"), content, foot)
	}
	content := fmt.Sprintf("  Freed:      %s\n", safeStyle.Render(util.FormatSize(m.cleanResult.Freed))) +
		fmt.Sprintf("  Files:      %s\n", valueStyle.Render(fmt.Sprintf("%d", m.cleanResult.Files)))
	if m.cleanResult.Skipped > 0 {
		label := "Skipped"
		if m.cleanResult.Cancelled {
			label = "Cancelled"
		}
		content += fmt.Sprintf("  %s:    %s\n", label, reviewStyle.Render(fmt.Sprintf("%d", m.cleanResult.Skipped)))
	}
	content += fmt.Sprintf("  Restore ID: %s\n", mutedStyle.Render(m.cleanResult.RestoreID))
	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("R", i18n.T("restore_last")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("cleanup_complete"), content, foot)
}

func (m model) viewRestoreConflict() string {
	content := dangerStyle.Render(fmt.Sprintf("  %d %s:", len(m.conflicts), i18n.T("files_already_exist"))) + "\n"
	for _, p := range m.conflicts {
		content += mutedStyle.Render("    "+p) + "\n"
	}
	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("O", i18n.T("overwrite_all")),
		keyHint("S", i18n.T("skip_all")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("restore_conflicts"), content, foot)
}

func (m model) formatSmoothedThroughput() string {
	if m.smoothBytesPerSec <= 0 {
		return ""
	}
	mbps := m.smoothBytesPerSec / (1024 * 1024)
	if mbps >= 1000 {
		return fmt.Sprintf("%.1f GB/s", mbps/1024)
	}
	if mbps >= 1 {
		return fmt.Sprintf("%.1f MB/s", mbps)
	}
	return fmt.Sprintf("%.1f KB/s", mbps*1024)
}

func (m model) formatSmoothedETA() string {
	if m.smoothETA <= 0 {
		return ""
	}
	d := m.smoothETA.Truncate(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
