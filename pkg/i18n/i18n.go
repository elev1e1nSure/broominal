package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var translations = map[string]map[string]string{
	"en": {
		"main_menu":               "Main Menu",
		"menu_scan_clean":         "Scan & Clean",
		"menu_restore":            "Restore Files",
		"menu_doctor":             "System Check",
		"menu_config":             "Settings",
		"menu_cleanup":            "Empty Bin",
		"menu_settings":           "Settings",
		"hint_select":             "Enter: select  Q: quit",
		"dashboard":               "Broominal — Dashboard",
		"total_found":             "Total found",
		"safe":                    "Safe",
		"review":                  "Review",
		"danger":                  "Danger",
		"hint_continue":           "Press Enter or Space to continue",
		"categories":              "Categories",
		"category":                "Category",
		"size":                    "Size",
		"files":                   "Files",
		"risk":                    "Risk",
		"risk_safe":               "Safe",
		"risk_review":             "Review",
		"risk_danger":             "Danger",
		"select":                  "Select",
		"hint_categories":         "Space: toggle  Enter: confirm  D: details  Q: quit",
		"details":                 "Details",
		"hint_back":               "Q/Esc: back",
		"confirm_cleanup":         "Confirm Cleanup",
		"dry_run":                 "[PREVIEW]",
		"hint_confirm":            "Enter: proceed  T: toggle preview  Esc: back",
		"will_free":               "Will free",
		"cleaning":                "Cleaning...",
		"cleanup_complete":        "Cleanup Complete",
		"freed":                   "Freed",
		"restore_id":              "Restore",
		"hint_result":             "R: restore last  Q: quit",
		"dry_run_complete":        "Preview Complete",
		"would_free":              "Would free",
		"restored":                "Restored",
		"hint_restored":           "Files restored successfully.",
		"restore_conflicts":       "Restore Conflicts",
		"files_already_exist":     "file(s) already exist at original paths",
		"hint_conflicts":          "O: overwrite all  S: skip all  C/Esc: cancel",
		"warning":                 "Warning",
		"recycle_bin_warn":        "Recycle Bin contains %d files.",
		"hint_recycle_warn":       "Opening details may be very slow.",
		"hint_recycle_continue":   "Enter: continue anyway  Esc: back",
		"error":                   "Error",
		"hint_error_quit":         "Press Q or Esc to quit",
		"scanning":                "Scanning...",
		"moving_files":            "Moving files to backup...",
		"please_wait":             "Please wait, this may take a while.",
		"quit":                    "quit",
		"restore":                 "Restore Files",
		"no_quarantines":          "No backups available.",
		"hint_restore":            "Enter: restore  Esc: back",
		"restored_n_skipped":      "Restored %d files (%d skipped)",
		"doctor":                  "System Check",
		"config":                  "Settings",
		"config_categories":       "Enabled Categories",
		"config_thresholds":       "Thresholds",
		"quarantine_cleanup":      "Quarantine Cleanup",
		"cleanup_desc":            "Delete old backups or clean everything.",
		"all_quarantines":         "Everything",
		"old_only":                "Old only",
		"mode":                    "Mode",
		"toggle_mode":             "toggle mode",
		"hint_cleanup":            "T: toggle preview  Enter: proceed  Esc: back",
		"settings":                "Settings",
		"language":                "Language",
		"select_language":         "Select Language",
		"hint_language":           "Enter: select  Esc: back",
		"english":                 "English",
		"russian":                 "Russian",
		"save":                    "Save",
		"change":                  "change",
		"old_installer_months":    "Old installers: minimum age",
		"large_file_min_size_mb":  "Large files: minimum size",
		"large_file_months":       "Large files: minimum age",
		"old_temp_days":           "Old temp files: minimum age",
		"old_extension_days":      "Old .tmp/.log/.bak: minimum age",
		"quarantine_max_age_days": "Auto-delete backups older than",
		"days":                    "days",
		"months":                  "months",
		"continue":                "continue",
		"toggle":                  "toggle",
		"confirm":                 "confirm",
		"proceed":                 "proceed",
		"back":                    "back",
		"overwrite_all":           "overwrite all",
		"skip_all":                "skip all",
		"cancel":                  "cancel",
		"continue_anyway":         "continue anyway",
		"toggle_dry_run":          "toggle preview",
		"restore_last":            "restore last",
		"check_admin":             "Admin privileges",
		"check_quarantine_dir":    "Quarantine directory",
		"check_reports_dir":       "Reports directory",
		"check_config_dir":        "Config directory",
		"check_temp_dir":          "Temp directory",
		"check_userprofile_dir":   "User profile directory",
		"check_manifests":         "Quarantine manifests",
		"check_stats":             "Quarantine stats",
		"running_as_admin":        "Running as administrator",
		"not_running_as_admin":    "Not running as administrator (some paths may be inaccessible)",
		"dir_not_writable":        "Cannot write to %s: %v",
		"no_backups_yet":          "No backups yet",
		"invalid_manifests":       "%d invalid manifest(s)",
		"valid_backups":           "%d valid backup(s)",
		"backups_files_size":      "%d backups, %d files, %s",
		"cat_temp":                "Temp",
		"cat_downloads":           "Downloads",
		"cat_browser_cache":       "Browser Cache",
		"cat_recycle_bin":         "Recycle Bin",
		"cat_logs":                "Logs",
		"cat_old_installers":      "Old Installers",
		"cat_large_old_files":     "Large Old Files",
		"cat_thumbnails_cache":    "Thumbnails Cache",
		"cat_directx_shader_cache":"DirectX Shader Cache",
		"cat_delivery_optimization":"Delivery Optimization",
		"cat_windows_error_reports":"Windows Error Reports",
		"cat_discord_cache":       "Discord Cache",
		"cat_steam_cache":         "Steam Cache",
		"cat_windows_update_cache":"Windows Update Cache",
		"cat_crash_memory_dumps":  "Crash & Memory Dumps",
		"cat_nvidia_installer_leftovers": "Nvidia Installer Leftovers",
		"cat_telegram_desktop_cache":"Telegram Desktop Cache",
		"cat_vscode_cache":        "VSCode Cache",
		"cat_edge_code_cache":     "Edge Code Cache",
		"cat_chrome_code_cache":   "Chrome Code Cache",
		"cat_firefox_cache2":      "Firefox Cache2",
		"cat_old_temp_files":      "Old Temp Files",
		"cat_old_tmp_files":       "Old .tmp Files",
		"cat_old_log_files":       "Old .log Files",
		"cat_old_bak_files":       "Old .bak Files",
		"cat_empty_folders":       "Empty Folders",
		"cat_npm_cache":           "npm Cache",
		"cat_pip_cache":           "pip Cache",
	},
	"ru": {
		"main_menu":               "Меню",
		"menu_scan_clean":         "Сканирование и очистка",
		"menu_restore":            "Восстановить файлы",
		"menu_doctor":             "Проверка системы",
		"menu_config":             "Настройки",
		"menu_cleanup":            "Очистить корзину",
		"menu_settings":           "Настройки",
		"hint_select":             "Enter: выбрать  Q: выйти",
		"dashboard":               "Broominal — Обзор",
		"total_found":             "Всего найдено",
		"safe":                    "Безопасно",
		"review":                  "Проверить",
		"danger":                  "Опасно",
		"hint_continue":           "Нажмите Enter или Space для продолжения",
		"categories":              "Категории",
		"category":                "Категория",
		"size":                    "Размер",
		"files":                   "Файлы",
		"risk":                    "Риск",
		"risk_safe":               "Безопасно",
		"risk_review":             "Проверить",
		"risk_danger":             "Опасно",
		"select":                  "Выбор",
		"hint_categories":         "Space: выбрать  Enter: подтвердить  D: детали  Q: выйти",
		"details":                 "Детали",
		"hint_back":               "Q/Esc: назад",
		"confirm_cleanup":         "Подтвердить очистку",
		"dry_run":                 "[ПРЕДПРОСМОТР]",
		"hint_confirm":            "Enter: продолжить  T: предпросмотр  Esc: назад",
		"will_free":               "Будет освобождено",
		"cleaning":                "Очистка...",
		"cleanup_complete":        "Очистка завершена",
		"freed":                   "Освобождено",
		"restore_id":              "Восстановление",
		"hint_result":             "R: восстановить последнее  Q: выйти",
		"dry_run_complete":        "Предпросмотр завершён",
		"would_free":              "Будет освобождено",
		"restored":                "Восстановлено",
		"hint_restored":           "Файлы успешно восстановлены.",
		"restore_conflicts":       "Конфликты восстановления",
		"files_already_exist":     "файл(ов) уже существуют по оригинальным путям",
		"hint_conflicts":          "O: перезаписать  S: пропустить  C/Esc: отмена",
		"warning":                 "Предупреждение",
		"recycle_bin_warn":        "Корзина содержит %d файлов.",
		"hint_recycle_warn":       "Открытие деталей может быть очень медленным.",
		"hint_recycle_continue":   "Enter: продолжить  Esc: назад",
		"error":                   "Ошибка",
		"hint_error_quit":         "Нажмите Q или Esc для выхода",
		"scanning":                "Сканирование...",
		"moving_files":            "Перемещение файлов в резервную копию...",
		"please_wait":             "Пожалуйста, подождите, это может занять некоторое время.",
		"quit":                    "выйти",
		"restore":                 "Восстановить файлы",
		"no_quarantines":          "Резервные копии отсутствуют.",
		"hint_restore":            "Enter: восстановить  Esc: назад",
		"restored_n_skipped":      "Восстановлено %d файлов (%d пропущено)",
		"doctor":                  "Проверка системы",
		"config":                  "Настройки",
		"config_categories":       "Категории",
		"config_thresholds":       "Пороги",
		"quarantine_cleanup":      "Очистка карантина",
		"cleanup_desc":            "Удалить старые резервные копии или очистить всё.",
		"all_quarantines":         "Всё",
		"old_only":                "Только старые",
		"mode":                    "Режим",
		"toggle_mode":             "сменить режим",
		"hint_cleanup":            "T: предпросмотр  Enter: запустить  Esc: назад",
		"settings":                "Настройки",
		"language":                "Язык",
		"select_language":         "Выбор языка",
		"hint_language":           "Enter: выбрать  Esc: назад",
		"english":                 "English",
		"russian":                 "Русский",
		"save":                    "Сохранить",
		"change":                  "изменить",
		"old_installer_months":    "Старые установщики: мин. возраст",
		"large_file_min_size_mb":  "Большие файлы: мин. размер",
		"large_file_months":       "Большие файлы: мин. возраст",
		"old_temp_days":           "Старые temp-файлы: мин. возраст",
		"old_extension_days":      "Старые .tmp/.log/.bak: мин. возраст",
		"quarantine_max_age_days": "Автоудаление резервных копий старше",
		"days":                    "дней",
		"months":                  "мес",
		"continue":                "продолжить",
		"toggle":                  "выбрать",
		"confirm":                 "подтвердить",
		"proceed":                 "продолжить",
		"back":                    "назад",
		"overwrite_all":           "перезаписать все",
		"skip_all":                "пропустить все",
		"cancel":                  "отмена",
		"continue_anyway":         "всё равно продолжить",
		"toggle_dry_run":          "предпросмотр",
		"restore_last":            "восстановить последнее",
		"check_admin":             "Права администратора",
		"check_quarantine_dir":    "Папка резервных копий",
		"check_reports_dir":       "Папка отчётов",
		"check_config_dir":        "Папка настроек",
		"check_temp_dir":          "Папка Temp",
		"check_userprofile_dir":   "Папка профиля",
		"check_manifests":         "Манифесты резервных копий",
		"check_stats":             "Статистика резервных копий",
		"running_as_admin":        "Запущено от имени администратора",
		"not_running_as_admin":    "Нет прав администратора (некоторые пути могут быть недоступны)",
		"dir_not_writable":        "Невозможно записать в %s: %v",
		"no_backups_yet":          "Резервных копий пока нет",
		"invalid_manifests":       "%d повреждённых манифестов",
		"valid_backups":           "%d действующих резервных копий",
		"backups_files_size":      "%d резервных копий, %d файлов, %s",
		"cat_temp":                "Временные файлы",
		"cat_downloads":           "Загрузки",
		"cat_browser_cache":       "Кэш браузера",
		"cat_recycle_bin":         "Корзина",
		"cat_logs":                "Логи",
		"cat_old_installers":      "Старые установщики",
		"cat_large_old_files":     "Большие старые файлы",
		"cat_thumbnails_cache":    "Кэш миниатюр",
		"cat_directx_shader_cache":"Кэш шейдеров DirectX",
		"cat_delivery_optimization":"Delivery Optimization",
		"cat_windows_error_reports":"Отчёты об ошибках Windows",
		"cat_discord_cache":       "Кэш Discord",
		"cat_steam_cache":         "Кэш Steam",
		"cat_windows_update_cache":"Кэш обновлений Windows",
		"cat_crash_memory_dumps":  "Дампы памяти и сбоев",
		"cat_nvidia_installer_leftovers": "Остатки установщика Nvidia",
		"cat_telegram_desktop_cache":"Кэш Telegram Desktop",
		"cat_vscode_cache":        "Кэш VSCode",
		"cat_edge_code_cache":     "Кэш кода Edge",
		"cat_chrome_code_cache":   "Кэш кода Chrome",
		"cat_firefox_cache2":      "Кэш Firefox 2",
		"cat_old_temp_files":      "Старые temp-файлы",
		"cat_old_tmp_files":       "Старые .tmp файлы",
		"cat_old_log_files":       "Старые .log файлы",
		"cat_old_bak_files":       "Старые .bak файлы",
		"cat_empty_folders":       "Пустые папки",
		"cat_npm_cache":           "Кэш npm",
		"cat_pip_cache":           "Кэш pip",
	},
}

var (
	currentLang   = "en"
	currentLangMu sync.RWMutex
)

func SetLanguage(lang string) {
	currentLangMu.Lock()
	defer currentLangMu.Unlock()
	if _, ok := translations[lang]; ok {
		currentLang = lang
	} else {
		currentLang = "en"
	}
}

func CurrentLanguage() string {
	currentLangMu.RLock()
	defer currentLangMu.RUnlock()
	return currentLang
}

func T(key string) string {
	currentLangMu.RLock()
	defer currentLangMu.RUnlock()
	if tr, ok := translations[currentLang][key]; ok {
		return tr
	}
	if tr, ok := translations["en"][key]; ok {
		return tr
	}
	return key
}

func SupportedLanguages() []string {
	return []string{"en", "ru"}
}

// categoryKeyMap maps internal category names to translation keys.
var categoryKeyMap = map[string]string{
	"Temp":                     "cat_temp",
	"Downloads":                "cat_downloads",
	"Browser Cache":            "cat_browser_cache",
	"Recycle Bin":              "cat_recycle_bin",
	"Logs":                     "cat_logs",
	"Old Installers":           "cat_old_installers",
	"Large Old Files":          "cat_large_old_files",
	"Thumbnails Cache":         "cat_thumbnails_cache",
	"DirectX Shader Cache":     "cat_directx_shader_cache",
	"Delivery Optimization":    "cat_delivery_optimization",
	"Windows Error Reports":    "cat_windows_error_reports",
	"Discord Cache":            "cat_discord_cache",
	"Steam Cache":              "cat_steam_cache",
	"Windows Update Cache":     "cat_windows_update_cache",
	"Crash & Memory Dumps":     "cat_crash_memory_dumps",
	"Nvidia Installer Leftovers": "cat_nvidia_installer_leftovers",
	"Telegram Desktop Cache":   "cat_telegram_desktop_cache",
	"VSCode Cache":             "cat_vscode_cache",
	"Edge Code Cache":          "cat_edge_code_cache",
	"Chrome Code Cache":        "cat_chrome_code_cache",
	"Firefox Cache2":           "cat_firefox_cache2",
	"Old Temp Files":           "cat_old_temp_files",
	"Old .tmp Files":           "cat_old_tmp_files",
	"Old .log Files":           "cat_old_log_files",
	"Old .bak Files":           "cat_old_bak_files",
	"Empty Folders":            "cat_empty_folders",
	"npm Cache":                "cat_npm_cache",
	"pip Cache":                "cat_pip_cache",
}

// CategoryName returns the translated display name for a scanner category.
func CategoryName(name string) string {
	if key, ok := categoryKeyMap[name]; ok {
		return T(key)
	}
	return name
}

// ipAPIResponse is the minimal struct we need from ipapi.co.
type ipAPIResponse struct {
	CountryCode string `json:"country_code"`
}

// DetectFromIP calls a public geo-IP service and returns a language code.
func DetectFromIP() (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://ipapi.co/json/")
	if err != nil {
		return "", fmt.Errorf("geo-ip request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("geo-ip status: %d", resp.StatusCode)
	}
	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("geo-ip decode: %w", err)
	}
	switch data.CountryCode {
	case "RU", "BY", "KZ", "UA":
		return "ru", nil
	default:
		return "en", nil
	}
}
