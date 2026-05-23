package supervisor

import (
	"sync"
	"time"
)

// OutputEntry holds a single captured line of process output.
type OutputEntry struct {
	Process   string    `json:"process"`
	Line      string    `json:"line"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Timestamp time.Time `json:"timestamp"`
}

// OutputBuffer stores recent output lines per process with a fixed capacity.
type OutputBuffer struct {
	mu      sync.RWMutex
	entries []OutputEntry
	maxSize int
}

// NewOutputBuffer creates an OutputBuffer that retains at most maxSize entries.
func NewOutputBuffer(maxSize int) *OutputBuffer {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &OutputBuffer{
		entries: make([]OutputEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record appends an output line for the given process and stream.
func (b *OutputBuffer) Record(process, stream, line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entry := OutputEntry{
		Process:   process,
		Line:      line,
		Stream:    stream,
		Timestamp: time.Now().UTC(),
	}
	if len(b.entries) >= b.maxSize {
		copy(b.entries, b.entries[1:])
		b.entries[len(b.entries)-1] = entry
	} else {
		b.entries = append(b.entries, entry)
	}
}

// ForProcess returns all buffered entries for the named process.
func (b *OutputBuffer) ForProcess(process string) []OutputEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []OutputEntry
	for _, e := range b.entries {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// All returns a copy of all buffered entries across every process.
func (b *OutputBuffer) All() []OutputEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]OutputEntry, len(b.entries))
	copy(out, b.entries)
	return out
}

// Len returns the current number of stored entries.
func (b *OutputBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries)
}

// Clear removes all entries from the buffer.
func (b *OutputBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
}
