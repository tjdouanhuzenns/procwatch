package supervisor

import (
	"errors"
	"fmt"
	"sync"
)

// EnvOverride holds per-process environment variable overrides applied at
// restart time, layered on top of the base process environment.
type EnvOverride struct {
	mu        sync.RWMutex
	overrides map[string]map[string]string // process name → key → value
}

// NewEnvOverride returns an initialised EnvOverride store.
func NewEnvOverride() *EnvOverride {
	return &EnvOverride{
		overrides: make(map[string]map[string]string),
	}
}

// Set stores one or more key=value overrides for the named process.
// Calling Set replaces only the supplied keys; existing keys are preserved.
func (e *EnvOverride) Set(name string, vars map[string]string) error {
	if name == "" {
		return errors.New("envoverride: process name must not be empty")
	}
	if len(vars) == 0 {
		return errors.New("envoverride: vars must not be empty")
	}
	for k := range vars {
		if k == "" {
			return fmt.Errorf("envoverride: empty key is not allowed")
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.overrides[name] == nil {
		e.overrides[name] = make(map[string]string)
	}
	for k, v := range vars {
		e.overrides[name][k] = v
	}
	return nil
}

// Get returns a copy of all env overrides for the named process.
// If no overrides exist an empty map is returned.
func (e *EnvOverride) Get(name string) map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	src := e.overrides[name]
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// Remove deletes all env overrides for the named process.
func (e *EnvOverride) Remove(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.overrides, name)
}

// All returns a deep copy of every stored override keyed by process name.
func (e *EnvOverride) All() map[string]map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]map[string]string, len(e.overrides))
	for name, vars := range e.overrides {
		copy := make(map[string]string, len(vars))
		for k, v := range vars {
			copy[k] = v
		}
		out[name] = copy
	}
	return out
}
