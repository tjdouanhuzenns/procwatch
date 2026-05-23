package supervisor

import (
	"errors"
	"sync"
)

// ProcessLabelStore stores arbitrary string labels keyed by process name and label key.
type ProcessLabelStore struct {
	mu     sync.RWMutex
	labels map[string]map[string]string // process -> key -> value
}

// NewProcessLabelStore returns an initialised ProcessLabelStore.
func NewProcessLabelStore() *ProcessLabelStore {
	return &ProcessLabelStore{
		labels: make(map[string]map[string]string),
	}
}

// Set adds or overwrites a label key/value for the given process.
func (s *ProcessLabelStore) Set(process, key, value string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if key == "" {
		return errors.New("label key must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.labels[process] == nil {
		s.labels[process] = make(map[string]string)
	}
	s.labels[process][key] = value
	return nil
}

// Get returns the value for a label key on a process, and whether it was found.
func (s *ProcessLabelStore) Get(process, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.labels[process]; ok {
		v, found := m[key]
		return v, found
	}
	return "", false
}

// All returns a copy of all labels for the given process.
func (s *ProcessLabelStore) All(process string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string)
	for k, v := range s.labels[process] {
		out[k] = v
	}
	return out
}

// Remove deletes a single label key from a process.
func (s *ProcessLabelStore) Remove(process, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.labels[process]; ok {
		delete(m, key)
	}
}

// RemoveAll deletes every label for the given process.
func (s *ProcessLabelStore) RemoveAll(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.labels, process)
}
