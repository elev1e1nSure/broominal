package tui

import (
"fmt"

"github.com/charmbracelet/bubbles/key"
tea "github.com/charmbracelet/bubbletea"
"github.com/elev1e1nSure/broominal/pkg/quarantine"
"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyRestore(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
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
return m, nil
}

func (m model) viewRestore() string {
var body string
body += titleStyle.Render(i18n.T("restore")) + "\n\n"
if len(m.restoreIDs) == 0 {
body += mutedStyle.Render("  " + i18n.T("no_quarantines")) + "\n"
} else {
for i, id := range m.restoreIDs {
if i == m.restoreIdx {
body += selectedStyle.Render(fmt.Sprintf("> %s", id)) + "\n"
} else {
body += mutedStyle.Render(fmt.Sprintf("  %s", id)) + "\n"
}
}
}
if m.restoreResult != "" {
body += "\n" + safeStyle.Render("  [OK] " + m.restoreResult) + "\n"
}
body += "\n" + footer(
keyHint("Enter", i18n.T("restore")),
keyHint("Esc", i18n.T("back")),
)
return body
}