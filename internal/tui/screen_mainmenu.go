package tui

import (
tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/bubbles/key"
"github.com/elev1e1nSure/broominal/pkg/doctor"
"github.com/elev1e1nSure/broominal/pkg/quarantine"
"github.com/elev1e1nSure/broominal/pkg/config"
"github.com/elev1e1nSure/broominal/pkg/i18n"
"github.com/elev1e1nSure/broominal/pkg/scanner"
"fmt"
)

func (m model) handleKeyMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
cfg, _ := config.Load()
if cfg == nil {
cfg = config.Default()
}
res, err := scanner.ScanWithConfig(cfg)
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
case 3: // Quarantine Cleanup
m.screen = ScreenQuarantineCleanup
return m, nil
case 4: // Settings
cfg, err := config.Load()
if err != nil {
m.err = err
m.screen = ScreenError
return m, nil
}
m.configCfg = cfg
m.selectedIdx = 0
m.screen = ScreenConfig
return m, nil
}
}
return m, nil
}

func (m model) viewMainMenu() string {
items := []string{
i18n.T("menu_scan_clean"),
i18n.T("menu_restore"),
i18n.T("menu_doctor"),
i18n.T("menu_cleanup"),
i18n.T("menu_settings"),
}
var body string
body += titleStyle.Render(i18n.T("main_menu")) + "\n\n"
for i, item := range items {
if i == m.selectedIdx {
body += selectedStyle.Render(fmt.Sprintf("> %s", item)) + "\n"
} else {
body += mutedStyle.Render(fmt.Sprintf("  %s", item)) + "\n"
}
}
body += "\n" + footer(keyHint("Enter", i18n.T("select")), keyHint("Q", i18n.T("quit")))
return body
}