package i18n

import (
	"testing"
)

func TestSetLanguage(t *testing.T) {
	SetLanguage("ru")
	if CurrentLanguage() != "ru" {
		t.Errorf("language = %q, want ru", CurrentLanguage())
	}

	SetLanguage("en")
	if CurrentLanguage() != "en" {
		t.Errorf("language = %q, want en", CurrentLanguage())
	}

	SetLanguage("unknown")
	if CurrentLanguage() != "en" {
		t.Errorf("unknown lang should fallback to en, got %q", CurrentLanguage())
	}
}

func TestT(t *testing.T) {
	SetLanguage("en")
	if got := T("main_menu"); got != "Broominal — Main Menu" {
		t.Errorf("en T(main_menu) = %q", got)
	}

	SetLanguage("ru")
	if got := T("main_menu"); got != "Broominal — Главное меню" {
		t.Errorf("ru T(main_menu) = %q", got)
	}

	// fallback to key if missing
	SetLanguage("en")
	if got := T("nonexistent_key"); got != "nonexistent_key" {
		t.Errorf("missing key fallback = %q, want key itself", got)
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) != 2 {
		t.Errorf("expected 2 languages, got %d", len(langs))
	}
	found := map[string]bool{}
	for _, l := range langs {
		found[l] = true
	}
	if !found["en"] || !found["ru"] {
		t.Error("expected en and ru in supported languages")
	}
}

func TestDetectFromIP(t *testing.T) {
	lang, err := DetectFromIP()
	if err != nil {
		t.Skipf("offline or geo-ip service unavailable: %v", err)
	}
	found := false
	for _, l := range SupportedLanguages() {
		if l == lang {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("detected language %q not in supported list", lang)
	}
}
