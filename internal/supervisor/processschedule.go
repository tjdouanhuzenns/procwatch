package supervisor

import (
	"errors"
	"sync"
	"time"
)

// Schedule holds a cron-style schedule string and optional window constraints.
type Schedule struct {
	Process  string
	Cron     string
	Enabled  bool
	LastRun  time.Time
	NextRun  time.Time
	UpdatedAt time.Time
}

// ProcessSchedule stores per-process schedule entries.
type ProcessSchedule struct {
	mu        sync.RWMutex
	schedules map[string]*Schedule
}

// NewProcessSchedule returns an initialised ProcessSchedule.
func NewProcessSchedule() *ProcessSchedule {
	return &ProcessSchedule{
		schedules: make(map[string]*Schedule),
	}
}

// Set registers or replaces a schedule for a process.
func (ps *ProcessSchedule) Set(process, cron string, nextRun time.Time) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if cron == "" {
		return errors.New("cron expression must not be empty")
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.schedules[process] = &Schedule{
		Process:   process,
		Cron:      cron,
		Enabled:   true,
		NextRun:   nextRun,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get returns the schedule for a process.
func (ps *ProcessSchedule) Get(process string) (*Schedule, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	s, ok := ps.schedules[process]
	if !ok {
		return nil, false
	}
	copy := *s
	return &copy, true
}

// RecordRun updates LastRun and advances NextRun for a process.
func (ps *ProcessSchedule) RecordRun(process string, next time.Time) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	s, ok := ps.schedules[process]
	if !ok {
		return errors.New("no schedule found for process: " + process)
	}
	s.LastRun = time.Now()
	s.NextRun = next
	s.UpdatedAt = time.Now()
	return nil
}

// SetEnabled enables or disables a schedule without removing it.
func (ps *ProcessSchedule) SetEnabled(process string, enabled bool) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	s, ok := ps.schedules[process]
	if !ok {
		return errors.New("no schedule found for process: " + process)
	}
	s.Enabled = enabled
	s.UpdatedAt = time.Now()
	return nil
}

// Remove deletes the schedule for a process.
func (ps *ProcessSchedule) Remove(process string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.schedules, process)
}

// All returns a snapshot of all schedules.
func (ps *ProcessSchedule) All() []Schedule {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make([]Schedule, 0, len(ps.schedules))
	for _, s := range ps.schedules {
		out = append(out, *s)
	}
	return out
}
