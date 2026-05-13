package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// ProcessStatus represents the last-known status of a supervised process.
type ProcessStatus struct {
	Name      string    `json:"name"`
	State     string    `json:"state"`
	PID       int       `json:"pid,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StatusReporter stores and retrieves process status entries.
type StatusReporter struct {
	mu       sync.RWMutex
	statuses map[string]ProcessStatus
}

// NewStatusReporter returns an initialised StatusReporter.
func NewStatusReporter() *StatusReporter {
	return &StatusReporter{statuses: make(map[string]ProcessStatus)}
}

// Update stores or overwrites the status for a process.
func (s *StatusReporter) Update(name string, ps ProcessStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps.UpdatedAt = time.Now()
	s.statuses[name] = ps
}

// Get retrieves the status for a named process.
func (s *StatusReporter) Get(name string) (ProcessStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ps, ok := s.statuses[name]
	if !ok {
		return ProcessStatus{}, fmt.Errorf("no status for process %q", name)
	}
	return ps, nil
}

// All returns a slice of all current process statuses.
func (s *StatusReporter) All() []ProcessStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessStatus, 0, len(s.statuses))
	for _, ps := range s.statuses {
		out = append(out, ps)
	}
	return out
}

// Remove deletes the status entry for a process.
func (s *StatusReporter) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.statuses, name)
}
