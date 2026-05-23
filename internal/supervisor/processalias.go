package supervisor

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessAlias maps human-friendly alias names to canonical process names.
// This allows operators to refer to processes by short or alternate identifiers.
type ProcessAlias struct {
	mu      sync.RWMutex
	aliases map[string]string // alias -> canonical name
	reverse map[string][]string // canonical name -> aliases
}

// NewProcessAlias creates an empty ProcessAlias registry.
func NewProcessAlias() *ProcessAlias {
	return &ProcessAlias{
		aliases: make(map[string]string),
		reverse: make(map[string][]string),
	}
}

// Set registers an alias for a canonical process name.
// Returns an error if alias or name is empty, or if the alias is already in use.
func (p *ProcessAlias) Set(alias, name string) error {
	if alias == "" {
		return errors.New("alias must not be empty")
	}
	if name == "" {
		return errors.New("process name must not be empty")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if existing, ok := p.aliases[alias]; ok && existing != name {
		return fmt.Errorf("alias %q already mapped to %q", alias, existing)
	}
	p.aliases[alias] = name
	for _, a := range p.reverse[name] {
		if a == alias {
			return nil
		}
	}
	p.reverse[name] = append(p.reverse[name], alias)
	return nil
}

// Resolve returns the canonical process name for the given alias.
// Returns an error if the alias is not registered.
func (p *ProcessAlias) Resolve(alias string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	name, ok := p.aliases[alias]
	if !ok {
		return "", fmt.Errorf("alias %q not found", alias)
	}
	return name, nil
}

// AliasesFor returns all aliases registered for a canonical process name.
func (p *ProcessAlias) AliasesFor(name string) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]string, len(p.reverse[name]))
	copy(result, p.reverse[name])
	return result
}

// Remove deletes an alias mapping. If the alias does not exist, it is a no-op.
func (p *ProcessAlias) Remove(alias string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	name, ok := p.aliases[alias]
	if !ok {
		return
	}
	delete(p.aliases, alias)
	list := p.reverse[name]
	for i, a := range list {
		if a == alias {
			p.reverse[name] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

// All returns a copy of the full alias-to-name mapping.
func (p *ProcessAlias) All() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[string]string, len(p.aliases))
	for k, v := range p.aliases {
		out[k] = v
	}
	return out
}
