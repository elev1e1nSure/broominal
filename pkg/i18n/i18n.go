package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/categories"
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
