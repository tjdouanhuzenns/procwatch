package supervisor

import (
	"sync"
	"time"
)

// Checkpoint represents a named recovery point for a process.
type Checkpoint struct {
	Process   string
	Name      string
	Metadata  map[string]string
	CreatedAt time.Time
}

// CheckpointStore persists named checkpoints per process for recovery tracking.
type CheckpointStore struct {
	mu    sync.RWMutex
	store map[string]map[string]Checkpoint // process -> name -> checkpoint
}

// NewCheckpointStore returns an initialised CheckpointStore.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{
		store: make(map[string]map[string]Checkpoint),
	}
}

// Save records or overwrites a checkpoint for the given process.
func (c *CheckpointStore) Save(process, name string, meta map[string]string) error {
	if process == "" {
		return errorf("process name must not be empty")
	}
	if name == "" {
		return errorf("checkpoint name must not be empty")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[process]; !ok {
		c.store[process] = make(map[string]Checkpoint)
	}
	copy := make(map[string]string, len(meta))
	for k, v := range meta {
		copy[k] = v
	}
	c.store[process][name] = Checkpoint{
		Process:   process,
		Name:      name,
		Metadata:  copy,
		CreatedAt: time.Now(),
	}
	return nil
}

// Get returns the checkpoint for a process by name.
func (c *CheckpointStore) Get(process, name string) (Checkpoint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if byName, ok := c.store[process]; ok {
		cp, found := byName[name]
		return cp, found
	}
	return Checkpoint{}, false
}

// ForProcess returns all checkpoints for a given process.
func (c *CheckpointStore) ForProcess(process string) []Checkpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	byName, ok := c.store[process]
	if !ok {
		return nil
	}
	out := make([]Checkpoint, 0, len(byName))
	for _, cp := range byName {
		out = append(out, cp)
	}
	return out
}

// Delete removes a specific checkpoint for a process.
func (c *CheckpointStore) Delete(process, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if byName, ok := c.store[process]; ok {
		delete(byName, name)
		if len(byName) == 0 {
			delete(c.store, process)
		}
	}
}

// Clear removes all checkpoints for a process.
func (c *CheckpointStore) Clear(process string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, process)
}
