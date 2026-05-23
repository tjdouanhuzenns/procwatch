package supervisor

import (
	"errors"
	"sync"
	"time"
)

// ProcessStatusEntry holds a named status string and when it was last updated.
type ProcessStatusEntry struct {
	Process   string
	Status    string
	UpdatedAt time.Time
}

// ProcessStatusStore tracks arbitrary named status strings per process.
type ProcessStatusStore struct {
	mu      sync.RWMutex
	entries map[string]ProcessStatusEntry
}

// NewProcessStatusStore returns an initialised ProcessStatusStore.
func NewProcessStatusStore() *ProcessStatusStore {
	return &ProcessStatusStore{
		entries: make(map[string]ProcessStatusEntry),
	}
}

// Set records a status string for the named process.
func (s *ProcessStatusStore) Set(process, status string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if status == "" {
		return errors.New("status must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[process] = ProcessStatusEntry{
		Process:   process,
		Status:    status,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get retrieves the current status entry for the named process.
func (s *ProcessStatusStore) Get(process string) (ProcessStatusEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[process]
	return e, ok
}

// Remove deletes the status entry for the named process.
func (s *ProcessStatusStore) Remove(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, process)
}

// All returns a copy of all current status entries.
func (s *ProcessStatusStore) All() []ProcessStatusEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessStatusEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
