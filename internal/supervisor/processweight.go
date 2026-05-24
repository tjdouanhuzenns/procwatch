package supervisor

import (
	"errors"
	"sync"
)

// ProcessWeight stores a numeric scheduling weight for each process.
// Higher weight means the process is given higher priority when resources
// are allocated or restart order is determined.
type ProcessWeight struct {
	mu      sync.RWMutex
	weights map[string]int
}

// NewProcessWeight returns an initialised ProcessWeight store.
func NewProcessWeight() *ProcessWeight {
	return &ProcessWeight{
		weights: make(map[string]int),
	}
}

// Set assigns a weight to a named process. Weight must be >= 0.
func (pw *ProcessWeight) Set(name string, weight int) error {
	if name == "" {
		return errors.New("process name must not be empty")
	}
	if weight < 0 {
		return errors.New("weight must be non-negative")
	}
	pw.mu.Lock()
	defer pw.mu.Unlock()
	pw.weights[name] = weight
	return nil
}

// Get returns the weight for a process and whether it was found.
func (pw *ProcessWeight) Get(name string) (int, bool) {
	pw.mu.RLock()
	defer pw.mu.RUnlock()
	w, ok := pw.weights[name]
	return w, ok
}

// Remove deletes the weight entry for a process.
func (pw *ProcessWeight) Remove(name string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	delete(pw.weights, name)
}

// All returns a snapshot of all process weights.
func (pw *ProcessWeight) All() map[string]int {
	pw.mu.RLock()
	defer pw.mu.RUnlock()
	out := make(map[string]int, len(pw.weights))
	for k, v := range pw.weights {
		out[k] = v
	}
	return out
}
