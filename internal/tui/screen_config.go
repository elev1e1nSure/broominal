package tui

import (
	"fmt"
	"sort"

	lipgloss "github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/update"
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
		case 1: // Categories
			if m.configCfg != nil {
				var items []configCategoryItem
				for cat, enabled := range m.configCfg.EnabledCategories {
					items = append(items, configCategoryItem{name: cat, enabled: enabled, group: categoryGroup(cat)})
				}
				sort.Slice(items, func(i, j int) bool {
					if items[i].group != items[j].group {
						return items[i].group < items[j].group
					}
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
		case 3: // Check for updates
			m.screen = ScreenUpdating
			m.updateProgress = i18n.T("checking_updates")
			m.updateFromConfig = true
			return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
				release, err := update.CheckForUpdates(m.version)
				return checkUpdateMsg{release, err}
			})
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
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "m"))) {
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
		if m.selectedIdx < 4 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.configCfg != nil {
			switch m.selectedIdx {
			case 0:
				m.configCfg.ApplyPreset(config.PresetQuick)
			case 1:
				m.configCfg.ApplyPreset(config.PresetStandard)
			case 2:
				m.configCfg.ApplyPreset(config.PresetDeveloper)
			case 3:
				m.configCfg.ApplyPreset(config.PresetGamer)
			case 4:
				m.configCfg.ApplyPreset(config.PresetDeep)
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
		i18n.T("check_updates"),
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

func categoryGroup(name string) string {
	switch name {
	case "Temp", "Logs", "Thumbnails Cache", "DirectX Shader Cache", "Delivery Optimization",
		"Windows Error Reports", "Windows Update Cache", "Crash & Memory Dumps",
		"Nvidia Installer Leftovers", "Windows Prefetch", "Icon Cache", "Empty Folders", "Windows Defender":
		return "1_system"
	case "Browser Cache", "Edge Code Cache", "Chrome Code Cache", "Firefox Cache2",
		"Opera Cache", "Brave Cache", "Vivaldi Cache", "Yandex Cache":
		return "2_browsers"
	case "Messenger Cache":
		return "3_messengers"
	case "Steam Cache", "Epic Games Cache", "Battle.net Cache", "Rockstar Cache",
		"EA App Cache", "Ubisoft Cache", "GOG Galaxy Cache":
		return "4_games"
	case "VSCode Cache", "npm Cache", "pip Cache", "Git Cache", "Visual Studio Cache",
		"Docker Cache", "JetBrains Cache", "Go Build Cache", "Rust Cache", "NuGet Cache", "Unity Cache":
		return "5_dev"
	case "Spotify Cache", "OneDrive Cache", "Office Cache", "Adobe Cache", "OBS Cache", "TeamViewer Logs":
		return "6_apps"
	default:
		return "0_user"
	}
}

func groupTitle(group string) string {
	switch group {
	case "1_system":
		return i18n.T("group_system")
	case "2_browsers":
		return i18n.T("group_browsers")
	case "3_messengers":
		return i18n.T("group_messengers")
	case "4_games":
		return i18n.T("group_games")
	case "5_dev":
		return i18n.T("group_dev")
	case "6_apps":
		return i18n.T("group_apps")
	default:
		return i18n.T("group_user")
	}
}

func (m model) viewConfigCategories() string {
	var body string
	body += m.appTitle(i18n.T("config_categories")) + "\n\n"
	visible := m.height - 6
	if visible < 5 {
		visible = 5
	}
	start, end := clampWindow(m.selectedIdx, len(m.configCategories), visible)
	var lastGroup string
	for i := start; i < end; i++ {
		c := m.configCategories[i]
		if c.group != lastGroup {
			lastGroup = c.group
			body += "\n" + lipgloss.NewStyle().Bold(true).Render("  "+groupTitle(c.group)) + "\n"
		}
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
		name     string
		desc     string
		preset   config.Preset
		style    lipgloss.Style
		expected int
	}{
		{"Quick", "~5 сек — браузеры, темпы, системный мусор", config.PresetQuick, safeStyle, 12},
		{"Standard", "Quick + мессенджеры, ланчеры, dev-инструменты", config.PresetStandard, reviewStyle, 34},
		{"Developer", "Standard + Docker, JetBrains, Go, Rust, Unity", config.PresetDeveloper, reviewStyle, 42},
		{"Gamer", "Quick + ланчеры, мессенджеры, большие файлы", config.PresetGamer, reviewStyle, 28},
		{"Deep", "Всё включено — максимальная очистка", config.PresetDeep, dangerStyle, 50},
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
	body += m.appTitle(i18n.T("config_presets")) + "\n\n"
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
