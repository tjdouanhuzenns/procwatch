package supervisor

import (
	"fmt"
	"sync"
)

// SignalAction defines what action to take when a signal is routed to a process.
type SignalAction struct {
	ProcessName string
	Signal      string
}

// SignalRouter maps named signals to one or more process actions.
type SignalRouter struct {
	mu      sync.RWMutex
	routes  map[string][]SignalAction
}

// NewSignalRouter creates an empty SignalRouter.
func NewSignalRouter() *SignalRouter {
	return &SignalRouter{
		routes: make(map[string][]SignalAction),
	}
}

// Register associates a signal name with a process action.
// Returns an error if signal or process name is empty.
func (r *SignalRouter) Register(signal, processName, action string) error {
	if signal == "" {
		return fmt.Errorf("signal name must not be empty")
	}
	if processName == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if action == "" {
		return fmt.Errorf("action must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[signal] = append(r.routes[signal], SignalAction{
		ProcessName: processName,
		Signal:      action,
	})
	return nil
}

// Lookup returns all actions registered for the given signal.
// Returns nil if no routes exist for the signal.
func (r *SignalRouter) Lookup(signal string) []SignalAction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actions, ok := r.routes[signal]
	if !ok {
		return nil
	}
	copy := make([]SignalAction, len(actions))
	copy_ := copy
	for i, a := range actions {
		copy_[i] = a
	}
	return copy_
}

// Remove deletes all routes for the given signal.
func (r *SignalRouter) Remove(signal string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.routes, signal)
}

// All returns a snapshot of all registered routes.
func (r *SignalRouter) All() map[string][]SignalAction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string][]SignalAction, len(r.routes))
	for sig, actions := range r.routes {
		cp := make([]SignalAction, len(actions))
		for i, a := range actions {
			cp[i] = a
		}
		result[sig] = cp
	}
	return result
}
