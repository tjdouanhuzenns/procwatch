package process

import (
	"fmt"
	"sync"
	"time"
)

// State represents the lifecycle state of a managed process.
type State int

const (
	StateIdle State = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
	StateFailed
	StateBackoff
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateFailed:
		return "failed"
	case StateBackoff:
		return "backoff"
	default:
		return "unknown"
	}
}

// StateChange records a transition between process states.
type StateChange struct {
	From      State
	To        State
	Timestamp time.Time
	Reason    string
}

// LifecycleTracker tracks state transitions for a named process.
type LifecycleTracker struct {
	mu      sync.RWMutex
	name    string
	current State
	history []StateChange
}

// NewLifecycleTracker creates a new tracker for the given process name.
func NewLifecycleTracker(name string) *LifecycleTracker {
	return &LifecycleTracker{
		name:    name,
		current: StateIdle,
		history: make([]StateChange, 0, 16),
	}
}

// Transition moves the tracker to a new state, recording the change.
// Returns an error if the transition is not permitted.
func (lt *LifecycleTracker) Transition(to State, reason string) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if !isValidTransition(lt.current, to) {
		return fmt.Errorf("process %q: invalid state transition %s -> %s", lt.name, lt.current, to)
	}

	lt.history = append(lt.history, StateChange{
		From:      lt.current,
		To:        to,
		Timestamp: time.Now(),
		Reason:    reason,
	})
	lt.current = to
	return nil
}

// Current returns the current state.
func (lt *LifecycleTracker) Current() State {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.current
}

// History returns a copy of all recorded state changes.
func (lt *LifecycleTracker) History() []StateChange {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	out := make([]StateChange, len(lt.history))
	copy(out, lt.history)
	return out
}

// isValidTransition defines the allowed state machine edges.
func isValidTransition(from, to State) bool {
	allowed := map[State][]State{
		StateIdle:     {StateStarting},
		StateStarting: {StateRunning, StateFailed, StateStopping},
		StateRunning:  {StateStopping, StateFailed, StateStopped},
		StateStopping: {StateStopped, StateFailed},
		StateStopped:  {StateStarting, StateBackoff},
		StateFailed:   {StateStarting, StateBackoff},
		StateBackoff:  {StateStarting, StateStopped},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}
