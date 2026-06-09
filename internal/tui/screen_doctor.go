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
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("f"))) {
		for i := range m.doctorChecks {
			if m.doctorChecks[i].FixKey != "" {
				msg, err := doctor.Fix(m.doctorChecks[i].FixKey)
				if err != nil {
					m.doctorFixResult = style.Failf(i18n.T("status_fail")) + " " + err.Error()
					return m, nil
				}
				m.doctorFixResult = style.Passf(i18n.T("status_pass")) + " " + msg
				if m.doctorChecks[i].FixKey == "admin" {
					return m, tea.Quit
				}
				break
			}
		}
		return m, nil
	}
	return m, nil
}

func (m model) viewDoctor() string {
	var body string
	body += m.appTitle(i18n.T("doctor")) + "\n"
	body += mutedStyle.Render("  "+i18n.T("doctor_desc")) + "\n\n"
	var hasFix bool
	for _, c := range m.doctorChecks {
		var marker string
		switch c.Status {
		case doctor.StatusPass:
			marker = safeStyle.Render(i18n.T("status_pass"))
		case doctor.StatusWarn:
			marker = reviewStyle.Render(i18n.T("status_warn"))
		case doctor.StatusFail:
			marker = dangerStyle.Render(i18n.T("status_fail"))
		}
		body += fmt.Sprintf("  %-28s %s  %s\n", c.Name, marker, mutedStyle.Render(c.Detail))
		if c.Status != doctor.StatusPass && c.Suggestion != "" {
			body += fmt.Sprintf("      %s %s\n", reviewStyle.Render("→"), mutedStyle.Render(c.Suggestion))
		}
		if c.FixKey != "" {
			hasFix = true
		}
	}
	// Show quarantine stats (lazy loaded)
	if m.doctorQuarantineStats.Name != "" {
		body += fmt.Sprintf("  %-28s %s  %s\n", m.doctorQuarantineStats.Name, safeStyle.Render(i18n.T("status_pass")), mutedStyle.Render(m.doctorQuarantineStats.Detail))
	}
	if m.doctorFixResult != "" {
		body += "\n  " + m.doctorFixResult + "\n"
	}
	var hints []string
	if hasFix {
		hints = append(hints, keyHint("F", i18n.T("fix_issue")))
	}
	hints = append(hints, keyHint("Esc", i18n.T("back")))
	body += "\n" + footer(hints...)
	return body
}
