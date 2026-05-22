package supervisor

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// notifyRecord is stored in-memory for the HTTP endpoint.
type notifyRecord struct {
	Process   string    `json:"process"`
	State     string    `json:"state"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NotifierLog is an in-memory log of recent notification events exposed via HTTP.
type NotifierLog struct {
	mu      sync.RWMutex
	events  []notifyRecord
	maxSize int
}

// NewNotifierLog creates a NotifierLog that retains up to maxSize events.
func NewNotifierLog(maxSize int) *NotifierLog {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &NotifierLog{maxSize: maxSize}
}

// Record appends an event to the log, evicting the oldest if at capacity.
func (l *NotifierLog) Record(e NotifyEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, notifyRecord{
		Process:   e.Process,
		State:     e.State,
		Message:   e.Message,
		Timestamp: e.Timestamp,
	})
}

// All returns a copy of all recorded events.
func (l *NotifierLog) All() []notifyRecord {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]notifyRecord, len(l.events))
	copy(out, l.events)
	return out
}

// RegisterNotifierRoutes mounts the notification log endpoint onto mux.
func RegisterNotifierRoutes(mux *http.ServeMux, log *NotifierLog) {
	mux.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(log.All())
	})
}
