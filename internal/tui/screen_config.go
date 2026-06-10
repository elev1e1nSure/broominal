package tui

import (
	"fmt"

	lipgloss "github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/pathman"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

func (m model) handleKeyConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) {
		m.screen = ScreenMainMenu
		m.selectedIdx = 0
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
		switch m.selectedIdx {
		case 0: // Presets
			m.selectedIdx = 0
			m.screen = ScreenConfigPresets
			return m, nil
		case 1: // Language
			m.screen = ScreenLanguage
			m.selectedIdx = 0
			return m, nil
		case 2: // Check for updates
			m.screen = ScreenUpdating
			m.updateProgress = i18n.T("checking_updates")
			m.updateFromConfig = true
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				release, err := update.CheckForUpdates(m.version)
				return checkUpdateMsg{release, err}
			})
		case 3: // Add/Remove PATH
			inPath, _ := pathman.IsInPath()
			if inPath {
				m.pathOperation = "remove"
			} else {
				m.pathOperation = "add"
			}
			m.pathConfirmIdx = 0
			m.screen = ScreenPathConfirm
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleKeyConfigPresets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		m.screen = ScreenConfig
		m.selectedIdx = 0
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
	if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
		if m.configCfg != nil {
			switch m.selectedIdx {
			case 0:
				m.configCfg.ApplyPreset(config.PresetQuick)
			case 1:
				m.configCfg.ApplyPreset(config.PresetStandard)
			case 2:
				m.configCfg.ApplyPreset(config.PresetDeep)
			}
			_ = config.Save(m.configCfg)
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m model) viewConfig() string {
	inPath, _ := pathman.IsInPath()
	pathLabel := i18n.T("config_path")
	if inPath {
		pathLabel = i18n.T("config_path_remove")
	}
	items := []string{
		i18n.T("config_presets"),
		i18n.T("config_language"),
		i18n.T("check_updates"),
		pathLabel,
	}
	var body string
	body += m.appTitle(i18n.T("config")) + "\n\n"
	for i, item := range items {
		marker := ""
		if i == 3 && inPath {
			marker = safeStyle.Render(" [x]")
		}
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %s", item)) + marker + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %s", item)) + marker + "\n"
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
	body += footer(
		keyHint("Space", i18n.T("apply")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
