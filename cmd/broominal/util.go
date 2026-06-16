package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/i18n"
	"github.com/elev1e1nSure/broominal/pkg/types"
)

// cliDate formats a timestamp for CLI output.
// Same day → "Today 16:55", yesterday → "Yesterday 12:18", older → "2025-06-14 09:00".
func cliDate(t time.Time) string {
	now := time.Now()
	ty, tm, td := t.Date()
	ny, nm, nd := now.Date()
	if ty == ny && tm == nm && td == nd {
		return "Today " + t.Format("15:04")
	}
	py, pm, pd := now.AddDate(0, 0, -1).Date()
	if ty == py && tm == pm && td == pd {
		return "Yesterday " + t.Format("15:04")
	}
	return t.Format("2006-01-02 15:04")
}

// applyPresetFlag applies the --preset flag value to cfg.
// An empty string is a no-op (keeps the config's active preset).
func applyPresetFlag(cfg *config.Config, preset string) error {
	if preset == "" {
		return nil
	}
	switch preset {
	case "quick":
		cfg.ApplyPreset(config.PresetQuick)
	case "standard":
		cfg.ApplyPreset(config.PresetStandard)
	case "deep":
		cfg.ApplyPreset(config.PresetDeep)
	default:
		return fmt.Errorf("unknown preset %q — use quick, standard, or deep", preset)
	}
	return nil
}

// riskDisplay returns a label and a style function for a risk level.
func riskDisplay(r types.RiskLevel) (string, func(string, ...any) string) {
	switch r {
	case types.RiskSafe:
		return "safe", func(f string, a ...any) string { return "\033[32m" + fmt.Sprintf(f, a...) + "\033[0m" }
	case types.RiskReview:
		return "review", func(f string, a ...any) string { return "\033[33m" + fmt.Sprintf(f, a...) + "\033[0m" }
	case types.RiskDanger:
		return "danger", func(f string, a ...any) string { return "\033[31m" + fmt.Sprintf(f, a...) + "\033[0m" }
	default:
		return string(r), func(f string, a ...any) string { return fmt.Sprintf(f, a...) }
	}
}

// formatCLICategories returns up to 3 translated category names joined by " · " with "+N" suffix.
func formatCLICategories(cats []string) string {
	if len(cats) == 0 {
		return ""
	}
	const maxShow = 3
	shown := cats
	extra := 0
	if len(cats) > maxShow {
		shown = cats[:maxShow]
		extra = len(cats) - maxShow
	}
	names := make([]string, len(shown))
	for i, c := range shown {
		names[i] = i18n.CategoryName(c)
	}
	result := strings.Join(names, " · ")
	if extra > 0 {
		result += fmt.Sprintf(" +%d", extra)
	}
	return result
}

// truncateCLI truncates s to at most n runes for table alignment.
func truncateCLI(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// pluralize returns singular or plural based on n.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
