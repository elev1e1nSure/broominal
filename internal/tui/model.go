package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/update"
)

// Screen identifies the currently visible TUI screen. The bubble-tea Update
// loop dispatches key events and renders views through screen-specific
// handlers registered in tui.go's init().
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
	ScreenDeletingQuarantine
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
	ScreenLanguage
	ScreenAdminPrompt
	ScreenUpdateAvailable
	ScreenUpdating
	ScreenNoUpdate
	ScreenPathSettings
)

type restoreEntry struct {
	id         string
	createdAt  time.Time
	totalSize  int64
	files      int
	categories []string
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
	progress              progress.Model
	err                   error
	width                 int
	height                int
	conflicts             []string
	restoreForceOverwrite bool
	version               string
	// Restore screen
	restoreEntries      []restoreEntry
	restoreIdx          int
	restoreResult       string
	deleteAllQuarantine bool
	// Doctor screen
	doctorChecks     []doctor.Check
	doctorFixResult  string
	doctorPendingFix string // fix key awaiting user confirmation
	// Config screen
	configCfg             *config.Config
	lastConfigIdx         int
	quarantineSettingsMsg string
	// Admin prompt
	adminPromptIdx int
	// Update screen
	updateAvailableRelease *update.Release
	updateError            error
	updateProgress         string
	checkUpdateOnStartup   bool
	updateFromConfig       bool
	// Scan progress
	scanCh        chan tea.Msg
	scanCompleted int
	scanTotal     int
	// Menu position memory
	lastMainMenuIdx int
	// Set to true when update is installed and process should restart
	restartAfterUpdate bool
	// Cancellation for long-running operations
	cleanCtxCancel  context.CancelFunc
	updateCancelled bool
}

type categoryItem struct {
	cat      types.CategorySummary
	selected bool
}

func (c categoryItem) FilterValue() string { return c.cat.Category }

// Styles
var (
	// Base Colors (Grayscale/White)
	colorBorder  = lipgloss.Color("#6b7280") // Gray 500
	colorFg      = lipgloss.Color("#f3f4f6") // Gray 100
	colorAccent  = lipgloss.Color("#e5e7eb") // Gray 200 (Almost white)
	colorSafe    = lipgloss.Color("#a6da95") // Green
	colorWarning = lipgloss.Color("#eed49f") // Yellow
	colorDanger  = lipgloss.Color("#ed8796") // Red
	colorMuted   = lipgloss.Color("#9ca3af") // Gray 400
	colorTrack   = lipgloss.Color("#374151") // Gray 700

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff"))
	safeStyle     = lipgloss.NewStyle().Bold(true).Foreground(colorSafe)
	reviewStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorWarning)
	dangerStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorDanger)
	mutedStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	disabledStyle = lipgloss.NewStyle().Faint(true).Foreground(colorMuted)
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	keyDescStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	valueStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorFg)

	// Layout Styles
	appFrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorBorder).
			Padding(1, 0)

	footerStyle = lipgloss.NewStyle().
			MarginTop(1)

	barTrackStyle = lipgloss.NewStyle().Background(colorTrack)
)

func keyHint(k, desc string) string {
	return keyStyle.Render("["+k+"]") + " " + keyDescStyle.Render(desc)
}

func footer(hints ...string) string {
	return strings.Join(hints, "  ")
}

func barFillStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(color)
}

func (m model) appFrame(title, content, foot string) string {
	// Don't render frame if dimensions are too small
	if m.width < 10 || m.height < 10 {
		return content
	}

	var head string
	var headHeight int

	if title != "" {
		head = headerStyle.Width(m.width - 4).Render(titleStyle.Render(title))
		headHeight = lipgloss.Height(head)
	}

	footRender := footerStyle.Width(m.width - 4).Render(foot)
	footHeight := lipgloss.Height(footRender)

	targetContentHeight := m.height - 2 - headHeight - footHeight // -2 for frame top/bottom border
	if targetContentHeight < 0 {
		targetContentHeight = 0
	}

	// Pad content to fill the remaining height so footer sticks to the bottom
	paddedContent := lipgloss.NewStyle().
		Height(targetContentHeight).
		PaddingTop(1).
		Render(content)

	var uiBlocks []string
	if head != "" {
		uiBlocks = append(uiBlocks, head)
	}
	uiBlocks = append(uiBlocks, paddedContent, footRender)

	ui := lipgloss.JoinVertical(lipgloss.Left, uiBlocks...)

	return appFrameStyle.
		Width(m.width - 2). // -2 for borders
		Height(m.height - 2).
		Render(ui)
}

func truncateDisplay(s string, width int) string {
	return ansi.Truncate(s, width, "…")
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
	s.Style = lipgloss.NewStyle().Foreground(colorAccent)

	p := progress.New(
		progress.WithSolidFill(string(colorAccent)),
		progress.WithoutPercentage(),
	)

	return model{
		screen:   ScreenMainMenu,
		spinner:  s,
		progress: p,
	}
}
