package tui

import (
	"fmt"

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
		if m.selectedIdx < 1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		switch m.selectedIdx {
		case 0: // Categories
			if m.configCfg != nil {
				var items []configCategoryItem
				for cat, enabled := range m.configCfg.EnabledCategories {
					items = append(items, configCategoryItem{name: cat, enabled: enabled})
				}
				m.configCategories = items
			}
			m.selectedIdx = 0
			m.screen = ScreenConfigCategories
			return m, nil
		case 1: // Thresholds
			if m.configCfg != nil {
				m.configThresholds = []configThresholdItem{
					{labelKey: "old_installer_months", value: m.configCfg.OldInstallerMonths, min: 1, step: 1},
					{labelKey: "large_file_min_size_mb", value: m.configCfg.LargeFileMinSizeMB, min: 1, step: 10},
					{labelKey: "large_file_months", value: m.configCfg.LargeFileMonths, min: 1, step: 1},
					{labelKey: "old_temp_days", value: m.configCfg.OldTempDays, min: 1, step: 1},
					{labelKey: "old_extension_days", value: m.configCfg.OldExtensionDays, min: 1, step: 1},
					{labelKey: "quarantine_max_age_days", value: m.configCfg.QuarantineMaxAgeDays, min: 1, step: 1},
				}
			}
			m.selectedIdx = 0
			m.screen = ScreenConfigThresholds
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

func (m model) handleKeyConfigThresholds(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.selectedIdx < len(m.configThresholds)-1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("+", "="))) {
		if m.selectedIdx < len(m.configThresholds) {
			it := &m.configThresholds[m.selectedIdx]
			it.value += it.step
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("-"))) {
		if m.selectedIdx < len(m.configThresholds) {
			it := &m.configThresholds[m.selectedIdx]
			it.value -= it.step
			if it.value < it.min {
				it.value = it.min
			}
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		if m.configCfg != nil && len(m.configThresholds) >= 6 {
			m.configCfg.OldInstallerMonths = m.configThresholds[0].value
			m.configCfg.LargeFileMinSizeMB = m.configThresholds[1].value
			m.configCfg.LargeFileMonths = m.configThresholds[2].value
			m.configCfg.OldTempDays = m.configThresholds[3].value
			m.configCfg.OldExtensionDays = m.configThresholds[4].value
			m.configCfg.QuarantineMaxAgeDays = m.configThresholds[5].value
			_ = config.Save(m.configCfg)
		}
		m.screen = ScreenConfig
		m.selectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m model) viewConfig() string {
	items := []string{
		i18n.T("config_categories"),
		i18n.T("config_thresholds"),
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

func (m model) viewConfigThresholds() string {
	var body string
	body += titleStyle.Render(i18n.T("config_thresholds")) + "\n\n"
	for i, c := range m.configThresholds {
		label := i18n.T(c.labelKey)
		val := fmt.Sprintf("%d", c.value)
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %-30s %s", label, val)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %-30s %s", label, val)) + "\n"
		}
	}
	body += "\n" + footer(
		keyHint("+/-", i18n.T("change")),
		keyHint("Enter", i18n.T("save")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
