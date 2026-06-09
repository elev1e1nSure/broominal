package risk

import (
	"strings"

	"github.com/elev1e1nSure/pclean/pkg/types"
)

// Classify классифицирует файл по риску на основе пути и категории
func Classify(path string, category string) types.RiskLevel {
	lp := strings.ToLower(path)

	// Системные пути — danger
	systemPaths := []string{
		"system32", "syswow64", "windows\\", "program files",
		"drivers", "\\winSxS", "\\sysnative",
	}
	for _, sp := range systemPaths {
		if strings.Contains(lp, sp) {
			return types.RiskDanger
		}
	}

	// Расширения системных файлов
	sysExts := []string{".sys", ".dll", ".drv", ".ocx"}
	ext := strings.ToLower(getExt(path))
	for _, se := range sysExts {
		if ext == se {
			return types.RiskDanger
		}
	}

	// По категории
	switch category {
	case "temp", "browser_cache", "recycle_bin", "logs":
		return types.RiskSafe
	case "downloads", "old_installers", "large_old_files":
		return types.RiskReview
	default:
		return types.RiskReview
	}
}

func getExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '\\' || path[i] == '/' {
			break
		}
	}
	return ""
}

// Label возвращает человекочитаемую метку
func Label(r types.RiskLevel) string {
	switch r {
	case types.RiskSafe:
		return "safe"
	case types.RiskReview:
		return "review"
	case types.RiskDanger:
		return "danger"
	default:
		return "unknown"
	}
}

// Color возвращает цветовой код для TUI
func Color(r types.RiskLevel) string {
	switch r {
	case types.RiskSafe:
		return "#4ade80" // green
	case types.RiskReview:
		return "#fbbf24" // yellow
	case types.RiskDanger:
		return "#f87171" // red
	default:
		return "#9ca3af" // gray
	}
}
