package supervisor

import (
	"errors"
	"fmt"
	"sync"
)

// DependencyGraph tracks inter-process start-order dependencies.
// A process may not start until all of its declared dependencies are running.
type DependencyGraph struct {
	mu   sync.RWMutex
	edges map[string][]string // process -> list of dependencies
}

// NewDependencyGraph returns an empty DependencyGraph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		edges: make(map[string][]string),
	}
}

// Add registers that process `name` depends on each entry in `deps`.
// Returns an error if `name` is empty or if adding the edges would
// introduce a cycle.
func (g *DependencyGraph) Add(name string, deps []string) error {
	if name == "" {
		return errors.New("dependency graph: process name must not be empty")
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	// Temporarily apply and check for cycles.
	prev := g.edges[name]
	g.edges[name] = deps
	if cycle := g.findCycle(); cycle != "" {
		g.edges[name] = prev
		return fmt.Errorf("dependency graph: cycle detected involving %q", cycle)
	}
	return nil
}

// Deps returns the direct dependencies of `name`.
func (g *DependencyGraph) Deps(name string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]string, len(g.edges[name]))
	copy(out, g.edges[name])
	return out
}

// ReadyToStart returns true when every declared dependency of `name`
// appears in the `running` set.
func (g *DependencyGraph) ReadyToStart(name string, running map[string]bool) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, dep := range g.edges[name] {
		if !running[dep] {
			return false
		}
	}
	return true
}

// Remove deletes all dependency information for `name`.
func (g *DependencyGraph) Remove(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.edges, name)
}

// findCycle performs a DFS and returns the name of any node involved in a
// cycle, or an empty string if the graph is acyclic. Must be called with
// g.mu held for writing.
func (g *DependencyGraph) findCycle() string {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	var dfs func(n string) string
	dfs = func(n string) string {
		visited[n] = true
		onStack[n] = true
		for _, dep := range g.edges[n] {
			if !visited[dep] {
				if hit := dfs(dep); hit != "" {
					return hit
				}
			} else if onStack[dep] {
				return dep
			}
		}
		onStack[n] = false
		return ""
	}
	for n := range g.edges {
		if !visited[n] {
			if hit := dfs(n); hit != "" {
				return hit
			}
		}
	}
	return ""
}
