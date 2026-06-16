package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func testModel() model {
	m := initialModel()
	m.width = 80
	m.height = 24
	return m
}

func TestInitialModel(t *testing.T) {
	m := testModel() // keep initialModel here to test defaults
	if m.screen != ScreenMainMenu {
		t.Errorf("screen = %d, want MainMenu", m.screen)
	}
	if m.result != nil {
		t.Error("result should be nil")
	}
}

func TestUpdateScanDone(t *testing.T) {
	m := testModel()
	res := &types.ScanResult{
		Categories: []types.CategorySummary{
			{Category: "Temp", Risk: types.RiskSafe, Size: 100, Files: 1},
			{Category: "Downloads", Risk: types.RiskReview, Size: 200, Files: 2},
		},
		TotalSize:  300,
		SafeSize:   100,
		ReviewSize: 200,
	}
	msg := scanDoneMsg{res}
	newM, _ := m.Update(msg)
	mm := newM.(model)
	if mm.screen != ScreenDashboard {
		t.Errorf("screen = %d, want Dashboard", mm.screen)
	}
	if mm.result == nil {
		t.Fatal("result should be set")
	}
	if len(mm.categories) != 2 {
		t.Fatalf("categories = %d, want 2", len(mm.categories))
	}
	// All categories auto-selected
	if !mm.categories[0].selected {
		t.Error("safe category should be auto-selected")
	}
	if !mm.categories[1].selected {
		t.Error("review category should be auto-selected")
	}
}

func TestHandleKeyNavigation(t *testing.T) {
	m := testModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "A"}},
		{cat: types.CategorySummary{Category: "B"}},
		{cat: types.CategorySummary{Category: "C"}},
	}
	m.selectedIdx = 1

	// Up
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	mm := newM.(model)
	if mm.selectedIdx != 0 {
		t.Errorf("after up: selectedIdx = %d, want 0", mm.selectedIdx)
	}

	// Down
	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm2 := newM2.(model)
	if mm2.selectedIdx != 1 {
		t.Errorf("after down: selectedIdx = %d, want 1", mm2.selectedIdx)
	}

	// Down at boundary
	mm2.selectedIdx = 2
	newM3, _ := mm2.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm3 := newM3.(model)
	if mm3.selectedIdx != 2 {
		t.Errorf("at boundary: selectedIdx = %d, want 2", mm3.selectedIdx)
	}
}

func TestHandleKeyToggle(t *testing.T) {
	m := testModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "A"}, selected: true},
	}
	m.selectedIdx = 0

	// Space toggles selection on categories screen
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm := newM.(model)
	if mm.categories[0].selected {
		t.Error("space should toggle selection off")
	}
	// Toggle back on
	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm2 := newM2.(model)
	if !mm2.categories[0].selected {
		t.Error("space should toggle selection on again")
	}
}

func TestHandleKeyConfirmTransition(t *testing.T) {
	m := testModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "Temp", Risk: types.RiskSafe, Size: 100, Files: 1}, selected: true},
	}
	m.result = &types.ScanResult{TotalSize: 100}

	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm := newM.(model)
	if mm.screen != ScreenConfirm {
		t.Errorf("screen = %d, want Confirm", mm.screen)
	}
	if !strings.Contains(mm.confirmMsg, i18n.T("will_free")) {
		t.Errorf("confirmMsg should contain %q", i18n.T("will_free"))
	}
}

func TestViewDashboard(t *testing.T) {
	m := testModel()
	m.screen = ScreenDashboard
	m.result = &types.ScanResult{
		Categories: []types.CategorySummary{
			{Category: "Temp", Risk: types.RiskSafe, Size: 100, Files: 1},
			{Category: "Downloads", Risk: types.RiskReview, Size: 200, Files: 2},
		},
		TotalSize:  300,
		SafeSize:   100,
		ReviewSize: 200,
		DangerSize: 0,
	}
	out := m.View()
	if !strings.Contains(out, "Dashboard") && !strings.Contains(out, "Scan Results") {
		t.Error("dashboard view should contain 'Dashboard' or 'Scan Results'")
	}
	if !strings.Contains(out, "300 B") {
		t.Error("dashboard view should contain total size")
	}
	if !strings.Contains(out, "3 Files") {
		t.Error("dashboard view should contain file count")
	}
	if strings.Contains(out, "░") || strings.Contains(out, "█") {
		t.Error("dashboard bars should use fixed background segments")
	}
}

func TestViewDashboardAlignsLongCategoryNames(t *testing.T) {
	m := testModel()
	m.screen = ScreenDashboard
	m.result = &types.ScanResult{
		Categories: []types.CategorySummary{
			{Category: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", Risk: types.RiskSafe, Size: 100, Files: 1},
			{Category: "BBBBBBBBBBBBBBBBBBBBBBBBBBBB", Risk: types.RiskReview, Size: 50, Files: 1},
		},
		TotalSize:  150,
		SafeSize:   100,
		ReviewSize: 50,
	}
	out := ansi.Strip(m.View())
	lines := strings.Split(out, "\n")
	var sizeFields []int
	for _, line := range lines {
		switch {
		case strings.Contains(line, "AAAA"):
			sizeFields = append(sizeFields, strings.LastIndex(line, "B"))
		case strings.Contains(line, "BBBB"):
			sizeFields = append(sizeFields, strings.LastIndex(line, "B"))
		}
	}
	if len(sizeFields) != 2 {
		t.Fatalf("expected 2 category rows, got %d", len(sizeFields))
	}
	if sizeFields[0] != sizeFields[1] {
		t.Fatalf("size column should align, got %v", sizeFields)
	}
}

func TestViewCategories(t *testing.T) {
	m := testModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "Temp", Size: 100, Files: 1, Risk: types.RiskSafe}, selected: true},
		{cat: types.CategorySummary{Category: "Downloads", Size: 200, Files: 2, Risk: types.RiskReview}, selected: false},
	}
	m.selectedIdx = 0
	out := m.View()
	if !strings.Contains(out, "Cleanup Selection") {
		t.Error("view should contain 'Cleanup Selection'")
	}
	if !strings.Contains(out, "Temp") {
		t.Error("view should contain 'Temp'")
	}
	if !strings.Contains(out, "Downloads") {
		t.Error("view should contain 'Downloads'")
	}
}

func TestViewCategoriesAlignsLongNames(t *testing.T) {
	m := testModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "BBBBBBBBBBBBBBBBBBBBBBBBBBBB", Size: 100, Files: 1, Risk: types.RiskSafe}, selected: true},
		{cat: types.CategorySummary{Category: "CCCCCCCCCCCCCCCCCCCCCCCCCCCC", Size: 50, Files: 2, Risk: types.RiskReview}, selected: false},
	}
	m.selectedIdx = 0

	out := ansi.Strip(m.View())
	var sizeCols []int
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(line, "BBBB"):
			sizeCols = append(sizeCols, strings.LastIndex(line, "B"))
		case strings.Contains(line, "CCCC"):
			sizeCols = append(sizeCols, strings.LastIndex(line, "B"))
		}
	}
	if len(sizeCols) != 2 {
		t.Fatalf("expected 2 category rows, got %d", len(sizeCols))
	}
	if sizeCols[0] != sizeCols[1] {
		t.Fatalf("size columns should align, got %v", sizeCols)
	}
}

func TestViewConfirm(t *testing.T) {
	m := testModel()
	m.screen = ScreenConfirm
	m.confirmMsg = "Will free: 100 B"
	out := m.View()
	if !strings.Contains(out, "Confirm Cleanup") {
		t.Error("view should contain 'Confirm Cleanup'")
	}
}

func TestViewResult(t *testing.T) {
	m := testModel()
	m.screen = ScreenResult
	m.cleanResult = &types.CleanResult{RestoreID: "abc", Freed: 100, Files: 1}
	out := m.View()
	if !strings.Contains(out, "Cleanup Complete") {
		t.Error("view should contain 'Cleanup Complete'")
	}
	if !strings.Contains(out, "abc") {
		t.Error("view should contain restore ID")
	}
}

func TestViewRestoreConflict(t *testing.T) {
	m := testModel()
	m.screen = ScreenRestoreConflict
	m.conflicts = []string{`C:\a.txt`, `C:\b.txt`}
	out := m.View()
	if !strings.Contains(out, "File Conflicts") && !strings.Contains(out, "Конфликты файлов") {
		t.Error("view should contain 'File Conflicts'")
	}
	if !strings.Contains(out, "2") {
		t.Error("view should contain conflict count")
	}
}

func TestViewWarnRecycleBin(t *testing.T) {
	m := testModel()
	m.screen = ScreenWarnRecycleBin
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "Recycle Bin", Files: 15000}},
	}
	m.detailCat = 0
	out := m.View()
	if !strings.Contains(out, "Warning") {
		t.Error("view should contain 'Warning'")
	}
	if !strings.Contains(out, "Recycle Bin") {
		t.Error("view should contain 'Recycle Bin'")
	}
}

func TestUpdateErrMsg(t *testing.T) {
	m := testModel()
	msg := errMsg{fmt.Errorf("scan failed")}
	newM, cmd := m.Update(msg)
	mm := newM.(model)
	if mm.screen != ScreenError {
		t.Errorf("screen = %d, want Error", mm.screen)
	}
	if mm.err == nil {
		t.Error("err should be set")
	}
	if cmd != nil {
		t.Error("cmd should be nil, not tea.Quit")
	}
}

func TestViewError(t *testing.T) {
	m := testModel()
	m.screen = ScreenError
	m.err = fmt.Errorf("something broke")
	out := m.View()
	if !strings.Contains(out, "Error") {
		t.Error("view should contain 'Error'")
	}
	if !strings.Contains(out, "something broke") {
		t.Error("view should contain error message")
	}
}

func TestViewCleaning(t *testing.T) {
	m := testModel()
	m.screen = ScreenCleaning
	out := m.View()
	if !strings.Contains(out, "Cleaning") {
		t.Error("view should contain 'Cleaning'")
	}
}

func TestUpdateCleanDoneMsgErrorShowsScreenError(t *testing.T) {
	m := testModel()
	newM, _ := m.Update(cleanDoneMsg{result: nil, err: fmt.Errorf("clean failed")})
	mm := newM.(model)
	if mm.screen != ScreenError {
		t.Errorf("screen = %d, want ScreenError", mm.screen)
	}
}

func TestHandleKeyErrorEscReturnsToMainMenu(t *testing.T) {
	m := testModel()
	m.screen = ScreenError
	newM, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	mm := newM.(model)
	if mm.screen != ScreenMainMenu {
		t.Errorf("screen should be MainMenu, got %d", mm.screen)
	}
	if cmd != nil {
		t.Error("expected nil command")
	}
}

func TestBuildConfirmMessage(t *testing.T) {
	cats := []categoryItem{
		{cat: types.CategorySummary{Category: "Temp", Risk: types.RiskSafe, Size: 100, Files: 2}, selected: true},
		{cat: types.CategorySummary{Category: "Downloads", Risk: types.RiskReview, Size: 200, Files: 1}, selected: true},
	}
	res := &types.ScanResult{}
	msg := buildConfirmMessage(cats, res)
	if !strings.Contains(msg, i18n.T("will_free")) {
		t.Errorf("should contain %q", i18n.T("will_free"))
	}
	if !strings.Contains(msg, i18n.T("files")) {
		t.Errorf("should contain %q", i18n.T("files"))
	}
	if !strings.Contains(msg, i18n.T("risk_safe")) {
		t.Errorf("should contain %q", i18n.T("risk_safe"))
	}
	if !strings.Contains(msg, i18n.T("risk_review")) {
		t.Errorf("should contain %q", i18n.T("risk_review"))
	}
}

func TestViewMainMenu(t *testing.T) {
	m := testModel()
	out := m.View()
	if !strings.Contains(out, "Main Menu") {
		t.Error("view should contain 'Main Menu'")
	}
	if !strings.Contains(out, "Scan & Clean") {
		t.Error("view should contain 'Scan & Clean'")
	}
}

func TestHandleKeyMainMenuNavigation(t *testing.T) {
	m := testModel()
	m.screen = ScreenMainMenu
	m.selectedIdx = 0

	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm := newM.(model)
	if mm.selectedIdx != 1 {
		t.Errorf("after down: selectedIdx = %d, want 1", mm.selectedIdx)
	}

	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	mm2 := newM2.(model)
	if mm2.selectedIdx != 0 {
		t.Errorf("after up: selectedIdx = %d, want 0", mm2.selectedIdx)
	}

	// Boundary: pressing Down from last item should stay at last item
	m3 := testModel()
	m3.screen = ScreenMainMenu
	m3.selectedIdx = 3
	newM3, _ := m3.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm3 := newM3.(model)
	if mm3.selectedIdx != 3 {
		t.Errorf("at boundary down: selectedIdx = %d, want 3", mm3.selectedIdx)
	}
}

func TestViewRestoreEmpty(t *testing.T) {
	m := testModel()
	m.screen = ScreenRestore
	m.restoreEntries = []restoreEntry{}
	out := m.View()
	if !strings.Contains(out, "Quarantine is empty") && !strings.Contains(out, "Карантин пуст") {
		t.Error("view should show empty quarantine message")
	}
}

func TestViewRestoreWithIDs(t *testing.T) {
	m := testModel()
	m.screen = ScreenRestore
	m.restoreEntries = []restoreEntry{
		{id: "2025-06-09-143052", createdAt: time.Now(), totalSize: 1024, files: 2},
		{id: "2025-06-08-120000", createdAt: time.Now().AddDate(0, 0, -1), totalSize: 2048, files: 5},
	}
	m.restoreIdx = 1
	out := m.View()
	// New format: entries show relative dates and sizes, not raw IDs
	todayKey := "Today"
	yesterdayKey := "Yesterday"
	if strings.Contains(out, "Сегодня") {
		todayKey = "Сегодня"
		yesterdayKey = "Вчера"
	}
	if !strings.Contains(out, todayKey) {
		t.Errorf("view should contain %q for today's entry", todayKey)
	}
	if !strings.Contains(out, yesterdayKey) {
		t.Errorf("view should contain %q for yesterday's entry", yesterdayKey)
	}
}

func TestViewDoctor(t *testing.T) {
	m := testModel()
	m.screen = ScreenDoctor
	m.doctorChecks = []doctor.Check{
		{Name: "Test", Status: doctor.StatusPass, Detail: "ok"},
	}
	out := m.View()
	if !strings.Contains(out, "Health Check") && !strings.Contains(out, "Самопроверка") {
		t.Error("view should contain doctor screen title")
	}
	if !strings.Contains(out, "Test") {
		t.Error("view should contain check name")
	}
}

func TestViewConfig(t *testing.T) {
	m := testModel()
	m.screen = ScreenConfig
	out := m.View()
	if !strings.Contains(out, "Settings") && !strings.Contains(out, "Config") {
		t.Error("view should contain Settings or Config")
	}
}

func TestQuarantineSettingsDisabledLocksAutoCleanup(t *testing.T) {
	i18n.SetLanguage("ru")
	defer i18n.SetLanguage("en")

	m := testModel()
	m.screen = ScreenQuarantineSettings
	m.configCfg = config.Default()
	m.configCfg.QuarantineEnabled = false
	m.configCfg.QuarantineAutoCleanupDays = 7

	out := m.View()
	if !strings.Contains(out, "Выключено") {
		t.Error("view should show quarantine as disabled")
	}
	if !strings.Contains(out, "[Сначала включите карантин]") {
		t.Error("view should explain that quarantine must be enabled first")
	}

	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm := newM.(model)
	if mm.selectedIdx != 0 {
		t.Errorf("disabled auto-cleanup should not be selectable, got selectedIdx %d", mm.selectedIdx)
	}

	m.selectedIdx = 1
	newM, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm = newM.(model)
	if mm.configCfg.QuarantineAutoCleanupDays != 7 {
		t.Errorf("disabled auto-cleanup should not change, got %d", mm.configCfg.QuarantineAutoCleanupDays)
	}
}

func TestHandleKeyConfigPresetsStayOnScreen(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())

	m := testModel()
	m.screen = ScreenConfigPresets
	m.selectedIdx = 1
	m.lastConfigIdx = 3
	m.configCfg = config.Default()

	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm := newM.(model)
	if mm.screen != ScreenConfigPresets {
		t.Errorf("screen should stay on presets, got %d", mm.screen)
	}
	if mm.selectedIdx != 1 {
		t.Errorf("selectedIdx should stay on selected preset, got %d", mm.selectedIdx)
	}
	if mm.configCfg.ActivePreset != string(config.PresetStandard) {
		t.Errorf("expected standard preset, got %q", mm.configCfg.ActivePreset)
	}

	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm2 := newM2.(model)
	if mm2.screen != ScreenConfigPresets {
		t.Errorf("screen should stay on presets after space, got %d", mm2.screen)
	}
	if mm2.selectedIdx != 1 {
		t.Errorf("selectedIdx should stay on selected preset after space, got %d", mm2.selectedIdx)
	}
	if mm2.configCfg.ActivePreset != string(config.PresetStandard) {
		t.Errorf("expected standard preset after space, got %q", mm2.configCfg.ActivePreset)
	}
}

func TestViewLanguage(t *testing.T) {
	m := testModel()
	m.screen = ScreenLanguage
	out := m.View()
	if !strings.Contains(out, "Language") && !strings.Contains(out, "Язык") {
		t.Error("view should contain 'Language'")
	}
	if !strings.Contains(out, "English") {
		t.Error("view should contain 'English'")
	}
}

func TestHandleKeyLanguageSelect(t *testing.T) {
	m := testModel()
	m.screen = ScreenLanguage
	m.selectedIdx = 0

	// Move down to Russian
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm := newM.(model)
	if mm.selectedIdx != 1 {
		t.Errorf("after down: selectedIdx = %d, want 1", mm.selectedIdx)
	}

	// Apply Russian with Space — should stay on language screen
	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm2 := newM2.(model)
	if mm2.screen != ScreenLanguage {
		t.Errorf("after space apply: screen = %d, want Language", mm2.screen)
	}
	// Enter now applies the same as Space — stays on language screen
	newM3, _ := mm2.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm3 := newM3.(model)
	if mm3.screen != ScreenLanguage {
		t.Errorf("after enter apply: screen = %d, want Language", mm3.screen)
	}
	// Esc returns to Config
	newM4, _ := mm3.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	mm4 := newM4.(model)
	if mm4.screen != ScreenConfig {
		t.Errorf("after esc: screen = %d, want Config", mm4.screen)
	}
}

func TestConfigEscReturnsToMainMenuWithLastIdx(t *testing.T) {
	m := testModel()
	m.screen = ScreenMainMenu
	m.selectedIdx = 3 // Settings
	m.lastMainMenuIdx = 3
	m.configCfg = config.Default()
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mm := newM.(model)
	if mm.screen != ScreenMainMenu {
		t.Errorf("screen should be MainMenu, got %d", mm.screen)
	}
	if mm.selectedIdx != 3 {
		t.Errorf("selectedIdx should restore to 3, got %d", mm.selectedIdx)
	}
}

func TestConfigSubScreensRestoreLastConfigIdx(t *testing.T) {
	m := testModel()
	m.screen = ScreenLanguage
	m.lastConfigIdx = 1
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	mm := newM.(model)
	if mm.screen != ScreenConfig {
		t.Errorf("screen should be Config, got %d", mm.screen)
	}
	if mm.selectedIdx != 1 {
		t.Errorf("selectedIdx should restore to 1, got %d", mm.selectedIdx)
	}
}
