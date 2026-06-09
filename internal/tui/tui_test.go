package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elev1e1nSure/broominal/pkg/doctor"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if m.screen != ScreenMainMenu {
		t.Errorf("screen = %d, want MainMenu", m.screen)
	}
	if m.result != nil {
		t.Error("result should be nil")
	}
}

func TestUpdateScanDone(t *testing.T) {
	m := initialModel()
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
	m := initialModel()
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
	m := initialModel()
	m.screen = ScreenCategories
	m.categories = []categoryItem{
		{cat: types.CategorySummary{Category: "A"}, selected: true},
	}
	m.selectedIdx = 0

	// Space no longer toggles selection on categories screen
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm := newM.(model)
	if !mm.categories[0].selected {
		t.Error("space should not toggle selection off")
	}
}

func TestHandleKeyConfirmTransition(t *testing.T) {
	m := initialModel()
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
	if !strings.Contains(mm.confirmMsg, "Will free") {
		t.Error("confirmMsg should contain 'Will free'")
	}
}

func TestViewDashboard(t *testing.T) {
	m := initialModel()
	m.screen = ScreenDashboard
	m.result = &types.ScanResult{
		TotalSize:  300,
		SafeSize:   100,
		ReviewSize: 200,
		DangerSize: 0,
	}
	out := m.View()
	if !strings.Contains(out, "Dashboard") {
		t.Error("dashboard view should contain 'Dashboard'")
	}
	if !strings.Contains(out, "Total found") {
		t.Error("dashboard view should contain 'Total found'")
	}
}

func TestViewCategories(t *testing.T) {
	m := initialModel()
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

func TestViewConfirm(t *testing.T) {
	m := initialModel()
	m.screen = ScreenConfirm
	m.confirmMsg = "Will free: 100 B"
	out := m.View()
	if !strings.Contains(out, "Confirm Cleanup") {
		t.Error("view should contain 'Confirm Cleanup'")
	}
}

func TestViewResult(t *testing.T) {
	m := initialModel()
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
	m := initialModel()
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
	m := initialModel()
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
	m := initialModel()
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
	m := initialModel()
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
	m := initialModel()
	m.screen = ScreenCleaning
	out := m.View()
	if !strings.Contains(out, "Cleaning") {
		t.Error("view should contain 'Cleaning'")
	}
}

func TestUpdateCleanDoneMsgErrorShowsScreenError(t *testing.T) {
	m := initialModel()
	newM, _ := m.Update(cleanDoneMsg{result: nil, err: fmt.Errorf("clean failed")})
	mm := newM.(model)
	if mm.screen != ScreenError {
		t.Errorf("screen = %d, want ScreenError", mm.screen)
	}
}

func TestHandleKeyErrorReturnsToMainMenu(t *testing.T) {
	m := initialModel()
	m.screen = ScreenError
	newM, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
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
	if !strings.Contains(msg, "Will free") {
		t.Error("should contain 'Will free'")
	}
	if !strings.Contains(msg, "Files") {
		t.Error("should contain 'Files'")
	}
	if !strings.Contains(msg, "Safe") {
		t.Error("should contain 'Safe'")
	}
	if !strings.Contains(msg, "Review") {
		t.Error("should contain 'Review'")
	}
}

func TestViewMainMenu(t *testing.T) {
	m := initialModel()
	out := m.View()
	if !strings.Contains(out, "Main Menu") {
		t.Error("view should contain 'Main Menu'")
	}
	if !strings.Contains(out, "Scan & Clean") {
		t.Error("view should contain 'Scan & Clean'")
	}
}

func TestHandleKeyMainMenuNavigation(t *testing.T) {
	m := initialModel()
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
}

func TestViewRestoreEmpty(t *testing.T) {
	m := initialModel()
	m.screen = ScreenRestore
	m.restoreEntries = []restoreEntry{}
	out := m.View()
	if !strings.Contains(out, "No backups") {
		t.Error("view should show empty message")
	}
}

func TestViewRestoreWithIDs(t *testing.T) {
	m := initialModel()
	m.screen = ScreenRestore
	m.restoreEntries = []restoreEntry{
		{id: "2025-06-09-143052", createdAt: time.Now(), totalSize: 100, files: 2},
		{id: "2025-06-08-120000", createdAt: time.Now().Add(-24 * time.Hour), totalSize: 200, files: 5},
	}
	m.restoreIdx = 1
	out := m.View()
	if !strings.Contains(out, "2025-06-09") {
		t.Error("view should contain first entry date")
	}
	if !strings.Contains(out, "2025-06-08") {
		t.Error("view should contain second entry date")
	}
}

func TestViewDoctor(t *testing.T) {
	m := initialModel()
	m.screen = ScreenDoctor
	m.doctorChecks = []doctor.Check{
		{Name: "Test", Status: doctor.StatusPass, Detail: "ok"},
	}
	out := m.View()
	if !strings.Contains(out, "Diagnostics") && !strings.Contains(out, "Диагностика") {
		t.Error("view should contain 'Diagnostics'")
	}
	if !strings.Contains(out, "Test") {
		t.Error("view should contain check name")
	}
}

func TestViewConfig(t *testing.T) {
	m := initialModel()
	m.screen = ScreenConfig
	out := m.View()
	if !strings.Contains(out, "Settings") && !strings.Contains(out, "Config") {
		t.Error("view should contain Settings or Config")
	}
}

func TestViewQuarantineCleanup(t *testing.T) {
	m := initialModel()
	m.screen = ScreenQuarantineCleanup
	out := m.View()
	if !strings.Contains(out, "Backup Cleanup") && !strings.Contains(out, "Очистка резервных копий") {
		t.Error("view should contain 'Backup Cleanup'")
	}
}

func TestViewLanguage(t *testing.T) {
	m := initialModel()
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
	m := initialModel()
	m.screen = ScreenLanguage
	m.selectedIdx = 0

	// Move down to Russian
	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	mm := newM.(model)
	if mm.selectedIdx != 1 {
		t.Errorf("after down: selectedIdx = %d, want 1", mm.selectedIdx)
	}

	// Select Russian — should go to main menu
	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm2 := newM2.(model)
	if mm2.screen != ScreenMainMenu {
		t.Errorf("after select: screen = %d, want MainMenu", mm2.screen)
	}
}

func TestHandleKeyMReturnsToMainMenu(t *testing.T) {
	screens := []Screen{
		ScreenResult, ScreenError, ScreenRestore, ScreenDoctor,
		ScreenConfig, ScreenQuarantineCleanup, ScreenLanguage,
	}
	for _, sc := range screens {
		m := initialModel()
		m.screen = sc
		m.selectedIdx = 5
		newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		mm := newM.(model)
		if mm.screen != ScreenMainMenu {
			t.Errorf("screen %d: M should go to MainMenu, got %d", sc, mm.screen)
		}
		if mm.selectedIdx != 0 {
			t.Errorf("screen %d: selectedIdx should reset to 0, got %d", sc, mm.selectedIdx)
		}
	}
}
