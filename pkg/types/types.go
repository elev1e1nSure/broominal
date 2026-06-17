package types

import (
	"fmt"
	"time"
)

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
	Cancelled bool   `json:"cancelled,omitempty"`
}

// Progress carries progress info for long-running operations. Stage is
// "scanning" or "cleaning". Callers compute ETA/throughput from StartedAt
// and the current counts.
type Progress struct {
	Stage      string
	Processed  int
	Total      int
	Bytes      int64
	TotalBytes int64
	StartedAt  time.Time
}

// ProgressFn is a callback for progress updates, throttled by the callee to
// ~250 ms so the consumer does not need its own rate-limiting.
type ProgressFn func(Progress)

// Elapsed returns the wall-clock duration since the operation started.
func (p Progress) Elapsed() time.Duration {
	return time.Since(p.StartedAt).Truncate(time.Second)
}

// ThroughputMBps returns megabytes per second, or 0 when there is no data yet.
func (p Progress) ThroughputMBps() float64 {
	elapsed := time.Since(p.StartedAt).Seconds()
	if elapsed < 0.1 || p.Bytes == 0 {
		return 0
	}
	return float64(p.Bytes) / elapsed / (1024 * 1024)
}

// ETA returns the estimated time remaining, or zero when it cannot be computed.
func (p Progress) ETA() time.Duration {
	if p.Processed == 0 || p.Total == 0 {
		return 0
	}
	elapsed := time.Since(p.StartedAt)
	if elapsed <= 0 {
		return 0
	}
	totalTime := time.Duration(float64(elapsed) / float64(p.Processed) * float64(p.Total))
	remaining := totalTime - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining.Truncate(time.Second)
}

// Percent returns the completion percentage, clamped to [0, 100].
func (p Progress) Percent() int {
	if p.Total == 0 {
		return 0
	}
	v := int(float64(p.Processed) / float64(p.Total) * 100)
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// FormatETA returns a human-readable ETA string (e.g. "1m30s") or empty string.
func (p Progress) FormatETA() string {
	d := p.ETA()
	if d <= 0 {
		return ""
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// FormatThroughput returns a human-readable throughput string (e.g. "12.3 MB/s").
func (p Progress) FormatThroughput() string {
	mbps := p.ThroughputMBps()
	if mbps <= 0 {
		return ""
	}
	if mbps >= 1000 {
		return fmt.Sprintf("%.1f GB/s", mbps/1024)
	}
	if mbps >= 1 {
		return fmt.Sprintf("%.1f MB/s", mbps)
	}
	return fmt.Sprintf("%.1f KB/s", mbps*1024)
}
