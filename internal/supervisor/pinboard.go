package supervisor

import (
	"fmt"
	"sync"
)

// Pinboard stores arbitrary key-value annotations for processes.
// Annotations are useful for attaching metadata such as deployment IDs,
// owner teams, or custom runbook URLs to a supervised process.
type Pinboard struct {
	mu    sync.RWMutex
	notes map[string]map[string]string // process name -> key -> value
}

// NewPinboard creates an empty Pinboard.
func NewPinboard() *Pinboard {
	return &Pinboard{
		notes: make(map[string]map[string]string),
	}
}

// Set stores a key-value annotation for the given process.
func (p *Pinboard) Set(process, key, value string) error {
	if process == "" {
		return fmt.Errorf("pinboard: process name must not be empty")
	}
	if key == "" {
		return fmt.Errorf("pinboard: annotation key must not be empty")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.notes[process]; !ok {
		p.notes[process] = make(map[string]string)
	}
	p.notes[process][key] = value
	return nil
}

// Get returns the annotation value for a process and key.
// The second return value is false if no annotation exists.
func (p *Pinboard) Get(process, key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if m, ok := p.notes[process]; ok {
		v, found := m[key]
		return v, found
	}
	return "", false
}

// All returns a copy of all annotations for a process.
// Returns nil if the process has no annotations.
func (p *Pinboard) All(process string) map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	m, ok := p.notes[process]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// Delete removes a single annotation key from a process.
func (p *Pinboard) Delete(process, key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if m, ok := p.notes[process]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(p.notes, process)
		}
	}
}

// Remove deletes all annotations for a process.
func (p *Pinboard) Remove(process string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.notes, process)
}
