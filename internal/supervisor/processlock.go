package supervisor

import (
	"fmt"
	"sync"
)

// ProcessLock provides named mutual exclusion for process operations,
// preventing concurrent conflicting actions on the same process.
type ProcessLock struct {
	mu    sync.Mutex
	locks map[string]*entry
}

type entry struct {
	mu      sync.Mutex
	held    bool
	holder  string
	waiters int
}

// NewProcessLock creates a new ProcessLock.
func NewProcessLock() *ProcessLock {
	return &ProcessLock{
		locks: make(map[string]*entry),
	}
}

func (pl *ProcessLock) getEntry(process string) *entry {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	e, ok := pl.locks[process]
	if !ok {
		e = &entry{}
		pl.locks[process] = e
	}
	return e
}

// TryLock attempts to acquire the lock for a process on behalf of an operation.
// Returns an error if the lock is already held.
func (pl *ProcessLock) TryLock(process, operation string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if operation == "" {
		return fmt.Errorf("operation must not be empty")
	}
	e := pl.getEntry(process)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.held {
		return fmt.Errorf("process %q is locked by operation %q", process, e.holder)
	}
	e.held = true
	e.holder = operation
	return nil
}

// Unlock releases the lock for a process.
func (pl *ProcessLock) Unlock(process string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	e := pl.getEntry(process)
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.held {
		return fmt.Errorf("process %q is not locked", process)
	}
	e.held = false
	e.holder = ""
	return nil
}

// Holder returns the current operation holding the lock, or empty string if unlocked.
func (pl *ProcessLock) Holder(process string) string {
	e := pl.getEntry(process)
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.held {
		return ""
	}
	return e.holder
}

// IsLocked reports whether the given process is currently locked.
func (pl *ProcessLock) IsLocked(process string) bool {
	return pl.Holder(process) != ""
}
