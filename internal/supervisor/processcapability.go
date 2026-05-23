package supervisor

import (
	"errors"
	"sync"
)

// ProcessCapability represents a named capability assigned to a process.
type ProcessCapability struct {
	Process      string
	Capabilities []string
}

// ProcessCapabilityStore tracks capabilities assigned to each process.
type ProcessCapabilityStore struct {
	mu   sync.RWMutex
	data map[string][]string
}

// NewProcessCapabilityStore creates an empty ProcessCapabilityStore.
func NewProcessCapabilityStore() *ProcessCapabilityStore {
	return &ProcessCapabilityStore{
		data: make(map[string][]string),
	}
}

// Set assigns a list of capabilities to the named process, replacing any prior set.
func (s *ProcessCapabilityStore) Set(process string, caps []string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if len(caps) == 0 {
		return errors.New("capabilities must not be empty")
	}
	copy := make([]string, len(caps))
	for i, c := range caps {
		if c == "" {
			return errors.New("capability must not be empty string")
		}
		copy[i] = c
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[process] = copy
	return nil
}

// Get returns the capabilities for the named process.
func (s *ProcessCapabilityStore) Get(process string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	caps, ok := s.data[process]
	if !ok {
		return nil, false
	}
	out := make([]string, len(caps))
	copy(out, caps)
	return out, true
}

// Has reports whether the named process has a specific capability.
func (s *ProcessCapabilityStore) Has(process, cap string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.data[process] {
		if c == cap {
			return true
		}
	}
	return false
}

// Remove deletes all capabilities for the named process.
func (s *ProcessCapabilityStore) Remove(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, process)
}

// All returns a snapshot of all process capability assignments.
func (s *ProcessCapabilityStore) All() []ProcessCapability {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessCapability, 0, len(s.data))
	for proc, caps := range s.data {
		cp := make([]string, len(caps))
		copy(cp, caps)
		out = append(out, ProcessCapability{Process: proc, Capabilities: cp})
	}
	return out
}
