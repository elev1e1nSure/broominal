package tui

import (
	"fmt"
	"strings"
	"testing"

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
	if mm.screen != ScreenCategories {
		t.Errorf("screen = %d, want Categories", mm.screen)
	}
	if mm.result == nil {
		t.Fatal("result should be set")
	}
	if len(mm.categories) != 2 {
		t.Fatalf("categories = %d, want 2", len(mm.categories))
	}
	// Safe auto-selected
	if !mm.categories[0].selected {
		t.Error("safe category should be auto-selected")
	}
	// Review not auto-selected
	if mm.categories[1].selected {
		t.Error("review category should not be auto-selected")
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
		{cat: types.CategorySummary{Category: "A"}, selected: false},
	}
	m.selectedIdx = 0

	newM, _ := m.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm := newM.(model)
	if !mm.categories[0].selected {
		t.Error("space should toggle selection on")
	}

	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	mm2 := newM2.(model)
	if mm2.categories[0].selected {
		t.Error("space should toggle selection off")
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
	if !strings.Contains(out, "Categories") {
		t.Error("view should contain 'Categories'")
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
	m.dryRun = true
	out := m.View()
	if !strings.Contains(out, "Confirm Cleanup") {
		t.Error("view should contain 'Confirm Cleanup'")
	}
	if !strings.Contains(out, "DRY-RUN") {
		t.Error("view should contain 'DRY-RUN' when dryRun is true")
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

func TestViewDryRunResult(t *testing.T) {
	m := initialModel()
	m.screen = ScreenResult
	m.dryRun = true
	m.cleanResult = &types.CleanResult{Freed: 100, Files: 1}
	out := m.View()
	if !strings.Contains(out, "Dry-Run Complete") {
		t.Error("view should contain 'Dry-Run Complete'")
	}
}

func TestViewRestoreConflict(t *testing.T) {
	m := initialModel()
	m.screen = ScreenRestoreConflict
	m.conflicts = []string{`C:\a.txt`, `C:\b.txt`}
	out := m.View()
	if !strings.Contains(out, "Restore Conflicts") {
		t.Error("view should contain 'Restore Conflicts'")
	}
	if !strings.Contains(out, "2 file(s)") {
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

func TestHandleKeyErrorScreenQuit(t *testing.T) {
	m := initialModel()
	m.screen = ScreenError
	newM, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mm := newM.(model)
	if mm.screen != ScreenError {
		t.Errorf("screen should remain ScreenError, got %d", mm.screen)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
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

func TestBuildDetailList(t *testing.T) {
	items := []types.Item{
		{Path: `C:\a.txt`, Size: 100, Risk: types.RiskSafe},
		{Path: `C:\b.txt`, Size: 200, Risk: types.RiskReview},
	}
	l := buildDetailList(items, 80, 20)
	if l.Title != "Files" {
		t.Errorf("list title = %q, want 'Files'", l.Title)
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
	m.restoreIDs = []string{}
	out := m.View()
	if !strings.Contains(out, "No quarantines") {
		t.Error("view should show empty message")
	}
}

func TestViewRestoreWithIDs(t *testing.T) {
	m := initialModel()
	m.screen = ScreenRestore
	m.restoreIDs = []string{"id-1", "id-2"}
	m.restoreIdx = 1
	out := m.View()
	if !strings.Contains(out, "id-1") {
		t.Error("view should contain id-1")
	}
	if !strings.Contains(out, "id-2") {
		t.Error("view should contain id-2")
	}
}

func TestViewDoctor(t *testing.T) {
	m := initialModel()
	m.screen = ScreenDoctor
	m.doctorChecks = []doctor.Check{
		{Name: "Test", Status: doctor.StatusPass, Detail: "ok"},
	}
	out := m.View()
	if !strings.Contains(out, "Doctor") {
		t.Error("view should contain 'Doctor'")
	}
	if !strings.Contains(out, "Test") {
		t.Error("view should contain check name")
	}
}

func TestViewConfig(t *testing.T) {
	m := initialModel()
	m.screen = ScreenConfig
	m.configView = "{\"test\": true}"
	out := m.View()
	if !strings.Contains(out, "Config") {
		t.Error("view should contain 'Config'")
	}
	if !strings.Contains(out, "test") {
		t.Error("view should contain config content")
	}
}

func TestViewQuarantineCleanup(t *testing.T) {
	m := initialModel()
	m.screen = ScreenQuarantineCleanup
	out := m.View()
	if !strings.Contains(out, "Quarantine Cleanup") {
		t.Error("view should contain 'Quarantine Cleanup'")
	}
}

func TestViewLanguage(t *testing.T) {
	m := initialModel()
	m.screen = ScreenLanguage
	out := m.View()
	if !strings.Contains(out, "Select Language") {
		t.Error("view should contain 'Select Language'")
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

	// Select Russian — should stay on language screen
	newM2, _ := mm.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	mm2 := newM2.(model)
	if mm2.screen != ScreenLanguage {
		t.Errorf("after select: screen = %d, want Language", mm2.screen)
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
