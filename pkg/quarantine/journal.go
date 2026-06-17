package quarantine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JournalEntry represents a single operation logged in the WAL.
type JournalEntry struct {
	Op          string `json:"op"`          // "begin" or "done"
	Original    string `json:"original"`    // Original file path
	Quarantined string `json:"quarantined"` // Destination file path in quarantine
	Size        int64  `json:"size"`        // File size (used for manifest repair)
	Category    string `json:"category"`    // File category (used for manifest repair)
}

// Journal manages a Write-Ahead Log to track file moves and guarantee 0% data loss.
type Journal struct {
	mu  sync.Mutex
	f   *os.File
	enc *json.Encoder
}

// NewJournal creates a new append-only journal in the given batch directory.
func NewJournal(dir string) (*Journal, error) {
	p := filepath.Join(dir, "journal.jsonl")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("create journal: %w", err)
	}
	return &Journal{
		f:   f,
		enc: json.NewEncoder(f),
	}, nil
}

// Begin logs the intent to move a file.
func (j *Journal) Begin(original, quarantined string, size int64, category string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry := JournalEntry{
		Op:          "begin",
		Original:    original,
		Quarantined: quarantined,
		Size:        size,
		Category:    category,
	}
	return j.enc.Encode(entry)
}

// Commit logs the successful completion of a file move.
func (j *Journal) Commit(original, quarantined string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry := JournalEntry{
		Op:          "done",
		Original:    original,
		Quarantined: quarantined,
	}
	return j.enc.Encode(entry)
}

// Close closes the journal file.
func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.f.Close()
}
