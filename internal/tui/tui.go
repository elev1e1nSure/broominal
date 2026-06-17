package tui

import (
	"context"
	"errors"

	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

// Start launches the TUI. The boolean return tells the caller whether to
// re-exec itself — true after a self-update, so the new binary replaces the
// running one before the process exits.
func Start(version string) (bool, error) {
	cfg, _ := config.Load()
	if cfg != nil && cfg.Language != "" {
		i18n.SetLanguage(cfg.Language)
	}
	// Past self-updates can leave .old binaries behind if the post-update
	// restart was interrupted; sweep them here so the next update can
	// rename the current binary to .old without colliding.
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
		// No language saved yet — detect from Windows locale, persist, and skip the
		// selection screen. The user can change it later from Settings → Language.
		lang := i18n.DetectFromWindowsLocale()
		i18n.SetLanguage(lang)
		if cfg == nil {
			cfg = config.Default()
		}
		cfg.Language = lang
		if err := config.Save(cfg); err != nil {
			slog.Warn("tui: failed to save language config", "error", err)
		}
		m.screen = ScreenUpdating
		m.updateProgress = i18n.T("checking_updates")
		m.checkUpdateOnStartup = true
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
		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				release, err := update.CheckForUpdates(m.version)
				return checkUpdateMsg{release, err}
			},
		)
	}
	return nil
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

type cleanProgressMsg struct {
	p types.Progress
}

type quarantineDeleteDoneMsg struct {
	count int
	freed int64
	err   error
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

// viewHandler is a screen-specific view renderer.
type viewHandler func(m model) string

var viewHandlers = map[Screen]viewHandler{}

func registerViewHandler(s Screen, h viewHandler) { viewHandlers[s] = h }

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
	registerKeyHandler(ScreenDeletingQuarantine, model.handleKeyDeletingQuarantine)
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
	registerKeyHandler(ScreenPathSettings, model.handleKeyPathSettings)

	registerViewHandler(ScreenMainMenu, model.viewMainMenu)
	registerViewHandler(ScreenDashboard, model.viewDashboard)
	registerViewHandler(ScreenCategories, model.viewCategories)
	registerViewHandler(ScreenWarnRecycleBin, model.viewWarnRecycleBin)
	registerViewHandler(ScreenWarnDuplicates, model.viewWarnDuplicates)
	registerViewHandler(ScreenCategoryInfo, model.viewCategoryInfo)
	registerViewHandler(ScreenConfirm, model.viewConfirm)
	registerViewHandler(ScreenCleaning, model.viewCleaning)
	registerViewHandler(ScreenResult, model.viewResult)
	registerViewHandler(ScreenRestoreConflict, model.viewRestoreConflict)
	registerViewHandler(ScreenError, model.viewError)
	registerViewHandler(ScreenRestore, model.viewRestore)
	registerViewHandler(ScreenConfirmDeleteQuarantine, model.viewConfirmDeleteQuarantine)
	registerViewHandler(ScreenConfirmDeleteAllQuarantine, model.viewConfirmDeleteAllQuarantine)
	registerViewHandler(ScreenDeletingQuarantine, model.viewDeletingQuarantine)
	registerViewHandler(ScreenDoctor, model.viewDoctor)
	registerViewHandler(ScreenConfig, model.viewConfig)
	registerViewHandler(ScreenConfigPresets, model.viewConfigPresets)
	registerViewHandler(ScreenQuarantineSettings, model.viewQuarantineSettings)
	registerViewHandler(ScreenLanguage, model.viewLanguage)
	registerViewHandler(ScreenAdminPrompt, model.viewAdminPrompt)
	registerViewHandler(ScreenUpdateAvailable, model.viewUpdateAvailable)
	registerViewHandler(ScreenUpdating, model.viewUpdating)
	registerViewHandler(ScreenNoUpdate, model.viewNoUpdate)
	registerViewHandler(ScreenPathSettings, model.viewPathSettings)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 6
		return m, nil

	case scanProgressMsg:
		m.scanCompleted = msg.completed
		if m.scanCh != nil {
			return m, func() tea.Msg { return <-m.scanCh }
		}
		return m, nil

	case cleanProgressMsg:
		m.cleanProgress = &msg.p
		// EMA smoothing for throughput (alpha=0.15, ~3s window).
		if m.smoothBytesPerSec == 0 {
			elapsed := time.Since(msg.p.StartedAt).Seconds()
			if elapsed > 0.1 && msg.p.Bytes > 0 {
				m.smoothBytesPerSec = float64(msg.p.Bytes) / elapsed
			}
		} else {
			elapsed := time.Since(msg.p.StartedAt).Seconds()
			if elapsed > 0.1 && msg.p.Bytes > 0 {
				instant := float64(msg.p.Bytes) / elapsed
				m.smoothBytesPerSec = 0.15*instant + 0.85*m.smoothBytesPerSec
			}
		}
		// EMA smoothing for ETA based on item count (matches progress bar), alpha=0.15.
		rawETA := msg.p.ETA()
		if rawETA > 0 {
			if m.smoothETA == 0 {
				m.smoothETA = rawETA
			} else {
				m.smoothETA = time.Duration(0.15*float64(rawETA) + 0.85*float64(m.smoothETA))
			}
		} else {
			m.smoothETA = 0
		}
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

	case quarantineDeleteDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.screen = ScreenError
			return m, nil
		}
		if m.deleteAllQuarantine {
			for _, entry := range m.restoreEntries {
				m.deletedQuarantines[entry.id] = true
			}
		} else if m.restoreIdx < len(m.restoreEntries) {
			m.deletedQuarantines[m.restoreEntries[m.restoreIdx].id] = true
		}
		m.restoreEntries = m.reloadEntriesFiltered()
		m.restoreIdx = 0
		m.screen = ScreenRestore
		return m, nil

	case cleanDoneMsg:
		m.scanCh = nil
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
			// nil release with no error is GitHub's "you are up to date" reply;
			// treat any failure the same way so the user sees a single "no
			// update" path rather than a scary error dialog for offline runs.
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

	case clearRestoreResultMsg:
		m.restoreResult = ""
		return m, nil

	case restartMsg:
		m.restartAfterUpdate = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.updateTick++
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// normalizeKey maps Russian JCUKEN layout keystrokes to their Latin QWERTY equivalents.
// This lets users navigate the TUI without switching keyboard layouts.
func normalizeKey(msg tea.KeyMsg) tea.KeyMsg {
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 {
		return msg
	}
	// Russian JCUKEN → QWERTY physical-key positions, so a user whose layout is set
	// to Russian can still navigate the TUI (Up/Down, y/n prompts, etc.)
	// without switching layouts first.
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
	if h, ok := viewHandlers[m.screen]; ok {
		return h(m)
	}
	return ""
}
