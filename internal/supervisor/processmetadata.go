package supervisor

import (
	"errors"
	"sync"
	"time"
)

// ProcessMetadata holds arbitrary key-value annotations attached to a process.
type ProcessMetadata struct {
	Process   string
	Key       string
	Value     string
	UpdatedAt time.Time
}

// ProcessMetadataStore stores metadata entries keyed by process name and key.
type ProcessMetadataStore struct {
	mu      sync.RWMutex
	entries map[string]map[string]ProcessMetadata
}

// NewProcessMetadataStore creates an empty ProcessMetadataStore.
func NewProcessMetadataStore() *ProcessMetadataStore {
	return &ProcessMetadataStore{
		entries: make(map[string]map[string]ProcessMetadata),
	}
}

// Set stores a metadata key-value pair for the given process.
func (s *ProcessMetadataStore) Set(process, key, value string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if key == "" {
		return errors.New("metadata key must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[process]; !ok {
		s.entries[process] = make(map[string]ProcessMetadata)
	}
	s.entries[process][key] = ProcessMetadata{
		Process:   process,
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get retrieves a metadata value for the given process and key.
func (s *ProcessMetadataStore) Get(process, key string) (ProcessMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.entries[process]; ok {
		if entry, ok := m[key]; ok {
			return entry, true
		}
	}
	return ProcessMetadata{}, false
}

// ForProcess returns all metadata entries for a given process.
func (s *ProcessMetadataStore) ForProcess(process string) []ProcessMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.entries[process]
	if !ok {
		return nil
	}
	out := make([]ProcessMetadata, 0, len(m))
	for _, entry := range m {
		out = append(out, entry)
	}
	return out
}

// Delete removes a specific metadata key for a process.
func (s *ProcessMetadataStore) Delete(process, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.entries[process]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(s.entries, process)
		}
	}
}

// All returns every metadata entry across all processes.
func (s *ProcessMetadataStore) All() []ProcessMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ProcessMetadata
	for _, m := range s.entries {
		for _, entry := range m {
			out = append(out, entry)
		}
	}
	return out
}
