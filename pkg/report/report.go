package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elev1e1nSure/broominal/pkg/config"
	"github.com/elev1e1nSure/broominal/pkg/types"
	"github.com/elev1e1nSure/broominal/pkg/util"
)

// BaseDir возвращает директорию отчётов
func BaseDir() string {
	return filepath.Join(config.AppDir(), "reports")
}

// Save сохраняет отчёт в JSON и выводит summary в stdout
func Save(result *types.ScanResult, cleaned *types.CleanResult) (string, error) {
	if err := os.MkdirAll(BaseDir(), 0700); err != nil {
		return "", fmt.Errorf("create reports dir: %w", err)
	}

	data := types.ReportData{
		Timestamp: time.Now(),
		Result:    *result,
		Cleaned:   cleaned,
	}

	filename := fmt.Sprintf("report_%s.json", time.Now().Format("20060102_150405.000"))
	path := filepath.Join(BaseDir(), filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create report: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return "", fmt.Errorf("encode report: %w", err)
	}

	return path, nil
}

// PrintSummary выводит краткую сводку в stdout
func PrintSummary(result *types.ScanResult, cleaned *types.CleanResult) {
	fmt.Println()
	fmt.Println("═ Broominal Report ════════════════════════")
	fmt.Printf("  Total found:    %s\n", util.FormatSize(result.TotalSize))
	fmt.Printf("  Safe:           %s\n", util.FormatSize(result.SafeSize))
	fmt.Printf("  Review:         %s\n", util.FormatSize(result.ReviewSize))
	fmt.Printf("  Danger:         %s\n", util.FormatSize(result.DangerSize))
	if cleaned != nil {
		fmt.Println()
		fmt.Printf("  Freed:          %s\n", util.FormatSize(cleaned.Freed))
		fmt.Printf("  Files removed:  %d\n", cleaned.Files)
		fmt.Printf("  Restore ID:     %s\n", cleaned.RestoreID)
	}
	fmt.Println("═══════════════════════════════════════")
}
