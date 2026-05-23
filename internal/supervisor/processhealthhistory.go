package supervisor

import (
	"errors"
	"sync"
	"time"
)

// HealthEvent records a single health check outcome for a process.
type HealthEvent struct {
	Process   string
	Healthy   bool
	Message   string
	Timestamp time.Time
}

// ProcessHealthHistory stores a bounded ring of health events per process.
type ProcessHealthHistory struct {
	mu      sync.RWMutex
	events  map[string][]HealthEvent
	maxSize int
}

// NewProcessHealthHistory creates a new ProcessHealthHistory with the given max events per process.
func NewProcessHealthHistory(maxSize int) *ProcessHealthHistory {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &ProcessHealthHistory{
		events:  make(map[string][]HealthEvent),
		maxSize: maxSize,
	}
}

// Record appends a health event for the named process.
func (h *ProcessHealthHistory) Record(process string, healthy bool, message string) error {
	if process == "" {
		return errors.New("process name must not be empty")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	ev := HealthEvent{
		Process:   process,
		Healthy:   healthy,
		Message:   message,
		Timestamp: time.Now(),
	}
	buf := h.events[process]
	buf = append(buf, ev)
	if len(buf) > h.maxSize {
		buf = buf[len(buf)-h.maxSize:]
	}
	h.events[process] = buf
	return nil
}

// ForProcess returns a copy of the health history for the named process.
func (h *ProcessHealthHistory) ForProcess(process string) []HealthEvent {
	h.mu.RLock()
	defer h.mu.RUnlock()
	src := h.events[process]
	out := make([]HealthEvent, len(src))
	copy(out, src)
	return out
}

// Clear removes all recorded events for a process.
func (h *ProcessHealthHistory) Clear(process string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.events, process)
}

// Len returns the number of recorded events for a process.
func (h *ProcessHealthHistory) Len(process string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events[process])
}
