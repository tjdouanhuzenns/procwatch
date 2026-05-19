package supervisor

import (
	"fmt"
	"sync"
)

// LabelFilter stores label sets per process and supports querying by label selectors.
type LabelFilter struct {
	mu     sync.RWMutex
	labels map[string]map[string]string // process name -> labels
}

// NewLabelFilter creates a new LabelFilter.
func NewLabelFilter() *LabelFilter {
	return &LabelFilter{
		labels: make(map[string]map[string]string),
	}
}

// Set assigns labels to a process, replacing any existing labels.
func (lf *LabelFilter) Set(process string, labels map[string]string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	lf.mu.Lock()
	defer lf.mu.Unlock()
	lf.labels[process] = copy
	return nil
}

// Get returns the labels for a process. Returns nil if not found.
func (lf *LabelFilter) Get(process string) map[string]string {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	raw, ok := lf.labels[process]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(raw))
	for k, v := range raw {
		copy[k] = v
	}
	return copy
}

// Remove deletes the label set for a process.
func (lf *LabelFilter) Remove(process string) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	delete(lf.labels, process)
}

// Match returns the names of all processes whose labels contain all key-value
// pairs in the selector. An empty selector matches all registered processes.
func (lf *LabelFilter) Match(selector map[string]string) []string {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	var result []string
	for name, lbls := range lf.labels {
		if matchesAll(lbls, selector) {
			result = append(result, name)
		}
	}
	return result
}

func matchesAll(labels, selector map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
