package risk

import (
	"testing"

	"github.com/elev1e1nSure/broominal/pkg/types"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		category string
		want     types.RiskLevel
	}{
		// System paths -> danger
		{name: "system32", path: `C:\Windows\System32\kernel32.dll`, category: "temp", want: types.RiskDanger},
		{name: "syswow64", path: `C:\Windows\SysWOW64\foo.dll`, category: "temp", want: types.RiskDanger},
		{name: "windows dir", path: `C:\Windows\explorer.exe`, category: "downloads", want: types.RiskDanger},
		{name: "program files", path: `C:\Program Files\App\app.exe`, category: "temp", want: types.RiskDanger},
		{name: "drivers", path: `C:\Windows\System32\drivers\foo.sys`, category: "temp", want: types.RiskDanger},
		{name: "winSxS", path: `C:\Windows\WinSxS\manifests\foo.manifest`, category: "temp", want: types.RiskDanger},
		{name: "sysnative", path: `C:\Windows\sysnative\bar.dll`, category: "temp", want: types.RiskDanger},

		// System extensions -> danger
		{name: "sys ext", path: `C:\foo.sys`, category: "downloads", want: types.RiskDanger},
		{name: "dll ext", path: `C:\foo.dll`, category: "downloads", want: types.RiskDanger},
		{name: "drv ext", path: `C:\foo.drv`, category: "downloads", want: types.RiskDanger},
		{name: "ocx ext", path: `C:\foo.ocx`, category: "downloads", want: types.RiskDanger},

		// Categories -> safe
		{name: "temp category", path: `C:\Temp\foo.txt`, category: "temp", want: types.RiskSafe},
		{name: "browser_cache category", path: `C:\cache\data`, category: "browser_cache", want: types.RiskSafe},
		{name: "recycle_bin category", path: `C:\$Recycle.Bin\file`, category: "recycle_bin", want: types.RiskSafe},
		{name: "logs category", path: `C:\logs\app.log`, category: "logs", want: types.RiskSafe},

		// Categories -> review
		{name: "downloads category", path: `C:\Users\x\Downloads\file.zip`, category: "downloads", want: types.RiskReview},
		{name: "old_installers", path: `C:\old\setup.exe`, category: "old_installers", want: types.RiskReview},
		{name: "large_old_files", path: `C:\big\movie.mkv`, category: "large_old_files", want: types.RiskReview},

		// Default -> review
		{name: "unknown category", path: `C:\misc\file.txt`, category: "unknown", want: types.RiskReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.path, tt.category)
			if got != tt.want {
				t.Errorf("Classify(%q, %q) = %v, want %v", tt.path, tt.category, got, tt.want)
			}
		})
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		level types.RiskLevel
		want  string
	}{
		{types.RiskSafe, "safe"},
		{types.RiskReview, "review"},
		{types.RiskDanger, "danger"},
		{types.RiskLevel("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			got := Label(tt.level)
			if got != tt.want {
				t.Errorf("Label(%v) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestColor(t *testing.T) {
	tests := []struct {
		level types.RiskLevel
		want  string
	}{
		{types.RiskSafe, "#4ade80"},
		{types.RiskReview, "#fbbf24"},
		{types.RiskDanger, "#f87171"},
		{types.RiskLevel("unknown"), "#9ca3af"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			got := Color(tt.level)
			if got != tt.want {
				t.Errorf("Color(%v) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}
