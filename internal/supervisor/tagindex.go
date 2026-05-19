package supervisor

import (
	"fmt"
	"sync"
)

// TagIndex maintains a mapping from tag strings to sets of process names,
// enabling efficient lookup of processes by tag.
type TagIndex struct {
	mu    sync.RWMutex
	index map[string]map[string]struct{} // tag -> set of process names
	reverse map[string]map[string]struct{} // process name -> set of tags
}

// NewTagIndex creates an empty TagIndex.
func NewTagIndex() *TagIndex {
	return &TagIndex{
		index:   make(map[string]map[string]struct{}),
		reverse: make(map[string]map[string]struct{}),
	}
}

// Set replaces all tags for the given process name.
func (t *TagIndex) Set(name string, tags []string) error {
	if name == "" {
		return fmt.Errorf("process name must not be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	// Remove old reverse mappings.
	if old, ok := t.reverse[name]; ok {
		for tag := range old {
			delete(t.index[tag], name)
			if len(t.index[tag]) == 0 {
				delete(t.index, tag)
			}
		}
	}

	if len(tags) == 0 {
		delete(t.reverse, name)
		return nil
	}

	tagSet := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
		if t.index[tag] == nil {
			t.index[tag] = make(map[string]struct{})
		}
		t.index[tag][name] = struct{}{}
	}
	t.reverse[name] = tagSet
	return nil
}

// Remove deletes all tag associations for the given process name.
func (t *TagIndex) Remove(name string) {
	_ = t.Set(name, nil)
}

// Lookup returns all process names that carry the given tag.
func (t *TagIndex) Lookup(tag string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	names := make([]string, 0, len(t.index[tag]))
	for name := range t.index[tag] {
		names = append(names, name)
	}
	return names
}

// Tags returns all tags currently associated with the given process name.
func (t *TagIndex) Tags(name string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tags := make([]string, 0, len(t.reverse[name]))
	for tag := range t.reverse[name] {
		tags = append(tags, tag)
	}
	return tags
}

// AllTags returns every tag currently tracked by the index.
func (t *TagIndex) AllTags() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]string, 0, len(t.index))
	for tag := range t.index {
		out = append(out, tag)
	}
	return out
}
