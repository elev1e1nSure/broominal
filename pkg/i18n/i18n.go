package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var translations = map[string]map[string]string{
	"en": {
		"main_menu":                    "Broominal — Main Menu",
		"menu_scan_clean":              "Scan & Clean",
		"menu_restore":                 "Restore from Quarantine",
		"menu_doctor":                  "Doctor (Health Checks)",
		"menu_config":                  "View Config",
		"menu_cleanup":                 "Quarantine Cleanup",
		"menu_settings":                "Settings",
		"hint_select":                  "Enter: select  Q: quit",
		"dashboard":                     "Broominal — Dashboard",
		"total_found":                   "Total found",
		"safe":                          "Safe",
		"review":                        "Review",
		"danger":                        "Danger",
		"hint_continue":                 "Press Enter or Space to continue",
		"categories":                    "Categories",
		"category":                      "Category",
		"size":                          "Size",
		"files":                         "Files",
		"risk":                          "Risk",
		"select":                        "Select",
		"hint_categories":               "Space: toggle  Enter: confirm  D: details  Q: quit",
		"details":                       "Details",
		"hint_back":                     "Q/Esc: back",
		"confirm_cleanup":               "Confirm Cleanup",
		"dry_run":                       "[DRY-RUN]",
		"hint_confirm":                  "Enter: proceed  T: toggle dry-run  Esc: back",
		"will_free":                     "Will free",
		"cleaning":                      "Cleaning...",
		"cleanup_complete":              "Cleanup Complete",
		"freed":                         "Freed",
		"restore_id":                    "Restore",
		"hint_result":                   "R: restore last  Q: quit",
		"dry_run_complete":              "Dry-Run Complete",
		"would_free":                    "Would free",
		"restored":                      "Restored",
		"hint_restored":                 "Files restored successfully.",
		"restore_conflicts":             "Restore Conflicts",
		"files_already_exist":           "file(s) already exist at original paths",
		"hint_conflicts":                "O: overwrite all  S: skip all  C/Esc: cancel",
		"warning":                       "Warning",
		"recycle_bin_warn":              "Recycle Bin contains %d files.",
		"hint_recycle_warn":             "Opening details may be very slow.",
		"hint_recycle_continue":         "Enter: continue anyway  Esc: back",
		"error":                         "Error",
		"hint_error_quit":               "Press Q or Esc to quit",
		"scanning":                      "Scanning...",
		"moving_files":                  "Moving files to quarantine...",
		"please_wait":                   "Please wait, this may take a while.",
		"quit":                          "quit",
		"restore":                       "Restore from Quarantine",
		"no_quarantines":                "No quarantines available.",
		"hint_restore":                  "Enter: restore  Q/Esc: back",
		"restored_n_skipped":            "Restored %d files (%d skipped)",
		"doctor":                        "Doctor — Health Checks",
		"config":                        "Current Config",
		"quarantine_cleanup":            "Quarantine Cleanup",
		"cleanup_desc":                  "Deletes quarantines older than max age days.",
		"hint_cleanup":                  "T: toggle dry-run  Enter: proceed  Q/Esc: back",
		"settings":                      "Settings",
		"language":                      "Language",
		"select_language":               "Select Language",
		"hint_language":                 "Enter: select  Q/Esc: back",
		"english":                       "English",
		"russian":                       "Russian",
		"save":                          "Save",
		"continue":                      "continue",
		"toggle":                        "toggle",
		"confirm":                       "confirm",
		"proceed":                       "proceed",
		"back":                          "back",
		"overwrite_all":                 "overwrite all",
		"skip_all":                      "skip all",
		"cancel":                        "cancel",
		"continue_anyway":               "continue anyway",
		"toggle_dry_run":                "toggle dry-run",
		"restore_last":                  "restore last",
	},
	"ru": {
		"main_menu":                    "Broominal — Главное меню",
		"menu_scan_clean":              "Сканирование и очистка",
		"menu_restore":                 "Восстановление из карантина",
		"menu_doctor":                  "Доктор (Проверка здоровья)",
		"menu_config":                  "Просмотр конфигурации",
		"menu_cleanup":                 "Очистка карантина",
		"menu_settings":                "Настройки",
		"hint_select":                  "Enter: выбрать  Q: выйти",
		"dashboard":                     "Broominal — Обзор",
		"total_found":                   "Всего найдено",
		"safe":                          "Безопасно",
		"review":                        "Проверить",
		"danger":                        "Опасно",
		"hint_continue":                 "Нажмите Enter или Space для продолжения",
		"categories":                    "Категории",
		"category":                      "Категория",
		"size":                          "Размер",
		"files":                         "Файлы",
		"risk":                          "Риск",
		"select":                        "Выбор",
		"hint_categories":               "Space: выбрать  Enter: подтвердить  D: детали  Q: выйти",
		"details":                       "Детали",
		"hint_back":                     "Q/Esc: назад",
		"confirm_cleanup":               "Подтвердить очистку",
		"dry_run":                       "[ТЕСТОВЫЙ РЕЖИМ]",
		"hint_confirm":                  "Enter: продолжить  T: тестовый режим  Esc: назад",
		"will_free":                     "Будет освобождено",
		"cleaning":                      "Очистка...",
		"cleanup_complete":              "Очистка завершена",
		"freed":                         "Освобождено",
		"restore_id":                    "Восстановление",
		"hint_result":                   "R: восстановить последнее  Q: выйти",
		"dry_run_complete":              "Тестовый режим завершен",
		"would_free":                    "Будет освобождено",
		"restored":                      "Восстановлено",
		"hint_restored":                 "Файлы успешно восстановлены.",
		"restore_conflicts":             "Конфликты восстановления",
		"files_already_exist":           "файл(ов) уже существуют по оригинальным путям",
		"hint_conflicts":                "O: перезаписать  S: пропустить  C/Esc: отмена",
		"warning":                       "Предупреждение",
		"recycle_bin_warn":              "Корзина содержит %d файлов.",
		"hint_recycle_warn":             "Открытие деталей может быть очень медленным.",
		"hint_recycle_continue":         "Enter: продолжить  Esc: назад",
		"error":                         "Ошибка",
		"hint_error_quit":               "Нажмите Q или Esc для выхода",
		"scanning":                      "Сканирование...",
		"moving_files":                  "Перемещение файлов в карантин...",
		"please_wait":                   "Пожалуйста, подождите, это может занять некоторое время.",
		"quit":                          "выйти",
		"restore":                       "Восстановление из карантина",
		"no_quarantines":                "Карантин пуст.",
		"hint_restore":                  "Enter: восстановить  Q/Esc: назад",
		"restored_n_skipped":            "Восстановлено %d файлов (%d пропущено)",
		"doctor":                        "Доктор — Проверка здоровья",
		"config":                        "Текущая конфигурация",
		"quarantine_cleanup":            "Очистка карантина",
		"cleanup_desc":                  "Удаляет карантин старше max age дней.",
		"hint_cleanup":                  "T: тестовый режим  Enter: запустить  Q/Esc: назад",
		"settings":                      "Настройки",
		"language":                      "Язык",
		"select_language":               "Выбор языка",
		"hint_language":                 "Enter: выбрать  Q/Esc: назад",
		"english":                       "English",
		"russian":                       "Русский",
		"save":                          "Сохранить",
		"continue":                      "продолжить",
		"toggle":                        "выбрать",
		"confirm":                       "подтвердить",
		"proceed":                       "продолжить",
		"back":                          "назад",
		"overwrite_all":                 "перезаписать все",
		"skip_all":                      "пропустить все",
		"cancel":                        "отмена",
		"continue_anyway":               "всё равно продолжить",
		"toggle_dry_run":                "тестовый режим",
		"restore_last":                  "восстановить последнее",
	},
}

var currentLang = "en"

func SetLanguage(lang string) {
	if _, ok := translations[lang]; ok {
		currentLang = lang
	} else {
		currentLang = "en"
	}
}

func CurrentLanguage() string {
	return currentLang
}

func T(key string) string {
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
