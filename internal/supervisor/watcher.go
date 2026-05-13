package supervisor

import (
	"context"
	"sync"
	"time"
)

// WatchEvent represents a change in process state observed by the watcher.
type WatchEvent struct {
	ProcessName string
	OldStatus   string
	NewStatus   string
	At          time.Time
}

// Watcher polls the StatusReporter and emits events when process statuses change.
type Watcher struct {
	reporter *StatusReporter
	interval time.Duration
	events   chan WatchEvent
	mu       sync.Mutex
	last     map[string]string
}

// NewWatcher creates a Watcher that polls at the given interval.
func NewWatcher(reporter *StatusReporter, interval time.Duration) *Watcher {
	return &Watcher{
		reporter: reporter,
		interval: interval,
		events:   make(chan WatchEvent, 32),
		last:     make(map[string]string),
	}
}

// Events returns the read-only channel of WatchEvents.
func (w *Watcher) Events() <-chan WatchEvent {
	return w.events
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(w.events)
			return
		case <-ticker.C:
			w.poll()
		}
	}
}

func (w *Watcher) poll() {
	all := w.reporter.All()
	w.mu.Lock()
	defer w.mu.Unlock()
	for name, st := range all {
		prev, seen := w.last[name]
		if !seen || prev != st.State {
			ev := WatchEvent{
				ProcessName: name,
				OldStatus:   prev,
				NewStatus:   st.State,
				At:          time.Now(),
			}
			select {
			case w.events <- ev:
			default:
			}
			w.last[name] = st.State
		}
	}
}
