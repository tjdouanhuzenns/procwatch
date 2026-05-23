package supervisor

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessEnvStore manages per-process environment variable snapshots.
type ProcessEnvStore struct {
	mu   sync.RWMutex
	envs map[string]map[string]string
}

// NewProcessEnvStore creates a new ProcessEnvStore.
func NewProcessEnvStore() *ProcessEnvStore {
	return &ProcessEnvStore{
		envs: make(map[string]map[string]string),
	}
}

// Set stores the environment variables for a named process.
func (s *ProcessEnvStore) Set(process string, env map[string]string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if len(env) == 0 {
		return errors.New("env map must not be empty")
	}
	for k := range env {
		if k == "" {
			return errors.New("env key must not be empty")
		}
	}
	copy := make(map[string]string, len(env))
	for k, v := range env {
		copy[k] = v
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.envs[process] = copy
	return nil
}

// Get returns the environment variables for a named process.
func (s *ProcessEnvStore) Get(process string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	env, ok := s.envs[process]
	if !ok {
		return nil, fmt.Errorf("no env for process %q", process)
	}
	copy := make(map[string]string, len(env))
	for k, v := range env {
		copy[k] = v
	}
	return copy, nil
}

// Remove deletes the environment snapshot for a named process.
func (s *ProcessEnvStore) Remove(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.envs, process)
}

// All returns a copy of all stored process environments.
func (s *ProcessEnvStore) All() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.envs))
	for proc, env := range s.envs {
		copy := make(map[string]string, len(env))
		for k, v := range env {
			copy[k] = v
		}
		out[proc] = copy
	}
	return out
}
