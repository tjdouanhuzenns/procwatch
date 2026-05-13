package supervisor

import (
	"sync"
	"time"
)

// ProcessStatus represents the current status of a supervised process.
type ProcessStatus struct {
	Name      string    `json:"name"`
	State     string    `json:"state"`
	PID       int       `json:"pid,omitempty"`
	Restarts  int       `json:"restarts"`
	StartedAt time.Time `json:"started_at,omitempty"`
	StoppedAt time.Time `json:"stopped_at,omitempty"`
}

// StatusReporter tracks and exposes process statuses.
type StatusReporter struct {
	mu       sync.RWMutex
	statuses map[string]*ProcessStatus
}

// NewStatusReporter creates a new StatusReporter.
func NewStatusReporter() *StatusReporter {
	return &StatusReporter{
		statuses: make(map[string]*ProcessStatus),
	}
}

// Update sets or updates the status for a named process.
func (s *StatusReporter) Update(name, state string, pid, restarts int, startedAt, stoppedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[name] = &ProcessStatus{
		Name:      name,
		State:     state,
		PID:       pid,
		Restarts:  restarts,
		StartedAt: startedAt,
		StoppedAt: stoppedAt,
	}
}

// Get returns the status for a named process, and whether it was found.
func (s *StatusReporter) Get(name string) (ProcessStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ps, ok := s.statuses[name]
	if !ok {
		return ProcessStatus{}, false
	}
	return *ps, true
}

// All returns a snapshot of all current process statuses.
func (s *StatusReporter) All() []ProcessStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessStatus, 0, len(s.statuses))
	for _, ps := range s.statuses {
		out = append(out, *ps)
	}
	return out
}

// Remove deletes the status entry for a named process.
func (s *StatusReporter) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.statuses, name)
}
