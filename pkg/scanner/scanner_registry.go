package scanner

import (
	"context"

	"github.com/elev1e1nSure/broominal/pkg/categories"
	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// CategoryScanner scans a single cleanup category.
type CategoryScanner interface {
	Name() string
	Risk() types.RiskLevel
	Scan(ctx context.Context, cfg *config.Config) ([]types.Item, error)
}

type catScanner struct {
	name string
	risk types.RiskLevel
	scan func(context.Context, *config.Config) ([]types.Item, error)
}

func (c catScanner) Name() string          { return c.name }
func (c catScanner) Risk() types.RiskLevel { return c.risk }
func (c catScanner) Scan(ctx context.Context, cfg *config.Config) ([]types.Item, error) {
	return c.scan(ctx, cfg)
}

// scanFuncs maps each category name (from categories.All) to its scan function.
// To add a category: add it to categories.All, implement the scan function in scanner.go,
// then add the name→function entry here.
var scanFuncs = map[string]func(context.Context, *config.Config) ([]types.Item, error){
	// Quick
	"Temp":                  scanTemp,
	"Browser Cache":         scanBrowserCache,
	"Thumbnails Cache":      scanThumbnails,
	"DirectX Shader Cache":  scanDirectXShaderCache,
	"Delivery Optimization": scanDeliveryOptimization,
	"Icon Cache":            scanIconCache,
	"Windows Error Reports": scanWindowsErrorReports,
	"Opera Cache":           scanOperaCache,
	"Brave Cache":           scanBraveCache,
	"Vivaldi Cache":         scanVivaldiCache,
	"Yandex Cache":          scanYandexCache,
	"Edge Code Cache":       scanEdgeCodeCache,
	"Chrome Code Cache":     scanChromeCodeCache,
	"Firefox Cache2":        scanFirefoxCache2,
	"Windows Prefetch":      scanWindowsPrefetch,
	"AMD GPU Cache":         scanAMDGPUCache,
	"NVIDIA Shader Cache":   scanNvidiaShaderCache,
	"Intel Graphics Cache":  scanIntelGpuCache,
	"Edge WebView2 Cache":   scanEdgeWebViewCache,
	// Standard
	"Empty Folders":              scanEmptyFolders,
	"Messenger Cache":            scanMessengerCache,
	"Game Launcher Cache":        scanGameLauncherCache,
	"Dev Cache":                  scanDevCache,
	"Logs":                       scanLogs,
	"Windows Update Cache":       scanWindowsUpdateCache,
	"Nvidia Installer Leftovers": scanNvidiaInstallerLeftovers,
	"Zoom Cache":                 scanZoomCache,
	"Epic Games Cache":           scanEpicGamesCache,
	"Adobe Cache":                scanAdobeCache,
	"JetBrains Cache":            scanJetBrainsCache,
	"Office Cache":               scanOfficeCache,
	"Java Cache":                 scanJavaCache,
	"Font Cache":                 scanFontCache,
	"Windows Setup Files":        scanWindowsSetupFiles,
	"Diagnostic Data":            scanDiagnosticData,
	"Downloaded Program Files":   scanDownloadedProgramFiles,
	"Feedback Hub Logs":          scanFeedbackHubLogs,
	"BranchCache":                scanBranchCache,
	"OneDrive Logs":              scanOneDriveLogs,
	"Remote Desktop Cache":       scanRdpCache,
	"PowerShell History":         scanPowershellHistory,
	"Postman Cache":              scanPostmanCache,
	"Package Manager Cache":      scanPkgManagerCache,
	"Archive Temp Files":         scanArchiveTemp,
	// Deep
	"Recycle Bin":               scanRecycleBin,
	"Windows Defender":          scanWindowsDefender,
	"Startup Leftovers":         scanStartupLeftovers,
	"Scheduled Tasks":           scanScheduledTasksLeftovers,
	"Duplicate Files":           scanDuplicateFiles,
	"Recent Documents":          scanRecentDocuments,
	"Old Chkdsk Files":          scanOldChkdskFiles,
	"RetailDemo Content":        scanRetailDemoContent,
	"Thumbs.db":                 scanThumbsDb,
	"Windows.old":               scanWindowsOld,
	"CBS & DISM Logs":           scanCbsDismLogs,
	"Windows Installer Patches": scanWindowsInstallerPatches,
	"Android SDK Cache":         scanAndroidCache,
	"Service Cache":             scanServiceCache,
	"Dev Package Cache":         scanDevPackageCache,
	"Crash & Memory Dumps":      scanCrashMemoryDumps,
	"Microsoft Store Cache":     scanMsStoreCache,
	"Printer Spooler":           scanPrinterSpooler,
	"Old Installers":            scanOldInstallersCat,
	"Large Old Files":           scanLargeOldFilesCat,
}

// allScanners is built from categories.All + scanFuncs.
// A panic at startup means a category in categories.All is missing from scanFuncs.
var allScanners = func() []CategoryScanner {
	result := make([]CategoryScanner, 0, len(categories.All))
	for _, def := range categories.All {
		fn, ok := scanFuncs[def.Name]
		if !ok {
			panic("scanner: no scan function registered for category: " + def.Name)
		}
		result = append(result, catScanner{def.Name, def.Risk, fn})
	}
	return result
}()
