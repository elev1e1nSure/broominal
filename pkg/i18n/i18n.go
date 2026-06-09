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
		"menu_restore":            "Backup Copies",
		"menu_doctor":             "Diagnostics",
		"menu_config":             "Settings",
		"menu_cleanup":            "Backup Cleanup",
		"menu_settings":           "Settings",
		"hint_select":             "Enter: select  Q: quit",
		"dashboard":               "Broominal — Dashboard",
		"total_found":             "Total found",
		"safe":                    "Safe",
		"review":                  "Review",
		"danger":                  "Danger",
		"hint_continue":           "Press Enter to continue",
		"categories":              "Categories",
		"category":                "Category",
		"size":                    "Size",
		"files":                   "Files",
		"risk":                    "Risk",
		"risk_safe":               "Safe",
		"risk_review":             "Review",
		"risk_danger":             "Danger",
		"select":                  "Select",
		"hint_categories":         "Space: select  Enter: confirm  D: details  Esc: back",
		"details":                 "Details",
		"hint_back":               "Esc: back",
		"confirm_cleanup":         "Confirm Cleanup",
		"hint_confirm":            "Enter: proceed  Esc: back",
		"will_free":               "Will free",
		"cleaning":                "Cleaning...",
		"cleanup_complete":        "Cleanup Complete",
		"freed":                   "Freed",
		"restore_id":              "Restore ID",
		"hint_result":             "R: restore last  Q: quit",
		"restored":                "Restored",
		"hint_restored":           "Files restored successfully.",
		"restore_conflicts":       "File Conflicts",
		"files_already_exist":     "already exist at destination",
		"hint_conflicts":          "O: overwrite  S: skip  Esc: cancel",
		"warning":                 "Warning",
		"recycle_bin_warn":        "Recycle Bin contains %d files.",
		"hint_recycle_warn":       "Opening details may be very slow.",
		"hint_recycle_continue":   "Enter: continue anyway  Esc: back",
		"error":                   "Error",
		"hint_error_quit":         "Press Esc to return to menu",
		"scanning":                "Scanning...",
		"moving_files":            "Moving files to backup...",
		"please_wait":             "Please wait, this may take a while.",
		"quit":                    "quit",
		"restore":                 "Restore Files",
		"no_quarantines":          "No backups available.",
		"hint_restore":            "Enter: restore  D: delete  A: delete all  Esc: back",
		"delete":                  "delete",
		"delete_all":              "delete all",
		"restored_n_skipped":      "Restored %d files (%d skipped)",
		"doctor":                  "Diagnostics",
		"config":                  "Settings",
		"config_categories":       "Cleanup Categories",
		"config_thresholds":       "Age & Size Limits",
		"quarantine_cleanup":      "Backup Cleanup",
		"cleanup_desc":            "Delete all backup copies.",
		"all_quarantines":         "All",
		"old_only":                "Old only",
		"mode":                    "Mode",
		"toggle_mode":             "toggle mode",
		"hint_cleanup":            "A: mode  Enter: proceed  Esc: back",
		"settings":                "Settings",
		"language":                "Language",
		"select_language":         "Language",
		"hint_language":           "Enter: select",
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
		"overwrite_all":           "overwrite",
		"skip_all":                "skip",
		"cancel":                  "cancel",
		"continue_anyway":         "continue anyway",
		"restore_last":            "restore last",
		"check_admin":             "Administrator rights",
		"check_quarantine_dir":    "Backup folder",
		"check_reports_dir":       "Reports folder",
		"check_config_dir":        "Settings folder",
		"check_temp_dir":          "Temp folder",
		"check_userprofile_dir":   "User profile folder",
		"check_manifests":         "Backup integrity",
		"check_stats":             "Backup storage",
		"running_as_admin":        "OK",
		"not_running_as_admin":    "Limited — some paths may be skipped",
		"dir_not_writable":        "Cannot write to %s: %v",
		"no_backups_yet":          "No backups yet",
		"invalid_manifests":       "%d backup(s) damaged",
		"valid_backups":           "%d backup(s) OK",
		"backups_files_size":      "%d backups, %d files, %s",
		"fix_issue":               "Fix issue",
		"suggest_admin":           "Restart the program as administrator",
		"suggest_check_permissions": "Check folder permissions or run as administrator",
		"suggest_env_missing":     "Environment variable is missing; restart your session",
		"suggest_remove_damaged":  "Go to Backup Cleanup and remove damaged backups",
		"admin_required":          "Administrator rights required",
		"admin_required_desc":       "Some cleanup paths require elevated privileges. Restart as administrator to continue.",
		"restart_as_admin":          "Restart as administrator",
		"exit":                      "Exit",
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
		"menu_scan_clean":         "Поиск и очистка",
		"menu_restore":            "Резервные копии",
		"menu_doctor":             "Диагностика",
		"menu_config":             "Настройки",
		"menu_cleanup":            "Очистка резервных копий",
		"menu_settings":           "Настройки",
		"hint_select":             "Enter: выбрать  Q: выйти",
		"dashboard":               "Broominal — Обзор",
		"total_found":             "Всего найдено",
		"safe":                    "Безопасно",
		"review":                  "Проверить",
		"danger":                  "Опасно",
		"hint_continue":           "Нажмите Enter для продолжения",
		"categories":              "Категории",
		"category":                "Категория",
		"size":                    "Размер",
		"files":                   "Файлы",
		"risk":                    "Риск",
		"risk_safe":               "Безопасно",
		"risk_review":             "Проверить",
		"risk_danger":             "Опасно",
		"select":                  "Выбор",
		"hint_categories":         "Space: выбрать  Enter: подтвердить  D: детали  Esc: назад",
		"details":                 "Детали",
		"hint_back":               "Esc: назад",
		"confirm_cleanup":         "Подтвердить очистку",
		"hint_confirm":            "Enter: очистить  Esc: назад",
		"will_free":               "Будет освобождено",
		"cleaning":                "Очистка...",
		"cleanup_complete":        "Очистка завершена",
		"freed":                   "Освобождено",
		"restore_id":              "ID восстановления",
		"hint_result":             "R: восстановить последнее  Q: выйти",
		"restored":                "Восстановлено",
		"hint_restored":           "Файлы успешно восстановлены.",
		"restore_conflicts":       "Конфликты файлов",
		"files_already_exist":     "уже существуют по назначению",
		"hint_conflicts":          "O: перезаписать  S: пропустить  Esc: отмена",
		"warning":                 "Предупреждение",
		"recycle_bin_warn":        "Корзина содержит %d файлов.",
		"hint_recycle_warn":       "Открытие деталей может быть очень медленным.",
		"hint_recycle_continue":   "Enter: продолжить  Esc: назад",
		"error":                   "Ошибка",
		"hint_error_quit":         "Нажмите Esc для возврата в меню",
		"scanning":                "Сканирование...",
		"moving_files":            "Перемещение файлов в резервную копию...",
		"please_wait":             "Пожалуйста, подождите, это может занять некоторое время.",
		"quit":                    "выйти",
		"restore":                 "Восстановить файлы",
		"no_quarantines":          "Резервные копии отсутствуют.",
		"hint_restore":            "Enter: восстановить  D: удалить  A: удалить всё  Esc: назад",
		"delete":                  "удалить",
		"delete_all":              "удалить всё",
		"restored_n_skipped":      "Восстановлено %d файлов (%d пропущено)",
		"doctor":                  "Диагностика",
		"config":                  "Настройки",
		"config_categories":       "Категории очистки",
		"config_thresholds":       "Пороги возраста и размера",
		"quarantine_cleanup":      "Очистка резервных копий",
		"cleanup_desc":            "Удалить все резервные копии.",
		"all_quarantines":         "Все",
		"old_only":                "Только старые",
		"mode":                    "Режим",
		"toggle_mode":             "сменить режим",
		"hint_cleanup":            "A: режим  Enter: запустить  Esc: назад",
		"settings":                "Настройки",
		"language":                "Язык",
		"select_language":         "Язык",
		"hint_language":           "Enter: выбрать",
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
		"overwrite_all":           "перезаписать",
		"skip_all":                "пропустить",
		"cancel":                  "отмена",
		"continue_anyway":         "всё равно продолжить",
		"restore_last":            "восстановить последнее",
		"check_admin":             "Права администратора",
		"check_quarantine_dir":    "Папка резервных копий",
		"check_reports_dir":       "Папка отчётов",
		"check_config_dir":        "Папка настроек",
		"check_temp_dir":          "Папка Temp",
		"check_userprofile_dir":   "Папка профиля",
		"check_manifests":         "Манифесты резервных копий",
		"check_stats":             "Статистика резервных копий",
		"running_as_admin":        "OK",
		"not_running_as_admin":    "Ограниченно — некоторые пути могут быть пропущены",
		"dir_not_writable":        "Невозможно записать в %s: %v",
		"no_backups_yet":          "Резервных копий пока нет",
		"invalid_manifests":       "%d резервных копий повреждено",
		"valid_backups":           "%d резервных копий в порядке",
		"backups_files_size":      "%d резервных копий, %d файлов, %s",
		"fix_issue":               "Исправить",
		"suggest_admin":           "Перезапустить программу от имени администратора",
		"suggest_check_permissions": "Проверьте права на папку или запустите от администратора",
		"suggest_env_missing":     "Переменная среды отсутствует; перезайдите в систему",
		"suggest_remove_damaged":  "Перейдите в Очистку резервных копий и удалите повреждённые",
		"admin_required":          "Требуются права администратора",
		"admin_required_desc":       "Некоторые пути очистки требуют повышенных прав. Перезапустите от имени администратора, чтобы продолжить.",
		"restart_as_admin":          "Перезапустить от администратора",
		"exit":                      "Выйти",
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
