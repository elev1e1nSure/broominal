package types

import "time"

// RiskLevel оценка риска удаления
type RiskLevel string

const (
	RiskSafe   RiskLevel = "safe"
	RiskReview RiskLevel = "review"
	RiskDanger RiskLevel = "danger"
)

// Item найденный файл или группа файлов
type Item struct {
	Category string    `json:"category"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Risk     RiskLevel `json:"risk"`
	Selected bool      `json:"selected"`
}

// CategorySummary сводка по категории
type CategorySummary struct {
	Category string    `json:"category"`
	Size     int64     `json:"size"`
	Files    int       `json:"files"`
	Risk     RiskLevel `json:"risk"`
	Items    []Item    `json:"items"`
}

// Manifest запись quarantine
type Manifest struct {
	ID        string         `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	Label     string         `json:"label"`
	TotalSize int64          `json:"total_size"`
	Files     int            `json:"files"`
	Items     []ManifestItem `json:"items"`
}

// ManifestItem запись в манифесте
type ManifestItem struct {
	Original    string `json:"original"`
	Quarantined string `json:"quarantined"`
	Size        int64  `json:"size"`
}

// ScanResult результат сканирования
type ScanResult struct {
	Categories []CategorySummary `json:"categories"`
	TotalSize  int64             `json:"total_size"`
	SafeSize   int64             `json:"safe_size"`
	ReviewSize int64             `json:"review_size"`
	DangerSize int64             `json:"danger_size"`
}

// ReportData данные отчёта
type ReportData struct {
	Timestamp time.Time    `json:"timestamp"`
	Result    ScanResult   `json:"result"`
	Cleaned   *CleanResult `json:"cleaned,omitempty"`
}

// CleanResult результат очистки
type CleanResult struct {
	RestoreID string `json:"restore_id"`
	Freed     int64  `json:"freed"`
	Files     int    `json:"files"`
	Skipped   int    `json:"skipped"`
}
