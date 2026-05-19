package supervisor

import (
	"fmt"
	"sync"
)

// ProcessRegistry maintains a thread-safe registry of active process names
// and their associated metadata tags, enabling lookup and enumeration.
type ProcessRegistry struct {
	mu      sync.RWMutex
	entries map[string]RegistryEntry
}

// RegistryEntry holds metadata associated with a registered process.
type RegistryEntry struct {
	Name    string
	Command string
	Tags    map[string]string
}

// NewProcessRegistry returns an initialised ProcessRegistry.
func NewProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		entries: make(map[string]RegistryEntry),
	}
}

// Register adds or replaces an entry for the given process name.
// Returns an error if name or command is empty.
func (r *ProcessRegistry) Register(name, command string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if command == "" {
		return fmt.Errorf("command must not be empty for process %q", name)
	}

	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[name] = RegistryEntry{Name: name, Command: command, Tags: copy}
	return nil
}

// Deregister removes the entry for the given process name.
// It is a no-op if the name is not registered.
func (r *ProcessRegistry) Deregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, name)
}

// Get returns the RegistryEntry for name and whether it was found.
func (r *ProcessRegistry) Get(name string) (RegistryEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[name]
	return e, ok
}

// All returns a snapshot of all registered entries.
func (r *ProcessRegistry) All() []RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RegistryEntry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}

// Len returns the number of registered processes.
func (r *ProcessRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}
