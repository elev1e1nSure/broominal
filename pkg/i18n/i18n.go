package i18n

import (
	"strings"
	"sync"

	"github.com/elev1e1nSure/broominal/pkg/categories"
	"golang.org/x/sys/windows/registry"
)

// translations is built from per-language string maps defined in strings_*.go.
var translations = map[string]map[string]string{
	"en": enStrings,
	"ru": ruStrings,
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

// CategoryName returns the translated display name for a scanner category.
func CategoryName(name string) string {
	for _, def := range categories.All {
		if def.Name == name {
			return T("cat_" + def.InternalKey)
		}
	}
	return name
}

// CategoryDescription returns a brief description of what a category contains and its safety.
func CategoryDescription(name string) string {
	for _, def := range categories.All {
		if def.Name == name {
			return T("cat_desc_" + def.InternalKey)
		}
	}
	return ""
}

func DetectFromWindowsLocale() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\International`, registry.QUERY_VALUE)
	if err != nil {
		return "en"
	}
	defer k.Close()
	locale, _, err := k.GetStringValue("LocaleName")
	if err != nil {
		return "en"
	}
	locale = strings.ToLower(locale)
	if strings.HasPrefix(locale, "ru") || strings.HasPrefix(locale, "uk") || strings.HasPrefix(locale, "be") {
		return "ru"
	}
	return "en"
}
