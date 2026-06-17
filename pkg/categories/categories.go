package categories

import "github.com/elev1e1nSure/broominal/pkg/types"

// Preset is the minimum preset level that enables a category.
type Preset int

const (
	Quick    Preset = 0 // enabled in all presets
	Standard Preset = 1 // enabled in Standard and Deep
	Deep     Preset = 2 // enabled only in Deep
)

// Def holds all static metadata for a scanner category.
// The scan function is wired separately in pkg/scanner.
type Def struct {
	Name        string // display name; also used as the config key
	InternalKey string // used in types.Item.Category; i18n key is "cat_"+InternalKey
	Risk        types.RiskLevel
	MinPreset   Preset // minimum preset that enables this category
}

// All is the single source of truth for every scanner category.
// To add a new category: add a Def here, add the scan function in pkg/scanner/scanner.go,
// and wire it in pkg/scanner/scanner_registry.go. Add i18n strings in pkg/i18n/i18n.go.
var All = []Def{
	// Quick (enabled in all presets)
	{Name: "Temp", InternalKey: "temp", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Browser Cache", InternalKey: "browser_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Thumbnails Cache", InternalKey: "thumbnails_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "DirectX Shader Cache", InternalKey: "directx_shader_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Delivery Optimization", InternalKey: "delivery_optimization", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Icon Cache", InternalKey: "icon_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Windows Error Reports", InternalKey: "windows_error_reports", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Opera Cache", InternalKey: "opera_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Brave Cache", InternalKey: "brave_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Vivaldi Cache", InternalKey: "vivaldi_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Yandex Cache", InternalKey: "yandex_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Edge Code Cache", InternalKey: "edge_code_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Chrome Code Cache", InternalKey: "chrome_code_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Firefox Cache2", InternalKey: "firefox_cache2", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Windows Prefetch", InternalKey: "windows_prefetch", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "AMD GPU Cache", InternalKey: "amd_gpu_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "NVIDIA Shader Cache", InternalKey: "nvidia_shader_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Intel Graphics Cache", InternalKey: "intel_gpu_cache", Risk: types.RiskSafe, MinPreset: Quick},
	{Name: "Edge WebView2 Cache", InternalKey: "edge_webview_cache", Risk: types.RiskSafe, MinPreset: Quick},
	// Standard
	{Name: "Empty Folders", InternalKey: "empty_folders", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Messenger Cache", InternalKey: "messenger_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Game Launcher Cache", InternalKey: "game_launcher_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Dev Cache", InternalKey: "dev_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Logs", InternalKey: "logs", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Windows Update Cache", InternalKey: "windows_update_cache", Risk: types.RiskReview, MinPreset: Standard},
	{Name: "Nvidia Installer Leftovers", InternalKey: "nvidia_installer_leftovers", Risk: types.RiskReview, MinPreset: Standard},
	{Name: "Zoom Cache", InternalKey: "zoom_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Epic Games Cache", InternalKey: "epic_games_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Adobe Cache", InternalKey: "adobe_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "JetBrains Cache", InternalKey: "jetbrains_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Office Cache", InternalKey: "office_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Java Cache", InternalKey: "java_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Font Cache", InternalKey: "font_cache", Risk: types.RiskReview, MinPreset: Standard},
	{Name: "Windows Setup Files", InternalKey: "windows_setup_files", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Diagnostic Data", InternalKey: "diagnostic_data", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Downloaded Program Files", InternalKey: "downloaded_program_files", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Feedback Hub Logs", InternalKey: "feedback_hub_logs", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "BranchCache", InternalKey: "branch_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "OneDrive Logs", InternalKey: "onedrive_logs", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Remote Desktop Cache", InternalKey: "rdp_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Postman Cache", InternalKey: "postman_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Package Manager Cache", InternalKey: "pkg_manager_cache", Risk: types.RiskSafe, MinPreset: Standard},
	{Name: "Archive Temp Files", InternalKey: "archive_temp", Risk: types.RiskSafe, MinPreset: Standard},
	// Deep
	{Name: "Recycle Bin", InternalKey: "recycle_bin", Risk: types.RiskDanger, MinPreset: Deep},
	{Name: "Windows Defender", InternalKey: "windows_defender", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Startup Leftovers", InternalKey: "startup_leftover", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Scheduled Tasks", InternalKey: "scheduled_tasks_leftover", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Duplicate Files", InternalKey: "duplicate_files", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Old Chkdsk Files", InternalKey: "old_chkdsk_files", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "RetailDemo Content", InternalKey: "retail_demo_content", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "Thumbs.db", InternalKey: "thumbs_db", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "Windows.old", InternalKey: "windows_old", Risk: types.RiskDanger, MinPreset: Deep},
	{Name: "CBS & DISM Logs", InternalKey: "cbs_dism_logs", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Windows Installer Patches", InternalKey: "windows_installer_patches", Risk: types.RiskDanger, MinPreset: Deep},
	{Name: "Android SDK Cache", InternalKey: "android_cache", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "Service Cache", InternalKey: "service_cache", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "Dev Package Cache", InternalKey: "dev_package_cache", Risk: types.RiskDanger, MinPreset: Deep},
	{Name: "Crash & Memory Dumps", InternalKey: "crash_memory_dumps", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Microsoft Store Cache", InternalKey: "ms_store_cache", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Printer Spooler", InternalKey: "printer_spooler", Risk: types.RiskReview, MinPreset: Deep},
	{Name: "Old Installers", InternalKey: "old_installers", Risk: types.RiskSafe, MinPreset: Deep},
	{Name: "Large Old Files", InternalKey: "large_old_files", Risk: types.RiskDanger, MinPreset: Deep},
}
