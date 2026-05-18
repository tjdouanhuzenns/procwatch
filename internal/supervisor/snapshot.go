package supervisor

import (
	"sync"
	"time"
)

// ProcessSnapshot captures a point-in-time view of a process's status and metrics.
type ProcessSnapshot struct {
	Name      string            `json:"name"`
	State     string            `json:"state"`
	Restarts  int               `json:"restarts"`
	Uptime    float64           `json:"uptime_seconds"`
	Timestamp time.Time         `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// SnapshotStore holds the most recent snapshot per process and supports
// diffing to detect changes between snapshots.
type SnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string]ProcessSnapshot
}

// NewSnapshotStore creates an empty SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string]ProcessSnapshot),
	}
}

// Save stores the latest snapshot for a process, returning true if the
// snapshot differs from the previously stored one (or is new).
func (s *SnapshotStore) Save(snap ProcessSnapshot) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, exists := s.snapshots[snap.Name]
	changed := !exists || prev.State != snap.State || prev.Restarts != snap.Restarts

	snap.Timestamp = time.Now()
	s.snapshots[snap.Name] = snap
	return changed
}

// Get returns the latest snapshot for the named process and whether it exists.
func (s *SnapshotStore) Get(name string) (ProcessSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.snapshots[name]
	return snap, ok
}

// All returns a copy of all stored snapshots.
func (s *SnapshotStore) All() []ProcessSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ProcessSnapshot, 0, len(s.snapshots))
	for _, snap := range s.snapshots {
		out = append(out, snap)
	}
	return out
}

// Remove deletes the snapshot for the named process.
func (s *SnapshotStore) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.snapshots, name)
}
