package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// ProcessVersion holds version metadata for a managed process.
type ProcessVersion struct {
	Name      string
	Version   string
	Commit    string
	BuiltAt   time.Time
	UpdatedAt time.Time
}

// ProcessVersionStore tracks version metadata keyed by process name.
type ProcessVersionStore struct {
	mu      sync.RWMutex
	entries map[string]ProcessVersion
}

// NewProcessVersionStore returns an empty ProcessVersionStore.
func NewProcessVersionStore() *ProcessVersionStore {
	return &ProcessVersionStore{
		entries: make(map[string]ProcessVersion),
	}
}

// Set records version metadata for the given process.
func (s *ProcessVersionStore) Set(name, version, commit string, builtAt time.Time) error {
	if name == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if version == "" {
		return fmt.Errorf("version must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[name] = ProcessVersion{
		Name:      name,
		Version:   version,
		Commit:    commit,
		BuiltAt:   builtAt,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get returns the version metadata for the given process.
func (s *ProcessVersionStore) Get(name string) (ProcessVersion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.entries[name]
	return v, ok
}

// Remove deletes the version record for the given process.
func (s *ProcessVersionStore) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, name)
}

// All returns a snapshot of all version records.
func (s *ProcessVersionStore) All() []ProcessVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessVersion, 0, len(s.entries))
	for _, v := range s.entries {
		out = append(out, v)
	}
	return out
}
