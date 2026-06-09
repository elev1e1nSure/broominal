package tui

import (
	"fmt"
	"sort"

	lipgloss "github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
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
		if m.selectedIdx < 2 {
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
		case 1: // Categories
			if m.configCfg != nil {
				var items []configCategoryItem
				for cat, enabled := range m.configCfg.EnabledCategories {
					items = append(items, configCategoryItem{name: cat, enabled: enabled})
				}
				sort.Slice(items, func(i, j int) bool {
					return items[i].name < items[j].name
				})
				m.configCategories = items
			}
			m.selectedIdx = 0
			m.screen = ScreenConfigCategories
			return m, nil
		case 2: // Language
			m.screen = ScreenLanguage
			m.selectedIdx = 0
			return m, nil
		}
	}
	return m, nil
}

func (m model) handleKeyConfigCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
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
		if m.selectedIdx < len(m.configCategories)-1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
		if m.selectedIdx < len(m.configCategories) {
			m.configCategories[m.selectedIdx].enabled = !m.configCategories[m.selectedIdx].enabled
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.configCfg != nil {
			for _, item := range m.configCategories {
				m.configCfg.EnabledCategories[item.name] = item.enabled
			}
			_ = config.Save(m.configCfg)
		}
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m model) handleKeyConfigPresets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc", "m"))) {
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.configCfg != nil {
			switch m.selectedIdx {
			case 0:
				m.configCfg.ApplyPreset(config.PresetSafe)
			case 1:
				m.configCfg.ApplyPreset(config.PresetNormal)
			case 2:
				m.configCfg.ApplyPreset(config.PresetHard)
			}
			_ = config.Save(m.configCfg)
		}
		return m, nil
	}
	return m, nil
}

func (m model) viewConfig() string {
	items := []string{
		i18n.T("config_presets"),
		i18n.T("config_categories"),
		i18n.T("config_language"),
	}
	var body string
	body += titleStyle.Render(i18n.T("config")) + "\n\n"
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

func (m model) viewConfigCategories() string {
	var body string
	body += titleStyle.Render(i18n.T("config_categories")) + "\n\n"
	for i, c := range m.configCategories {
		marker := "[ ]"
		if c.enabled {
			marker = safeStyle.Render("[x]")
		}
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %-30s %s", c.name, marker)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %-30s %s", c.name, marker)) + "\n"
		}
	}
	body += "\n" + footer(
		keyHint("Space", i18n.T("toggle")),
		keyHint("Enter", i18n.T("save")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}

func (m model) viewConfigPresets() string {
	presets := []struct {
		name        string
		desc        string
		preset      config.Preset
		style       lipgloss.Style
		expected    int
	}{
		{"Safe", "Кэш браузеров, темпы, системный мусор", config.PresetSafe, safeStyle, 28},
		{"Normal", "Safe + Telegram", config.PresetNormal, reviewStyle, 29},
		{"Hard", "Всё включено - максимальная очистка", config.PresetHard, dangerStyle, 32},
	}

	currentPreset := -1
	if m.configCfg != nil {
		enabledCount := 0
		for _, v := range m.configCfg.EnabledCategories {
			if v {
				enabledCount++
			}
		}
		for i, p := range presets {
			if enabledCount == p.expected {
				currentPreset = i
				break
			}
		}
	}

	var body string
	body += titleStyle.Render(i18n.T("config_presets")) + "\n\n"
	for i, p := range presets {
		marker := "   "
		if currentPreset == i {
			marker = safeStyle.Render("[x]")
		}
		if i == m.selectedIdx {
			body += selectedStyle.Render("> ") + p.style.Render(p.name) + " " + marker + "\n"
			body += selectedStyle.Render(fmt.Sprintf("  %s\n", p.desc)) + "\n"
		} else {
			body += mutedStyle.Render("  ") + p.style.Render(p.name) + " " + marker + "\n"
			body += mutedStyle.Render(fmt.Sprintf("  %s\n", p.desc)) + "\n"
		}
	}
	body += "\n" + footer(
		keyHint("Enter", i18n.T("apply")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
