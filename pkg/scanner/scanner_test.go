package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestMergeItems(t *testing.T) {
	cats := make(map[string]*types.CategorySummary)
	items := []types.Item{
		{Category: "temp", Path: `C:\a.txt`, Size: 100},
		{Category: "temp", Path: `C:\b.txt`, Size: 200},
	}
	mergeItems(cats, "Temp", types.RiskSafe, items)

	cat := cats["Temp"]
	if cat == nil {
		t.Fatal("expected Temp category")
	}
	if cat.Size != 300 {
		t.Errorf("Size = %d, want 300", cat.Size)
	}
	if cat.Files != 2 {
		t.Errorf("Files = %d, want 2", cat.Files)
	}
	if len(cat.Items) != 2 {
		t.Errorf("Items = %d, want 2", len(cat.Items))
	}

	// Empty items should not create category
	mergeItems(cats, "Empty", types.RiskSafe, nil)
	if cats["Empty"] != nil {
		t.Error("Empty items should not create category")
	}
}

func TestScanDir(t *testing.T) {
	tmp := t.TempDir()
	// Create nested structure
	_ = os.MkdirAll(filepath.Join(tmp, "sub"), 0755)
	_ = os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "sub", "b.txt"), []byte("world"), 0644)
	_ = os.WriteFile(filepath.Join(tmp, "sub", "c.log"), []byte("log"), 0644)

	cfg := config.Default()

	// Recursive, all files
	items, err := scanDir(context.Background(), tmp, "test", types.RiskSafe, nil, true, cfg)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Non-recursive
	items2, err := scanDir(context.Background(), tmp, "test", types.RiskSafe, nil, false, cfg)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if len(items2) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items2))
	}

	// Match .txt only
	items3, err := scanDir(context.Background(), tmp, "test", types.RiskSafe, []string{".txt"}, true, cfg)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if len(items3) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items3))
	}

	// Exclusions
	cfg2 := config.Default()
	cfg2.Exclusions = []string{"sub"}
	items4, err := scanDir(context.Background(), tmp, "test", types.RiskSafe, nil, true, cfg2)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if len(items4) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items4))
	}
}

func TestScanOldInstallers(t *testing.T) {
	tmp := t.TempDir()
	old := time.Now().AddDate(0, -12, 0)

	// Create old .exe
	exe := filepath.Join(tmp, "old_setup.exe")
	_ = os.WriteFile(exe, []byte("exe"), 0644)
	_ = os.Chtimes(exe, old, old)

	// Create old .msi
	msi := filepath.Join(tmp, "old_setup.msi")
	_ = os.WriteFile(msi, []byte("msi"), 0644)
	_ = os.Chtimes(msi, old, old)

	// Create recent .exe (should be skipped)
	recentExe := filepath.Join(tmp, "recent_setup.exe")
	_ = os.WriteFile(recentExe, []byte("exe"), 0644)
	_ = os.Chtimes(recentExe, time.Now(), time.Now())

	// Create non-installer old file (should be skipped)
	other := filepath.Join(tmp, "old_readme.txt")
	_ = os.WriteFile(other, []byte("txt"), 0644)
	_ = os.Chtimes(other, old, old)

	cfg := config.Default()
	cfg.OldInstallerMonths = 6

	items, err := scanOldInstallers(context.Background(), tmp, cfg)
	if err != nil {
		t.Fatalf("scanOldInstallers error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	for _, it := range items {
		if it.Category != "old_installers" {
			t.Errorf("category = %q, want old_installers", it.Category)
		}
	}
}

func TestScanLargeOldFiles(t *testing.T) {
	tmp := t.TempDir()
	old := time.Now().AddDate(0, -12, 0)

	// Create large old file (150 MB)
	big := filepath.Join(tmp, "big_old.bin")
	f, _ := os.Create(big)
	f.Truncate(150 * 1024 * 1024)
	f.Close()
	_ = os.Chtimes(big, old, old)

	// Create small old file (should be skipped)
	small := filepath.Join(tmp, "small_old.bin")
	_ = os.WriteFile(small, []byte("tiny"), 0644)
	_ = os.Chtimes(small, old, old)

	// Create large recent file (should be skipped)
	recent := filepath.Join(tmp, "big_recent.bin")
	f2, _ := os.Create(recent)
	f2.Truncate(150 * 1024 * 1024)
	f2.Close()
	_ = os.Chtimes(recent, time.Now(), time.Now())

	cfg := config.Default()
	cfg.LargeFileMonths = 6
	cfg.LargeFileMinSizeMB = 100

	items, err := scanLargeOldFiles(context.Background(), tmp, cfg)
	if err != nil {
		t.Fatalf("scanLargeOldFiles error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Path != big {
		t.Errorf("path = %q, want %q", items[0].Path, big)
	}
}

func TestScanWithConfigDisabledCategories(t *testing.T) {
	cfg := config.Default()
	cfg.EnabledCategories["Temp"] = false
	cfg.EnabledCategories["Downloads"] = false

	res, err := ScanWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("ScanWithConfig error: %v", err)
	}

	for _, cat := range res.Categories {
		if cat.Category == "Temp" || cat.Category == "Downloads" {
			t.Errorf("category %q should be disabled", cat.Category)
		}
	}
}

func TestScanEmptyFolders(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TEMP", tmp)
	t.Setenv("USERPROFILE", tmp)

	empty := filepath.Join(tmp, "empty_folder")
	_ = os.MkdirAll(empty, 0755)

	cfg := config.Default()
	items, err := scanEmptyFolders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scanEmptyFolders error: %v", err)
	}
	found := false
	for _, it := range items {
		if it.Path == empty {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find %q in empty folders", empty)
	}
}

func TestScanFirefoxCache(t *testing.T) {
	tmp := t.TempDir()
	cache2 := filepath.Join(tmp, "cache2")
	_ = os.MkdirAll(cache2, 0755)
	_ = os.WriteFile(filepath.Join(cache2, "entry1"), []byte("data"), 0644)

	cfg := config.Default()
	items, err := scanFirefoxCache(context.Background(), tmp, cfg)
	if err != nil {
		t.Fatalf("scanFirefoxCache error: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected some cache items from cache2")
	}
}

func TestScanStartupLeftovers(t *testing.T) {
	appdata := t.TempDir()
	programdata := t.TempDir()
	t.Setenv("APPDATA", appdata)
	t.Setenv("PROGRAMDATA", programdata)

	// User startup folder — 2 shortcut files + 1 non-shortcut (filtered out)
	startupDir := filepath.Join(appdata, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	_ = os.MkdirAll(startupDir, 0755)
	_ = os.WriteFile(filepath.Join(startupDir, "OldApp.lnk"), []byte("lnk"), 0644)
	_ = os.WriteFile(filepath.Join(startupDir, "OldApp.url"), []byte("url"), 0644)
	_ = os.WriteFile(filepath.Join(startupDir, "readme.txt"), []byte("txt"), 0644)

	cfg := config.Default()
	items, err := scanStartupLeftovers(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scanStartupLeftovers error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (.lnk + .url), got %d", len(items))
	}
	for _, it := range items {
		if it.Category != "startup_leftover" {
			t.Errorf("category = %q, want startup_leftover", it.Category)
		}
		if it.Risk != types.RiskReview {
			t.Errorf("risk = %v, want RiskReview", it.Risk)
		}
	}
}

func TestExtractTaskCommand(t *testing.T) {
	cases := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "UTF-8 XML",
			xml: `<?xml version="1.0"?><Task><Actions><Exec><Command>C:\Broken\app.exe</Command></Exec></Actions></Task>`,
			want: `C:\Broken\app.exe`,
		},
		{
			name: "quoted command",
			xml: `<Task><Actions><Exec><Command>"C:\My App\app.exe"</Command></Exec></Actions></Task>`,
			want: `C:\My App\app.exe`,
		},
		{
			name: "no command tag",
			xml:  `<Task><Actions></Actions></Task>`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractTaskCommand([]byte(tc.xml))
			if got != tc.want {
				t.Errorf("extractTaskCommand = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestScanDuplicateFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp)

	downloads := filepath.Join(tmp, "Downloads")
	_ = os.MkdirAll(downloads, 0755)

	// Two 2-MB files with identical content — one duplicate expected
	content := make([]byte, 2*1024*1024)
	for i := range content {
		content[i] = 0xAB
	}
	_ = os.WriteFile(filepath.Join(downloads, "file1.bin"), content, 0644)
	_ = os.WriteFile(filepath.Join(downloads, "file2.bin"), content, 0644)

	// Different content — should NOT be flagged
	other := make([]byte, 2*1024*1024)
	for i := range other {
		other[i] = 0xCD
	}
	_ = os.WriteFile(filepath.Join(downloads, "file3.bin"), other, 0644)

	// Small file (< 1 MB) — should be ignored even if duplicated
	small := []byte("tiny")
	_ = os.WriteFile(filepath.Join(downloads, "small1.txt"), small, 0644)
	_ = os.WriteFile(filepath.Join(downloads, "small2.txt"), small, 0644)

	cfg := config.Default()
	items, err := scanDuplicateFiles(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scanDuplicateFiles error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(items))
	}
	if items[0].Category != "duplicate_files" {
		t.Errorf("category = %q, want duplicate_files", items[0].Category)
	}
	if items[0].Risk != types.RiskReview {
		t.Errorf("risk = %v, want RiskReview", items[0].Risk)
	}
}

func TestScanEdgeWebViewCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	cacheDir := filepath.Join(tmp, "Microsoft", "EdgeWebView", "User", "Default", "Cache")
	_ = os.MkdirAll(cacheDir, 0755)
	_ = os.WriteFile(filepath.Join(cacheDir, "data_0"), []byte("cache"), 0644)

	cfg := config.Default()
	items, err := scanEdgeWebViewCache(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Category != "edge_webview_cache" {
		t.Errorf("category = %q, want edge_webview_cache", items[0].Category)
	}
}

func TestScanJetBrainsCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	cachePath := filepath.Join(tmp, "JetBrains", "IntelliJIdea2024.1", "caches")
	_ = os.MkdirAll(cachePath, 0755)
	_ = os.WriteFile(filepath.Join(cachePath, "vfs.db"), []byte("cache"), 0644)

	cfg := config.Default()
	items, err := scanJetBrainsCache(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Category != "jetbrains_cache" {
		t.Errorf("category = %q, want jetbrains_cache", items[0].Category)
	}
}

func TestScanOfficeCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmp)

	cachePath := filepath.Join(tmp, "Microsoft", "Office", "16.0", "OfficeFileCache")
	_ = os.MkdirAll(cachePath, 0755)
	_ = os.WriteFile(filepath.Join(cachePath, "FSD.dat"), []byte("data"), 0644)

	cfg := config.Default()
	items, err := scanOfficeCache(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Category != "office_cache" {
		t.Errorf("category = %q, want office_cache", items[0].Category)
	}
}

func TestScanRecentDocuments(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)

	recentDir := filepath.Join(tmp, "Microsoft", "Windows", "Recent")
	_ = os.MkdirAll(recentDir, 0755)
	_ = os.WriteFile(filepath.Join(recentDir, "report.lnk"), []byte("lnk"), 0644)
	_ = os.WriteFile(filepath.Join(recentDir, "AutomaticDestinations"), []byte("data"), 0644) // not .lnk, filtered

	cfg := config.Default()
	items, err := scanRecentDocuments(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 .lnk item, got %d", len(items))
	}
	if items[0].Category != "recent_documents" {
		t.Errorf("category = %q, want recent_documents", items[0].Category)
	}
	if items[0].Risk != types.RiskReview {
		t.Errorf("risk = %v, want RiskReview", items[0].Risk)
	}
}

func TestScanDirMaxFilesLimit(t *testing.T) {
	tmp := t.TempDir()
	for i := 0; i < maxScanFiles+5; i++ {
		p := filepath.Join(tmp, fmt.Sprintf("file%d.txt", i))
		_ = os.WriteFile(p, []byte("x"), 0644)
	}
	cfg := config.Default()
	items, err := scanDir(context.Background(), tmp, "test", types.RiskSafe, nil, true, cfg)
	if err != nil {
		t.Fatalf("scanDir error: %v", err)
	}
	if len(items) != maxScanFiles {
		t.Errorf("expected %d items, got %d", maxScanFiles, len(items))
	}
}
