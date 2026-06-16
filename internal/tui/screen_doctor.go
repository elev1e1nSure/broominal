package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/style"
)

func (m model) handleKeyDoctor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}

	// Confirmation mode: Enter confirms, Esc cancels.
	if m.doctorPendingFix != "" {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			result, err := doctor.Fix(m.doctorPendingFix)
			m.doctorPendingFix = ""
			if err != nil {
				m.doctorFixResult = style.Failf("[FAIL]") + " " + err.Error()
			} else {
				m.doctorFixResult = style.Passf("[OK]") + " " + result
				// Refresh checks so the WARN disappears.
				m.doctorChecks = doctor.Run()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.doctorPendingFix = ""
		}
		return m, nil
	}

	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("f"))) {
		for i := range m.doctorChecks {
			if m.doctorChecks[i].FixKey == "" {
				continue
			}
			fixKey := m.doctorChecks[i].FixKey
			if fixKey == "purge_damaged" {
				// Require confirmation before deleting.
				m.doctorPendingFix = fixKey
				m.doctorFixResult = ""
			} else {
				result, err := doctor.Fix(fixKey)
				if err != nil {
					m.doctorFixResult = style.Failf("[FAIL]") + " " + err.Error()
				} else {
					m.doctorFixResult = style.Passf("[OK]") + " " + result
					if fixKey == "admin" {
						return m, tea.Quit
					}
				}
			}
			break
		}
		return m, nil
	}
	return m, nil
}

func (m model) viewDoctor() string {
	var body string
	body += m.appTitle(i18n.T("doctor")) + "\n\n"

	var hasFix bool
	for _, c := range m.doctorChecks {
		var marker string
		switch c.Status {
		case doctor.StatusPass:
			marker = safeStyle.Render("[ OK ]")
		case doctor.StatusWarn:
			marker = reviewStyle.Render("[WARN]")
		case doctor.StatusFail:
			marker = dangerStyle.Render("[FAIL]")
		}
		body += fmt.Sprintf("  %-30s %s  %s\n", c.Name, marker, mutedStyle.Render(c.Detail))
		if c.Status != doctor.StatusPass {
			if c.FixKey != "" {
				body += mutedStyle.Render(fmt.Sprintf("    → %s", i18n.T("suggest_press_f_to_fix"))) + "\n"
			} else if c.Suggestion != "" {
				body += mutedStyle.Render(fmt.Sprintf("    → %s", c.Suggestion)) + "\n"
			}
		}
		if c.FixKey != "" {
			hasFix = true
		}
	}

	if m.doctorPendingFix != "" {
		body += "\n  " + reviewStyle.Render(i18n.T("confirm_purge_damaged")) + "\n"
		body += "\n" + footer(
			keyHint("Enter", i18n.T("confirm_yes")),
			keyHint("Esc", i18n.T("confirm_no")),
		)
		return body
	}

	if m.doctorFixResult != "" {
		body += "\n  " + m.doctorFixResult + "\n"
	}

	var hints []string
	hints = append(hints, keyHint("Q", i18n.T("quit")))
	if hasFix {
		hints = append(hints, keyHint("F", i18n.T("fix_issue")))
	}
	hints = append(hints, keyHint("Esc", i18n.T("back")))
	body += "\n" + footer(hints...)
	return body
}
