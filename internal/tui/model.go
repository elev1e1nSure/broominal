package tui

import (
"fmt"
"strings"

"github.com/charmbracelet/bubbles/list"
"github.com/charmbracelet/bubbles/spinner"
lipgloss "github.com/charmbracelet/lipgloss"
"github.com/elev1e1nSure/broominal/pkg/config"
"github.com/elev1e1nSure/broominal/pkg/doctor"
"github.com/elev1e1nSure/broominal/pkg/types"
"github.com/elev1e1nSure/broominal/pkg/util"
)

// Screen — текущий экран
type Screen int

const (
ScreenMainMenu Screen = iota
ScreenDashboard
ScreenCategories
ScreenWarnRecycleBin
ScreenDetails
ScreenConfirm
ScreenCleaning
ScreenResult
ScreenRestoreConflict
ScreenError
ScreenRestore
ScreenDoctor
ScreenConfig
ScreenConfigCategories
ScreenConfigThresholds
ScreenQuarantineCleanup
ScreenLanguage
)

type model struct {
screen                Screen
result                *types.ScanResult
categories            []categoryItem
selectedIdx           int
detailCat             int
detailList            list.Model
confirmMsg            string
cleanResult           *types.CleanResult
spinner               spinner.Model
err                   error
width                 int
height                int
dryRun                bool
conflicts             []string
restoreForceOverwrite bool
// Restore screen
restoreIDs    []string
restoreIdx    int
restoreResult string
// Doctor screen
doctorChecks []doctor.Check
// Config screen
configView           string
configCategories     []configCategoryItem
configThresholds     []configThresholdItem
configCfg            *config.Config
// Cleanup screen
cleanupResult        string
quarantineCleanupAll bool
}

type configCategoryItem struct {
name    string
enabled bool
}

type configThresholdItem struct {
labelKey string
value    int
min      int
step     int
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
keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fbbf24"))
valueStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e5e7eb"))
)

func keyHint(k, desc string) string {
return keyStyle.Render("["+k+"]") + " " + desc
}

func footer(hints ...string) string {
return mutedStyle.Render(strings.Join(hints, "  "))
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

type detailItem struct {
item types.Item
}

func (d detailItem) FilterValue() string { return d.item.Path }

func (d detailItem) Title() string { return d.item.Path }
func (d detailItem) Description() string {
return fmt.Sprintf("%s  %s", util.FormatSize(d.item.Size), d.item.Risk)
}