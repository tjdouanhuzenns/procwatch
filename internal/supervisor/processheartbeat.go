package supervisor

import (
	"errors"
	"sync"
	"time"
)

// HeartbeatEntry records the last heartbeat time and staleness threshold for a process.
type HeartbeatEntry struct {
	Process   string
	LastBeat  time.Time
	Threshold time.Duration
}

// ProcessHeartbeat tracks heartbeat timestamps for supervised processes.
type ProcessHeartbeat struct {
	mu      sync.RWMutex
	entries map[string]HeartbeatEntry
}

// NewProcessHeartbeat returns an initialised ProcessHeartbeat.
func NewProcessHeartbeat() *ProcessHeartbeat {
	return &ProcessHeartbeat{
		entries: make(map[string]HeartbeatEntry),
	}
}

// Beat records the current time as the last heartbeat for the named process.
// threshold is the maximum acceptable duration between beats before the process
// is considered stale. A non-positive threshold defaults to 30 seconds.
func (h *ProcessHeartbeat) Beat(process string, threshold time.Duration) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if threshold <= 0 {
		threshold = 30 * time.Second
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries[process] = HeartbeatEntry{
		Process:   process,
		LastBeat:  time.Now(),
		Threshold: threshold,
	}
	return nil
}

// IsAlive reports whether the named process has a heartbeat recorded within
// its configured threshold. Returns false if the process is unknown.
func (h *ProcessHeartbeat) IsAlive(process string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	e, ok := h.entries[process]
	if !ok {
		return false
	}
	return time.Since(e.LastBeat) <= e.Threshold
}

// Get returns the HeartbeatEntry for the named process and whether it exists.
func (h *ProcessHeartbeat) Get(process string) (HeartbeatEntry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	e, ok := h.entries[process]
	return e, ok
}

// Remove deletes the heartbeat record for the named process.
func (h *ProcessHeartbeat) Remove(process string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.entries, process)
}

// All returns a snapshot of all heartbeat entries.
func (h *ProcessHeartbeat) All() []HeartbeatEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HeartbeatEntry, 0, len(h.entries))
	for _, e := range h.entries {
		out = append(out, e)
	}
	return out
}
