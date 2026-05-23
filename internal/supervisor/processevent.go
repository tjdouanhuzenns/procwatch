package supervisor

import (
	"errors"
	"sync"
	"time"
)

// ProcessEvent represents a discrete event emitted by a process lifecycle.
type ProcessEvent struct {
	Process   string
	Kind      string
	Message   string
	Timestamp time.Time
}

// ProcessEventLog stores a bounded, ordered log of process events.
type ProcessEventLog struct {
	mu      sync.Mutex
	events  []ProcessEvent
	maxSize int
}

// NewProcessEventLog creates a ProcessEventLog with the given max capacity.
func NewProcessEventLog(maxSize int) *ProcessEventLog {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &ProcessEventLog{maxSize: maxSize}
}

// Record appends an event to the log, evicting the oldest if at capacity.
func (l *ProcessEventLog) Record(process, kind, message string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	if kind == "" {
		return errors.New("event kind must not be empty")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, ProcessEvent{
		Process:   process,
		Kind:      kind,
		Message:   message,
		Timestamp: time.Now(),
	})
	return nil
}

// ForProcess returns all events recorded for the given process name.
func (l *ProcessEventLog) ForProcess(process string) []ProcessEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []ProcessEvent
	for _, e := range l.events {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// All returns a copy of all recorded events.
func (l *ProcessEventLog) All() []ProcessEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]ProcessEvent, len(l.events))
	copy(out, l.events)
	return out
}

// Len returns the current number of recorded events.
func (l *ProcessEventLog) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.events)
}
