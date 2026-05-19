package supervisor

import (
	"sync"
	"time"
)

// RunbookEntry records a notable operational event with a human-readable note.
type RunbookEntry struct {
	Process   string    `json:"process"`
	Event     string    `json:"event"`
	Note      string    `json:"note"`
	Timestamp time.Time `json:"timestamp"`
}

// Runbook stores per-process operational notes keyed by event type.
type Runbook struct {
	mu      sync.RWMutex
	entries []RunbookEntry
	maxSize int
}

// NewRunbook creates a Runbook that retains at most maxSize entries.
func NewRunbook(maxSize int) *Runbook {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &Runbook{maxSize: maxSize}
}

// Record appends a new entry, evicting the oldest when the log is full.
func (r *Runbook) Record(process, event, note string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := RunbookEntry{
		Process:   process,
		Event:     event,
		Note:      note,
		Timestamp: time.Now().UTC(),
	}
	if len(r.entries) >= r.maxSize {
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, entry)
}

// ForProcess returns all entries recorded for the given process name.
func (r *Runbook) ForProcess(process string) []RunbookEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []RunbookEntry
	for _, e := range r.entries {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// All returns a copy of all recorded entries in insertion order.
func (r *Runbook) All() []RunbookEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RunbookEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// Len returns the current number of entries.
func (r *Runbook) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Clear removes all entries from the runbook.
func (r *Runbook) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = nil
}
