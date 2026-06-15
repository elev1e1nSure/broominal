package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

// Start — запускает TUI. Возвращает true если после выхода нужно перезапустить процесс.
func Start(version string) (bool, error) {
	cfg, _ := config.Load()
	if cfg != nil && cfg.Language != "" {
		i18n.SetLanguage(cfg.Language)
	}
	// Clean up leftover .old backup files from previous self-updates
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		entries, _ := os.ReadDir(exeDir)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".old") {
				_ = os.Remove(filepath.Join(exeDir, e.Name()))
			}
		}
	}
	m := initialModel()
	m.version = version
	if !doctor.IsAdmin() {
		m.screen = ScreenAdminPrompt
	} else if cfg == nil || cfg.Language == "" {
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
	} else {
		// Check for updates on startup
		m.screen = ScreenUpdating
		m.updateProgress = i18n.T("checking_updates")
		m.checkUpdateOnStartup = true
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}
	if fm, ok := finalModel.(model); ok {
		return fm.restartAfterUpdate, nil
	}
	return false, nil
}

func (m model) Init() tea.Cmd {
	if m.checkUpdateOnStartup {
		return func() tea.Msg {
			release, err := update.CheckForUpdates(m.version)
			return checkUpdateMsg{release, err}
		}
	}
	return nil
}

func (m model) appTitle(subtitle string) string {
	const maxSubtitleWidth = 40
	if len(subtitle) > maxSubtitleWidth {
		subtitle = subtitle[:maxSubtitleWidth-3] + "..."
	}
	title := fmt.Sprintf("broominal [%s] | %s", m.version, subtitle)
	return titleStyle.Render(title)
}

type scanDoneMsg struct {
	result *types.ScanResult
}

type scanProgressMsg struct{ completed int }

type errMsg struct {
	err error
}

type cleanDoneMsg struct {
	result *types.CleanResult
	err    error
}

type checkUpdateMsg struct {
	release *update.Release
	err     error
}

type downloadUpdateMsg struct {
	path string
	err  error
}

type restartMsg struct{}

// keyHandler is a screen-specific key event handler.
type keyHandler func(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd)

var keyHandlers = map[Screen]keyHandler{}

func registerKeyHandler(s Screen, h keyHandler) { keyHandlers[s] = h }

func init() {
	registerKeyHandler(ScreenMainMenu, model.handleKeyMainMenu)
	registerKeyHandler(ScreenDashboard, model.handleKeyDashboard)
	registerKeyHandler(ScreenCategories, model.handleKeyCategories)
	registerKeyHandler(ScreenWarnRecycleBin, model.handleKeyWarnRecycleBin)
	registerKeyHandler(ScreenWarnDuplicates, model.handleKeyWarnDuplicates)
	registerKeyHandler(ScreenCategoryInfo, model.handleKeyCategoryInfo)
	registerKeyHandler(ScreenConfirm, model.handleKeyConfirm)
	registerKeyHandler(ScreenCleaning, model.handleKeyCleaning)
	registerKeyHandler(ScreenResult, model.handleKeyResult)
	registerKeyHandler(ScreenRestoreConflict, model.handleKeyRestoreConflict)
	registerKeyHandler(ScreenRestore, model.handleKeyRestore)
	registerKeyHandler(ScreenConfirmDeleteQuarantine, model.handleKeyConfirmDeleteQuarantine)
	registerKeyHandler(ScreenConfirmDeleteAllQuarantine, model.handleKeyConfirmDeleteAllQuarantine)
	registerKeyHandler(ScreenDoctor, model.handleKeyDoctor)
	registerKeyHandler(ScreenConfig, model.handleKeyConfig)
	registerKeyHandler(ScreenConfigPresets, model.handleKeyConfigPresets)
	registerKeyHandler(ScreenQuarantineSettings, model.handleKeyQuarantineSettings)
	registerKeyHandler(ScreenLanguage, model.handleKeyLanguage)
	registerKeyHandler(ScreenAdminPrompt, model.handleKeyAdminPrompt)
	registerKeyHandler(ScreenError, model.handleKeyError)
	registerKeyHandler(ScreenUpdateAvailable, model.handleKeyUpdateAvailable)
	registerKeyHandler(ScreenUpdating, model.handleKeyUpdating)
	registerKeyHandler(ScreenNoUpdate, model.handleKeyNoUpdate)
	registerKeyHandler(ScreenPathConfirm, model.handleKeyPathConfirm)
	registerKeyHandler(ScreenPathResult, model.handleKeyPathResult)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case scanProgressMsg:
		m.scanCompleted = msg.completed
		if m.scanCh != nil {
			return m, func() tea.Msg { return <-m.scanCh }
		}
		return m, nil

	case scanDoneMsg:
		m.scanCh = nil
		m.result = msg.result
		m.categories = make([]categoryItem, 0, len(msg.result.Categories))
		for _, c := range msg.result.Categories {
			m.categories = append(m.categories, categoryItem{cat: c, selected: true})
		}
		m.screen = ScreenDashboard
		return m, nil

	case errMsg:
		m.err = msg.err
		m.screen = ScreenError
		return m, nil

	case cleanDoneMsg:
		if msg.err != nil {
			if errors.Is(msg.err, context.Canceled) {
				m.screen = ScreenMainMenu
				m.selectedIdx = m.lastMainMenuIdx
				return m, nil
			}
			m.err = msg.err
			m.screen = ScreenError
			return m, nil
		}
		m.cleanResult = msg.result
		m.screen = ScreenResult
		return m, nil

	case checkUpdateMsg:
		if msg.err != nil || msg.release == nil {
			// No internet, API error, or no update available
			if m.updateFromConfig {
				m.screen = ScreenNoUpdate
			} else {
				m.screen = ScreenMainMenu
			}
			return m, nil
		}
		m.updateAvailableRelease = msg.release
		m.screen = ScreenUpdateAvailable
		return m, nil

	case downloadUpdateMsg:
		if m.updateCancelled {
			return m, nil
		}
		if msg.err != nil {
			m.updateError = msg.err
			m.updateProgress = i18n.T("download_failed")
			return m, nil
		}
		m.updateProgress = i18n.T("installing_update")
		return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
			err := update.InstallUpdate(msg.path)
			return installUpdateMsg{err}
		})

	case installUpdateMsg:
		if m.updateCancelled {
			return m, nil
		}
		if msg.err != nil {
			m.updateError = msg.err
			m.updateProgress = i18n.T("install_failed")
			return m, nil
		}
		m.updateProgress = i18n.T("update_restarting")
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg { return restartMsg{} })

	case restartMsg:
		m.restartAfterUpdate = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// normalizeKey maps Russian ЙЦУКЕН layout keystrokes to their Latin QWERTY equivalents.
// This lets users navigate the TUI without switching keyboard layouts.
func normalizeKey(msg tea.KeyMsg) tea.KeyMsg {
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 {
		return msg
	}
	// ЙЦУКЕН — QWERTY
	m := map[rune]rune{
		'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
		'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
		'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm', 'б': ',', 'ю': '.',
		'Й': 'Q', 'Ц': 'W', 'У': 'E', 'К': 'R', 'Е': 'T', 'Н': 'Y', 'Г': 'U', 'Ш': 'I', 'Щ': 'O', 'З': 'P', 'Х': '{', 'Ъ': '}',
		'Ф': 'A', 'Ы': 'S', 'В': 'D', 'А': 'F', 'П': 'G', 'Р': 'H', 'О': 'J', 'Л': 'K', 'Д': 'L', 'Ж': ':', 'Э': '"',
		'Я': 'Z', 'Ч': 'X', 'С': 'C', 'М': 'V', 'И': 'B', 'Т': 'N', 'Ь': 'M', 'Б': '<', 'Ю': '>',
	}
	out := make([]rune, len(msg.Runes))
	for i, r := range msg.Runes {
		if v, ok := m[r]; ok {
			out[i] = v
		} else {
			out[i] = r
		}
	}
	msg.Runes = out
	return msg
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	msg = normalizeKey(msg)
	if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
		return m, tea.Quit
	}
	if h, ok := keyHandlers[m.screen]; ok {
		return h(m, msg)
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
	case ScreenWarnDuplicates:
		return m.viewWarnDuplicates()
	case ScreenCategoryInfo:
		return m.viewCategoryInfo()
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
	case ScreenConfirmDeleteQuarantine:
		return m.viewConfirmDeleteQuarantine()
	case ScreenConfirmDeleteAllQuarantine:
		return m.viewConfirmDeleteAllQuarantine()
	case ScreenDoctor:
		return m.viewDoctor()
	case ScreenConfig:
		return m.viewConfig()
	case ScreenConfigPresets:
		return m.viewConfigPresets()
	case ScreenQuarantineSettings:
		return m.viewQuarantineSettings()
	case ScreenLanguage:
		return m.viewLanguage()
	case ScreenAdminPrompt:
		return m.viewAdminPrompt()
	case ScreenUpdateAvailable:
		return m.viewUpdateAvailable()
	case ScreenUpdating:
		return m.viewUpdating()
	case ScreenNoUpdate:
		return m.viewNoUpdate()
	case ScreenPathConfirm:
		return m.viewPathConfirm()
	case ScreenPathResult:
		return m.viewPathResult()
	}
	return ""
}
