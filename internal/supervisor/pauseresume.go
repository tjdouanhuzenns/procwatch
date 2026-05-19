package supervisor

import (
	"fmt"
	"sync"
)

// PauseState represents whether a process is paused or active.
type PauseState int

const (
	PauseStateActive PauseState = iota
	PauseStatePaused
)

func (s PauseState) String() string {
	switch s {
	case PauseStateActive:
		return "active"
	case PauseStatePaused:
		return "paused"
	default:
		return "unknown"
	}
}

// PauseResumeController tracks the pause/resume state for supervised processes.
type PauseResumeController struct {
	mu     sync.RWMutex
	states map[string]PauseState
}

// NewPauseResumeController creates a new PauseResumeController.
func NewPauseResumeController() *PauseResumeController {
	return &PauseResumeController{
		states: make(map[string]PauseState),
	}
}

// Pause marks the named process as paused, preventing restarts.
// Returns an error if the process is already paused.
func (c *PauseResumeController) Pause(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.states[name] == PauseStatePaused {
		return fmt.Errorf("process %q is already paused", name)
	}
	c.states[name] = PauseStatePaused
	return nil
}

// Resume marks the named process as active, allowing restarts.
// Returns an error if the process is not currently paused.
func (c *PauseResumeController) Resume(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.states[name] != PauseStatePaused {
		return fmt.Errorf("process %q is not paused", name)
	}
	c.states[name] = PauseStateActive
	return nil
}

// IsPaused reports whether the named process is currently paused.
func (c *PauseResumeController) IsPaused(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.states[name] == PauseStatePaused
}

// State returns the current PauseState for the named process.
func (c *PauseResumeController) State(name string) PauseState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.states[name]
}

// All returns a snapshot of all tracked pause states.
func (c *PauseResumeController) All() map[string]PauseState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]PauseState, len(c.states))
	for k, v := range c.states {
		out[k] = v
	}
	return out
}

// Remove deletes the tracked state for the named process.
func (c *PauseResumeController) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.states, name)
}
