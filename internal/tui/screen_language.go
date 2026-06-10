package tui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
)

func (m model) handleKeyLanguage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
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
		langs := i18n.SupportedLanguages()
		if m.selectedIdx < len(langs)-1 {
			m.selectedIdx++
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
		langs := i18n.SupportedLanguages()
		if m.selectedIdx < len(langs) {
			lang := langs[m.selectedIdx]
			i18n.SetLanguage(lang)
			cfg, err := config.Load()
			if err != nil {
				slog.Warn("tui: failed to load config for language change", "error", err)
			}
			if cfg != nil {
				cfg.Language = lang
				if err := config.Save(cfg); err != nil {
					slog.Warn("tui: failed to save language config", "error", err)
				}
			}
		}
		return m, nil
	}
	if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		m.screen = ScreenConfig
		m.selectedIdx = m.lastConfigIdx
		return m, nil
	}
	return m, nil
}

func (m model) viewLanguage() string {
	langs := i18n.SupportedLanguages()
	labels := map[string]string{"en": i18n.T("english"), "ru": i18n.T("russian")}
	var body string
	body += m.appTitle(i18n.T("select_language")) + "\n\n"
	for i, lang := range langs {
		label := labels[lang]
		if label == "" {
			label = lang
		}
		marker := ""
		if lang == i18n.CurrentLanguage() {
			marker = safeStyle.Render(" [x]")
		}
		if i == m.selectedIdx {
			body += selectedStyle.Render(fmt.Sprintf("> %s%s", label, marker)) + "\n"
		} else {
			body += mutedStyle.Render(fmt.Sprintf("  %s%s", label, marker)) + "\n"
		}
	}
	body += "\n" + footer(
		keyHint("Space", i18n.T("apply")),
		keyHint("Enter", i18n.T("confirm")),
		keyHint("Esc", i18n.T("back")),
	)
	return body
}
