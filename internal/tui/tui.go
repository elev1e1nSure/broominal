package tui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// Start запускает TUI
func Start() error {
	cfg, _ := config.Load()
	if cfg != nil && cfg.Language != "" {
		i18n.SetLanguage(cfg.Language)
	}
	m := initialModel()
	if cfg == nil || cfg.Language == "" {
		// first run: try to auto-detect, then show language picker
		if lang, err := i18n.DetectFromIP(); err == nil {
			i18n.SetLanguage(lang)
			if cfg != nil {
				cfg.Language = lang
				if err := config.Save(cfg); err != nil {
					slog.Warn("tui: failed to save language config", "error", err)
				}
			}
		}
		m.screen = ScreenLanguage
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

type scanDoneMsg struct {
	result *types.ScanResult
}

type errMsg struct {
	err error
}

type cleanDoneMsg struct {
	result *types.CleanResult
	err    error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.screen == ScreenDetails && m.height > 4 {
			m.detailList.SetWidth(msg.Width)
			m.detailList.SetHeight(msg.Height - 4)
		}
		return m, nil

	case scanDoneMsg:
		m.result = msg.result
		m.categories = make([]categoryItem, 0, len(msg.result.Categories))
		for _, c := range msg.result.Categories {
			// auto-select safe items
			sel := c.Risk == types.RiskSafe
			m.categories = append(m.categories, categoryItem{cat: c, selected: sel})
		}
		m.screen = ScreenCategories
		return m, nil

	case errMsg:
		m.err = msg.err
		m.screen = ScreenError
		return m, nil

	case cleanDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.screen = ScreenError
			return m, nil
		}
		m.cleanResult = msg.result
		m.screen = ScreenResult
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.screen == ScreenDetails {
		var cmd tea.Cmd
		m.detailList, cmd = m.detailList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// global quit
	if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
		return m, tea.Quit
	}

	switch m.screen {
	case ScreenMainMenu:
		return m.handleKeyMainMenu(msg)
	case ScreenDashboard:
		return m.handleKeyDashboard(msg)
	case ScreenCategories:
		return m.handleKeyCategories(msg)
	case ScreenWarnRecycleBin:
		return m.handleKeyWarnRecycleBin(msg)
	case ScreenDetails:
		return m.handleKeyDetails(msg)
	case ScreenConfirm:
		return m.handleKeyConfirm(msg)
	case ScreenResult:
		return m.handleKeyResult(msg)
	case ScreenRestoreConflict:
		return m.handleKeyRestoreConflict(msg)
	case ScreenRestore:
		return m.handleKeyRestore(msg)
	case ScreenDoctor:
		return m.handleKeyDoctor(msg)
	case ScreenConfig:
		return m.handleKeyConfig(msg)
	case ScreenConfigCategories:
		return m.handleKeyConfigCategories(msg)
	case ScreenConfigThresholds:
		return m.handleKeyConfigThresholds(msg)
	case ScreenQuarantineCleanup:
		return m.handleKeyQuarantineCleanup(msg)
	case ScreenLanguage:
		return m.handleKeyLanguage(msg)
	case ScreenError:
		return m.handleKeyError(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case ScreenMainMenu:
		return m.viewMainMenu()
	case ScreenDashboard:
		return m.viewDashboard()
	case ScreenCategories:
		return m.viewCategories()
	case ScreenWarnRecycleBin:
		return m.viewWarnRecycleBin()
	case ScreenDetails:
		return m.viewDetails()
	case ScreenConfirm:
		return m.viewConfirm()
	case ScreenCleaning:
		return m.viewCleaning()
	case ScreenResult:
		return m.viewResult()
	case ScreenRestoreConflict:
		return m.viewRestoreConflict()
	case ScreenError:
		return m.viewError()
	case ScreenRestore:
		return m.viewRestore()
	case ScreenDoctor:
		return m.viewDoctor()
	case ScreenConfig:
		return m.viewConfig()
	case ScreenConfigCategories:
		return m.viewConfigCategories()
	case ScreenConfigThresholds:
		return m.viewConfigThresholds()
	case ScreenQuarantineCleanup:
		return m.viewQuarantineCleanup()
	case ScreenLanguage:
		return m.viewLanguage()
	}
	return ""
}
