package tui

import (
	"fmt"

	lipgloss "github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return m, tea.Quit
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
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
		case 2: // PATH Settings
			m.screen = ScreenPathSettings
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))) {
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
	pathLabel := i18n.T("config_path")

	quarantineLabel := i18n.T("config_quarantine")

	items := []string{
		i18n.T("config_presets"),
		i18n.T("config_language"),
		pathLabel,
		quarantineLabel,
	}
	var content string
	for i, item := range items {
		if i == m.selectedIdx {
			content += selectedStyle.Render(fmt.Sprintf("► %s", item)) + "\n"
		} else {
			content += mutedStyle.Render(fmt.Sprintf("  %s", item)) + "\n"
		}
	}
	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("Enter", i18n.T("select")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("config"), content, foot)
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

	var content string
	for i, p := range presets {
		marker := "[ ]"
		if activePreset == string(p.preset) {
			marker = safeStyle.Render("[x]")
		}
		if i == m.selectedIdx {
			content += selectedStyle.Render("► ") + p.style.Render(i18n.T(p.nameKey)) + " " + marker + "\n"
			content += selectedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.descKey))) + "\n"
			content += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.catsKey))) + "\n\n"
		} else {
			content += mutedStyle.Render("  ") + p.style.Render(i18n.T(p.nameKey)) + " " + marker + "\n"
			content += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.descKey))) + "\n"
			content += mutedStyle.Render(fmt.Sprintf("  %s", i18n.T(p.catsKey))) + "\n\n"
		}
	}
	content += mutedStyle.Render(i18n.T("preset_note")) + "\n"
	foot := footer(
		keyHint("Q", i18n.T("quit")),
		keyHint("Enter/Space", i18n.T("apply")),
		keyHint("Esc", i18n.T("back")),
	)
	return m.appFrame(i18n.T("config_presets"), content, foot)
}
