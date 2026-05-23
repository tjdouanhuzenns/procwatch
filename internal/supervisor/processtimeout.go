package supervisor

import (
	"errors"
	"sync"
	"time"
)

// ProcessTimeout stores per-process execution timeout configuration.
type ProcessTimeout struct {
	mu      sync.RWMutex
	entries map[string]time.Duration
}

// NewProcessTimeout creates a new ProcessTimeout store.
func NewProcessTimeout() *ProcessTimeout {
	return &ProcessTimeout{
		entries: make(map[string]time.Duration),
	}
}

// Set assigns a timeout duration to a named process.
// Returns an error if name is empty or duration is non-positive.
func (pt *ProcessTimeout) Set(name string, d time.Duration) error {
	if name == "" {
		return errors.New("process name must not be empty")
	}
	if d <= 0 {
		return errors.New("timeout duration must be positive")
	}
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.entries[name] = d
	return nil
}

// Get retrieves the timeout for a named process.
// Returns 0 and false if not set.
func (pt *ProcessTimeout) Get(name string) (time.Duration, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	d, ok := pt.entries[name]
	return d, ok
}

// Remove deletes the timeout entry for a named process.
func (pt *ProcessTimeout) Remove(name string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.entries, name)
}

// All returns a snapshot of all process timeout entries.
func (pt *ProcessTimeout) All() map[string]time.Duration {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	out := make(map[string]time.Duration, len(pt.entries))
	for k, v := range pt.entries {
		out[k] = v
	}
	return out
}

// Exceeded reports whether the given duration exceeds the stored timeout
// for the named process. Returns false if no timeout is configured.
func (pt *ProcessTimeout) Exceeded(name string, elapsed time.Duration) bool {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	d, ok := pt.entries[name]
	if !ok {
		return false
	}
	return elapsed >= d
}
