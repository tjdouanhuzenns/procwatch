package supervisor

import (
	"errors"
	"sync"
	"time"
)

// ResourceLimits holds CPU and memory constraints for a process.
type ResourceLimits struct {
	MaxCPUPercent float64
	MaxMemoryMB   uint64
	UpdatedAt     time.Time
}

// ProcessResourceStore tracks resource limits per process.
type ProcessResourceStore struct {
	mu      sync.RWMutex
	limits  map[string]ResourceLimits
}

// NewProcessResourceStore creates an empty ProcessResourceStore.
func NewProcessResourceStore() *ProcessResourceStore {
	return &ProcessResourceStore{
		limits: make(map[string]ResourceLimits),
	}
}

// Set stores resource limits for the named process.
func (s *ProcessResourceStore) Set(name string, limits ResourceLimits) error {
	if name == "" {
		return errors.New("process name must not be empty")
	}
	if limits.MaxCPUPercent < 0 {
		return errors.New("MaxCPUPercent must be non-negative")
	}
	if limits.MaxMemoryMB == 0 && limits.MaxCPUPercent == 0 {
		return errors.New("at least one resource limit must be set")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	limits.UpdatedAt = time.Now()
	s.limits[name] = limits
	return nil
}

// Get returns the resource limits for the named process.
func (s *ProcessResourceStore) Get(name string) (ResourceLimits, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.limits[name]
	return l, ok
}

// Remove deletes the resource limits for the named process.
func (s *ProcessResourceStore) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.limits, name)
}

// All returns a snapshot of all stored resource limits.
func (s *ProcessResourceStore) All() map[string]ResourceLimits {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]ResourceLimits, len(s.limits))
	for k, v := range s.limits {
		out[k] = v
	}
	return out
}
