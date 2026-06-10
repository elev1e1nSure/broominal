package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

// Screen — текущий экран
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenDashboard
	ScreenCategories
	ScreenWarnRecycleBin
	ScreenWarnDuplicates
	ScreenCategoryInfo
	// Quarantine deletion confirmations
	ScreenConfirmDeleteQuarantine
	ScreenConfirmDeleteAllQuarantine
	ScreenConfirm
	ScreenCleaning
	ScreenResult
	ScreenRestoreConflict
	ScreenError
	ScreenRestore
	ScreenDoctor
	ScreenConfig
	ScreenConfigPresets
	ScreenQuarantineSettings
	ScreenQuarantineCleanup
	ScreenQuarantineCleaning
	ScreenLanguage
	ScreenAdminPrompt
	ScreenUpdateAvailable
	ScreenUpdating
	ScreenNoUpdate
	ScreenPathConfirm
	ScreenPathResult
)

type restoreEntry struct {
	id        string
	createdAt time.Time
	totalSize int64
	files     int
	label     string
}

type model struct {
	screen                Screen
	result                *types.ScanResult
	categories            []categoryItem
	selectedIdx           int
	detailCat             int
	confirmMsg            string
	cleanResult           *types.CleanResult
	spinner               spinner.Model
	err                   error
	width                 int
	height                int
	conflicts             []string
	restoreForceOverwrite bool
	version               string
	// Restore screen
	restoreEntries []restoreEntry
	restoreIdx     int
	restoreResult  string
	// Doctor screen
	doctorChecks    []doctor.Check
	doctorFixResult string
	// Config screen
	configCfg             *config.Config
	lastConfigIdx         int
	quarantineSettingsMsg string
	// Cleanup screen
	cleanupResult string
	// Admin prompt
	adminPromptIdx int
	// Update screen
	updateAvailableRelease *update.Release
	updateError            error
	updateProgress         string
	checkUpdateOnStartup   bool
	updateFromConfig       bool
	// Path screen
	pathConfirmIdx int
	pathOperation  string
	pathResultMsg  string
	// Scan progress
	scanCh        chan tea.Msg
	scanCompleted int
	scanTotal     int
	// Menu position memory
	lastMainMenuIdx int
	// Set to true when update is installed and process should restart
	restartAfterUpdate bool
}

type categoryItem struct {
	cat      types.CategorySummary
	selected bool
}

func (c categoryItem) FilterValue() string { return c.cat.Category }

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#60a5fa"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff"))
	safeStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4ade80"))
	reviewStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fbbf24"))
	dangerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f87171"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca3af"))
	disabledStyle = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("#6b7280"))
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d6b15f"))
	valueStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e5e7eb"))
	backgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("#111827"))
)

func keyHint(k, desc string) string {
	return keyStyle.Render("["+k+"]") + " " + desc
}

func footer(hints ...string) string {
	return mutedStyle.Render(strings.Join(hints, "  "))
}

// clampWindow returns start/end indices so that idx is visible within a window of size visible.
func clampWindow(idx, total, visible int) (int, int) {
	if total <= visible || visible <= 0 {
		return 0, total
	}
	start := idx - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > total {
		end = total
		start = end - visible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	return model{
		screen:  ScreenMainMenu,
		spinner: s,
	}
}
