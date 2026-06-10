package tui

import (
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
	case ScreenWarnDuplicates:
		return m.handleKeyWarnDuplicates(msg)
	case ScreenCategoryInfo:
		return m.handleKeyCategoryInfo(msg)
	case ScreenConfirm:
		return m.handleKeyConfirm(msg)
	case ScreenResult:
		return m.handleKeyResult(msg)
	case ScreenRestoreConflict:
		return m.handleKeyRestoreConflict(msg)
	case ScreenRestore:
		return m.handleKeyRestore(msg)
	case ScreenConfirmDeleteQuarantine:
		return m.handleKeyConfirmDeleteQuarantine(msg)
	case ScreenConfirmDeleteAllQuarantine:
		return m.handleKeyConfirmDeleteAllQuarantine(msg)
	case ScreenDoctor:
		return m.handleKeyDoctor(msg)
	case ScreenConfig:
		return m.handleKeyConfig(msg)
	case ScreenConfigPresets:
		return m.handleKeyConfigPresets(msg)
	case ScreenQuarantineSettings:
		return m.handleKeyQuarantineSettings(msg)
	case ScreenQuarantineCleanup:
		return m.handleKeyQuarantineCleanup(msg)
	case ScreenLanguage:
		return m.handleKeyLanguage(msg)
	case ScreenAdminPrompt:
		return m.handleKeyAdminPrompt(msg)
	case ScreenError:
		return m.handleKeyError(msg)
	case ScreenUpdateAvailable:
		return m.handleKeyUpdateAvailable(msg)
	case ScreenUpdating:
		return m.handleKeyUpdating(msg)
	case ScreenNoUpdate:
		return m.handleKeyNoUpdate(msg)
	case ScreenPathConfirm:
		return m.handleKeyPathConfirm(msg)
	case ScreenPathResult:
		return m.handleKeyPathResult(msg)
	}
	return m, nil
}

func (m model) View() string {
	var content string
	switch m.screen {
	case ScreenMainMenu:
		content = m.viewMainMenu()
	case ScreenDashboard:
		content = m.viewDashboard()
	case ScreenCategories:
		content = m.viewCategories()
	case ScreenWarnRecycleBin:
		content = m.viewWarnRecycleBin()
	case ScreenWarnDuplicates:
		content = m.viewWarnDuplicates()
	case ScreenCategoryInfo:
		content = m.viewCategoryInfo()
	case ScreenConfirm:
		content = m.viewConfirm()
	case ScreenCleaning:
		content = m.viewCleaning()
	case ScreenResult:
		content = m.viewResult()
	case ScreenRestoreConflict:
		content = m.viewRestoreConflict()
	case ScreenError:
		content = m.viewError()
	case ScreenRestore:
		content = m.viewRestore()
	case ScreenConfirmDeleteQuarantine:
		content = m.viewConfirmDeleteQuarantine()
	case ScreenConfirmDeleteAllQuarantine:
		content = m.viewConfirmDeleteAllQuarantine()
	case ScreenDoctor:
		content = m.viewDoctor()
	case ScreenConfig:
		content = m.viewConfig()
	case ScreenConfigPresets:
		content = m.viewConfigPresets()
	case ScreenQuarantineSettings:
		content = m.viewQuarantineSettings()
	case ScreenQuarantineCleanup:
		content = m.viewQuarantineCleanup()
	case ScreenQuarantineCleaning:
		content = m.viewQuarantineCleaning()
	case ScreenLanguage:
		content = m.viewLanguage()
	case ScreenAdminPrompt:
		content = m.viewAdminPrompt()
	case ScreenUpdateAvailable:
		content = m.viewUpdateAvailable()
	case ScreenUpdating:
		content = m.viewUpdating()
	case ScreenNoUpdate:
		content = m.viewNoUpdate()
	case ScreenPathConfirm:
		content = m.viewPathConfirm()
	case ScreenPathResult:
		content = m.viewPathResult()
	}
	return backgroundStyle.Width(m.width).Height(m.height).Render(content)
}
