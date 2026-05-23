package supervisor

import (
	"errors"
	"fmt"
	"sync"
)

// ProcessGroup manages named groups of processes, allowing bulk operations
// such as starting, stopping, or querying all members of a group.
type ProcessGroup struct {
	mu     sync.RWMutex
	groups map[string]map[string]struct{} // group -> set of process names
}

// NewProcessGroup returns an empty ProcessGroup.
func NewProcessGroup() *ProcessGroup {
	return &ProcessGroup{
		groups: make(map[string]map[string]struct{}),
	}
}

// Add adds a process to a group, creating the group if it does not exist.
func (pg *ProcessGroup) Add(group, process string) error {
	if group == "" {
		return errors.New("group name must not be empty")
	}
	if process == "" {
		return errors.New("process name must not be empty")
	}
	pg.mu.Lock()
	defer pg.mu.Unlock()
	if _, ok := pg.groups[group]; !ok {
		pg.groups[group] = make(map[string]struct{})
	}
	pg.groups[group][process] = struct{}{}
	return nil
}

// Remove removes a process from a group. If the group becomes empty it is deleted.
func (pg *ProcessGroup) Remove(group, process string) error {
	if group == "" {
		return errors.New("group name must not be empty")
	}
	if process == "" {
		return errors.New("process name must not be empty")
	}
	pg.mu.Lock()
	defer pg.mu.Unlock()
	members, ok := pg.groups[group]
	if !ok {
		return fmt.Errorf("group %q not found", group)
	}
	delete(members, process)
	if len(members) == 0 {
		delete(pg.groups, group)
	}
	return nil
}

// Members returns the process names belonging to group.
func (pg *ProcessGroup) Members(group string) ([]string, error) {
	if group == "" {
		return nil, errors.New("group name must not be empty")
	}
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	members, ok := pg.groups[group]
	if !ok {
		return nil, fmt.Errorf("group %q not found", group)
	}
	out := make([]string, 0, len(members))
	for p := range members {
		out = append(out, p)
	}
	return out, nil
}

// Groups returns all known group names.
func (pg *ProcessGroup) Groups() []string {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	out := make([]string, 0, len(pg.groups))
	for g := range pg.groups {
		out = append(out, g)
	}
	return out
}

// DeleteGroup removes an entire group and all its members.
func (pg *ProcessGroup) DeleteGroup(group string) error {
	if group == "" {
		return errors.New("group name must not be empty")
	}
	pg.mu.Lock()
	defer pg.mu.Unlock()
	if _, ok := pg.groups[group]; !ok {
		return fmt.Errorf("group %q not found", group)
	}
	delete(pg.groups, group)
	return nil
}
