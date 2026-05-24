package supervisor

import (
	"errors"
	"sync"
	"time"
)

// TraceEntry records a single trace span for a process operation.
type TraceEntry struct {
	Process   string
	Operation string
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
	Error     string
}

// ProcessTracingStore records and retrieves trace spans per process.
type ProcessTracingStore struct {
	mu      sync.RWMutex
	entries []TraceEntry
	maxSize int
}

// NewProcessTracingStore creates a store with the given max capacity.
func NewProcessTracingStore(maxSize int) *ProcessTracingStore {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &ProcessTracingStore{maxSize: maxSize}
}

// Record adds a trace entry. StartedAt and Duration are set automatically if EndedAt is set.
func (s *ProcessTracingStore) Record(process, operation string, start, end time.Time, err error) error {
	if process == "" {
		return errors.New("process name required")
	}
	if operation == "" {
		return errors.New("operation name required")
	}
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	entry := TraceEntry{
		Process:   process,
		Operation: operation,
		StartedAt: start,
		EndedAt:   end,
		Duration:  end.Sub(start),
		Error:     errStr,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.maxSize {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
	return nil
}

// ForProcess returns all trace entries for the given process.
func (s *ProcessTracingStore) ForProcess(process string) []TraceEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []TraceEntry
	for _, e := range s.entries {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// All returns a copy of all trace entries.
func (s *ProcessTracingStore) All() []TraceEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]TraceEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Len returns the current number of stored entries.
func (s *ProcessTracingStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
