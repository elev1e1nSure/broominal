package tui

import (
	"fmt"

	lipgloss "github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/pathman"
)

func (m model) handleKeyConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = m.lastMainMenuIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.selectedIdx < 3 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.lastConfigIdx = m.selectedIdx
		switch m.selectedIdx {
		case 0: // Presets
			m.selectedIdx = 0
			m.screen = ScreenConfigPresets
			return m, nil
		case 1: // Language
			m.screen = ScreenLanguage
			m.selectedIdx = 0
			return m, nil
		case 2: // Add/Remove PATH
			inPath, _ := pathman.IsInPath()
			if inPath {
				m.pathOperation = "remove"
			} else {
				m.pathOperation = "add"
			}
			m.pathConfirmIdx = 0
			m.screen = ScreenPathConfirm
			return m, nil
		case 3: // Quarantine settings submenu
			m.selectedIdx = 0
			m.quarantineSettingsMsg = ""
			m.screen = ScreenQuarantineSettings
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleKeyConfigPresets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))) {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))) {
		if m.selectedIdx < 2 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", "space"))) {
		m.applySelectedPreset()
		return m, nil
	}
	return m, nil
}

func (m *model) applySelectedPreset() {
	if m.configCfg == nil {
		return
	}
	switch m.selectedIdx {
	case 0:
		m.configCfg.ApplyPreset(config.PresetQuick)
	case 1:
		m.configCfg.ApplyPreset(config.PresetStandard)
	case 2:
		m.configCfg.ApplyPreset(config.PresetDeep)
	default:
		return
	}
	_ = config.Save(m.configCfg)
}

func (m model) viewConfig() string {
	inPath, _ := pathman.IsInPath()
	pathLabel := i18n.T("config_path")
	if inPath {
		pathLabel = i18n.T("config_path_remove")
	}

	quarantineLabel := i18n.T("config_quarantine")

	items := []string{
		i18n.T("config_presets"),
		i18n.T("config_language"),
		pathLabel,
		quarantineLabel,
	}
	var body string
	body += m.appTitle(i18n.T("config")) + "\n\n"
	for i, item := range items {
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %s", item)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %s", item)) + "\n"
		}
	}
	body += "\n" + footer(
		keyHint("Enter", i18n.T("select")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewConfigPresets() string {
	type presetDef struct {
		nameKey string
		descKey string
		catsKey string
		preset  config.Preset
		style   lipgloss.Style
	}
	presets := []presetDef{
		{"preset_quick", "preset_quick_desc", "preset_quick_cats", config.PresetQuick, safeStyle},
		{"preset_standard", "preset_standard_desc", "preset_standard_cats", config.PresetStandard, reviewStyle},
		{"preset_deep", "preset_deep_desc", "preset_deep_cats", config.PresetDeep, dangerStyle},
	}

	activePreset := ""
	if m.configCfg != nil {
		activePreset = m.configCfg.ActivePreset
	}

	var body string
	body += m.appTitle(i18n.T("config_presets")) + "\n\n"
	for i, p := range presets {
		marker := "   "
		if activePreset == string(p.preset) {
			marker = safeStyle.Render("[x]")
		}
		if i == m.selectedIdx {
			body += selectedStyle.Render("> ") + p.style.Render(i18n.T(p.nameKey)) + " " + marker + "\n"
			body += selectedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.descKey))) + "\n"
			body += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.catsKey))) + "\n\n"
		} else {
			body += mutedStyle.Render("  ") + p.style.Render(i18n.T(p.nameKey)) + " " + marker + "\n"
			body += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.descKey))) + "\n"
			body += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.catsKey))) + "\n\n"
		}
	}
	body += mutedStyle.Render("  "+i18n.T("preset_note")) + "\n\n"
	body += footer(
		keyHint("Enter/Space", i18n.T("apply")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
