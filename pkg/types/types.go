package types

import "time"

// RiskLevel expresses how dangerous it is to delete a file. It drives the UI color
// and gates which items the `clean` command refuses to touch without --danger.
type RiskLevel string

const (
	RiskSafe   RiskLevel = "safe"
	RiskReview RiskLevel = "review"
	RiskDanger RiskLevel = "danger"
)

// Item is a single cleanup candidate produced by a scanner. It points at one
// on-disk path and remembers its category, size, and risk for downstream filtering.
type Item struct {
	Category string    `json:"category"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Risk     RiskLevel `json:"risk"`
	Selected bool      `json:"selected"`
}

// CategorySummary aggregates every item a scanner found under one category name.
// TUI rows and the `scan` CLI summary both render from this struct.
type CategorySummary struct {
	Category string    `json:"category"`
	Size     int64     `json:"size"`
	Files    int       `json:"files"`
	Risk     RiskLevel `json:"risk"`
	Items    []Item    `json:"items"`
}

// Manifest is the per-cleanup record persisted inside the quarantine directory.
// It is the only authoritative source for what a restore has to put back, so it
// must be written atomically and read on every restore attempt.
type Manifest struct {
	ID         string         `json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	Label      string         `json:"label"`
	Categories []string       `json:"categories,omitempty"`
	TotalSize  int64          `json:"total_size"`
	Files      int            `json:"files"`
	Items      []ManifestItem `json:"items"`
}

// ManifestItem pairs a file's original path with the path it now has inside the
// quarantine dir, so a restore can move the bytes back without ambiguity.
type ManifestItem struct {
	Original    string `json:"original"`
	Quarantined string `json:"quarantined"`
	Size        int64  `json:"size"`
}

// ScanResult is the output of a full scan across every enabled category.
// The Safe/Review/Danger size fields are pre-computed so the TUI and the
// report CLI can render risk totals without re-walking the categories.
type ScanResult struct {
	Categories []CategorySummary `json:"categories"`
	TotalSize  int64             `json:"total_size"`
	SafeSize   int64             `json:"safe_size"`
	ReviewSize int64             `json:"review_size"`
	DangerSize int64             `json:"danger_size"`
}

// ReportData bundles a scan with the optional clean result so a single JSON
// report can answer both "what was found" and "what was removed".
type ReportData struct {
	Timestamp time.Time    `json:"timestamp"`
	Result    ScanResult   `json:"result"`
	Cleaned   *CleanResult `json:"cleaned,omitempty"`
}

// CleanResult summarises one clean operation: bytes freed, file count, items
// skipped (e.g. locked or out of policy), and the restore ID for the
// quarantine batch, or empty when quarantine was disabled.
type CleanResult struct {
	RestoreID string `json:"restore_id"`
	Freed     int64  `json:"freed"`
	Files     int    `json:"files"`
	Skipped   int    `json:"skipped"`
}
