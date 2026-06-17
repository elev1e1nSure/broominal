# Cleanup Categories

62 categories across three presets. Risk levels: **Safe** (auto-selected), **Review** (manual select), **Danger** (never auto-selected).

---

## Quick — 19 categories

Enabled in all presets. Temporary data that rebuilds automatically.

| Category | Paths |
|----------|-------|
| **Temp** | `%TEMP%` |
| **Browser Cache** | Chrome, Edge, Firefox cache dirs |
| **Thumbnails Cache** | `thumbcache_*.db` |
| **DirectX Shader Cache** | `%LOCALAPPDATA%\D3DSCache` |
| **Delivery Optimization** | `SoftwareDistribution\DeliveryOptimization` |
| **Icon Cache** | `%LOCALAPPDATA%\IconCache.db` |
| **Windows Error Reports** | `%ProgramData%\Microsoft\Windows\WER`, `%LOCALAPPDATA%\Microsoft\Windows\WER` |
| **Opera Cache** | `%LOCALAPPDATA%\Opera Software\Opera Stable\Cache` |
| **Brave Cache** | `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data\Default\Cache` |
| **Vivaldi Cache** | `%LOCALAPPDATA%\Vivaldi\User Data\Default\Cache` |
| **Yandex Cache** | `%LOCALAPPDATA%\Yandex\YandexBrowser\User Data\Default\Cache` |
| **Edge Code Cache** | `%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\Code Cache` |
| **Chrome Code Cache** | `%LOCALAPPDATA%\Google\Chrome\User Data\Default\Code Cache` |
| **Firefox Cache2** | `%LOCALAPPDATA%\Mozilla\Firefox\Profiles\*\cache2` |
| **Windows Prefetch** | `%SystemRoot%\Prefetch` |
| **AMD GPU Cache** | `%LOCALAPPDATA%\AMD\DxCache`, `CLCache` |
| **NVIDIA Shader Cache** | `%LOCALAPPDATA%\NVIDIA\DXCache`, `GLCache`, `NV_Cache` |
| **Intel Graphics Cache** | `%LOCALAPPDATA%\Intel\ShaderCache`, `%ProgramData%\Intel\ShaderCache` |
| **Edge WebView2 Cache** | `%LOCALAPPDATA%\Microsoft\EdgeWebView\User\Default\Cache`, `Code Cache`, `GPUCache` |

---

## Standard — 24 additional categories

App caches and logs. Enabled in Standard and Deep presets.

| Category | Paths |
|----------|-------|
| **Empty Folders** | Empty dirs in `%TEMP%` |
| **Messenger Cache** | Discord (Cache, Code Cache), Telegram, Slack (Cache, Code Cache, GPUCache), Teams |
| **Game Launcher Cache** | Steam, Epic, Battle.net, Rockstar, EA, Ubisoft, GOG |
| **Dev Cache** (⚠ Review) | VSCode, npm, pip, Git, Visual Studio, JetBrains, Go build, Rust cargo |
| **Logs** | `*.log` files in `%TEMP%` |
| **Windows Update Cache** (⚠ Review) | `%SystemRoot%\SoftwareDistribution\Download` |
| **Nvidia Installer Leftovers** (⚠ Review) | `C:\NVIDIA\DisplayDriver`, `%ProgramData%\NVIDIA Corporation\Downloader` |
| **Zoom Cache** | `%APPDATA%\Zoom\data`, `logs` |
| **Epic Games Cache** | `%LOCALAPPDATA%\EpicGamesLauncher\Saved\webcache`, `Logs`, `crashes` |
| **Adobe Cache** | `%APPDATA%\Adobe\Common\Media Cache` |
| **JetBrains Cache** | `%LOCALAPPDATA%\JetBrains\*\caches` |
| **Office Cache** | `%LOCALAPPDATA%\Microsoft\Office\*\OfficeFileCache` |
| **Java Cache** | `%APPDATA%\LocalLow\Sun\Java\Deployment\cache` |
| **Font Cache** (⚠ Review) | `%SystemRoot%\ServiceProfiles\LocalService\AppData\Local\FontCache` |
| **Windows Setup Files** | `%SystemRoot%\Panther` |
| **Diagnostic Data** | `%ProgramData%\Microsoft\Diagnosis\ETLLogs` |
| **Downloaded Program Files** | `%SystemRoot%\Downloaded Program Files` |
| **Feedback Hub Logs** | `%LOCALAPPDATA%\Packages\Microsoft.WindowsFeedbackHub_*\LocalState\DiagOutputDir` |
| **BranchCache** | `%SystemRoot%\ServiceProfiles\NetworkService\AppData\Local\PeerDistRepub` |
| **OneDrive Logs** | `%LOCALAPPDATA%\Microsoft\OneDrive\logs`, `%USERPROFILE%\OneDrive\Logs` |
| **Remote Desktop Cache** | `%LOCALAPPDATA%\Microsoft\Terminal Server Client\Cache` |
| **PowerShell History** | `%APPDATA%\Microsoft\Windows\PowerShell\PSReadLine\*` |
| **Postman Cache** | `%APPDATA%\Postman\Cache`, `Code Cache`, `GPUCache` |
| **Package Manager Cache** | Chocolatey `<ChocolateyInstall>\cache`, Scoop `%USERPROFILE%\scoop\cache`, winget source indexes, pip wheels, Conda `pkgs\` |

---

## Deep — 20 additional categories

Areas that may contain user data or affect system behavior. Review carefully.

| Category | Paths |
|----------|-------|
| **Recycle Bin** | `$Recycle.Bin\*` |
| **Windows Defender** (⚠ Review) | `%ProgramData%\Microsoft\Windows Defender\Scans\History` |
| **Startup Leftovers** (⚠ Review) | `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup` |
| **Scheduled Tasks** (⚠ Review) | `%SystemRoot%\System32\Tasks` (non-Microsoft, orphaned) |
| **Duplicate Files** (⚠ Review) | Downloads, Desktop, Documents (by MD5 content) |
| **Recent Documents** (⚠ Review) | `%APPDATA%\Microsoft\Windows\Recent\*.lnk` |
| **Old Chkdsk Files** (⚠ Review) | `FOUND.*` dirs on all drives |
| **RetailDemo Content** | `%SystemRoot%\ServiceProfiles\LocalService\AppData\Local\Microsoft\Windows\RetailDemo` |
| **Thumbs.db** | `Thumbs.db`, `ehthumbs.db` in user profile |
| **Windows.old** (⚠ Review) | `C:\Windows.old`, `C:\$WinREAgent` |
| **CBS & DISM Logs** (⚠ Review) | `%SystemRoot%\Logs\CBS`, `DISM` |
| **Windows Installer Patches** (⚠ Review) | `%SystemRoot%\Installer\$PatchCache$` |
| **Android SDK Cache** | `%LOCALAPPDATA%\Android\Sdk\.temp`, `.gradle\caches\build-cache-1` |
| **Service Cache** (⚠ Review) | Spotify, OneDrive, Office, Adobe, OBS, TeamViewer caches |
| **Dev Package Cache** (⚠ Review) | Docker, NuGet, Unity — breaks builds if deleted |
| **Crash & Memory Dumps** (⚠ Review) | CrashDumps, Minidump, MEMORY.DMP (older than 7 days) |
| **Microsoft Store Cache** (⚠ Review) | UWP/Store app caches — all apps including games |
| **Printer Spooler** (⚠ Review) | Stuck print jobs older than 1 hour |
| **Old Installers** (⚠ Review) | Installer artifacts in `%TEMP%`, `%APPDATA%` — `.exe`, `.msi` older than `old_installer_months` |
| **Large Old Files** (⚠ Review) | Files larger than `large_file_size_mb` unused since `large_file_months` months |

---

## Adding a Category

1. Add a `Def` entry to `pkg/categories/categories.go`  
2. Write a scan function in `pkg/scanner/scanner_xxx.go` (in the appropriate scanner file, e.g. scanner_windows.go, scanner_browsers.go, etc.)  
3. Wire it in `pkg/scanner/scanner_registry.go`  
4. Add `cat_*` and `cat_desc_*` translation strings in `pkg/i18n/strings_en.go` and `pkg/i18n/strings_ru.go`
