package supervisor

import (
	"sync"
	"time"
)

// EventKind describes the type of supervisor event.
type EventKind string

const (
	EventStarted   EventKind = "started"
	EventStopped   EventKind = "stopped"
	EventRestarted EventKind = "restarted"
	EventFailed    EventKind = "failed"
	EventThrottled EventKind = "throttled"
)

// Event represents a single recorded supervisor event for a process.
type Event struct {
	Process   string
	Kind      EventKind
	Timestamp time.Time
	Message   string
}

// EventLog records a bounded, ordered history of supervisor events.
type EventLog struct {
	mu     sync.Mutex
	events []Event
	cap    int
}

// NewEventLog creates an EventLog that retains at most maxEvents entries.
// Older entries are evicted when the log is full.
func NewEventLog(maxEvents int) *EventLog {
	if maxEvents <= 0 {
		maxEvents = 256
	}
	return &EventLog{
		events: make([]Event, 0, maxEvents),
		cap:    maxEvents,
	}
}

// Record appends an event to the log, evicting the oldest entry if necessary.
func (el *EventLog) Record(process string, kind EventKind, message string) {
	el.mu.Lock()
	defer el.mu.Unlock()

	ev := Event{
		Process:   process,
		Kind:      kind,
		Timestamp: time.Now(),
		Message:   message,
	}

	if len(el.events) >= el.cap {
		// Evict oldest (index 0) by shifting.
		copy(el.events, el.events[1:])
		el.events[len(el.events)-1] = ev
	} else {
		el.events = append(el.events, ev)
	}
}

// All returns a snapshot of all recorded events in chronological order.
func (el *EventLog) All() []Event {
	el.mu.Lock()
	defer el.mu.Unlock()

	snap := make([]Event, len(el.events))
	copy(snap, el.events)
	return snap
}

// ForProcess returns events recorded for the given process name.
func (el *EventLog) ForProcess(name string) []Event {
	el.mu.Lock()
	defer el.mu.Unlock()

	var out []Event
	for _, ev := range el.events {
		if ev.Process == name {
			out = append(out, ev)
		}
	}
	return out
}

// Len returns the current number of recorded events.
func (el *EventLog) Len() int {
	el.mu.Lock()
	defer el.mu.Unlock()
	return len(el.events)
}
