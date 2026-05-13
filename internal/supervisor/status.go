package supervisor

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/user/procwatch/internal/process"
)

// ProcessStatus is a snapshot of a single process's runtime state.
type ProcessStatus struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Retries int    `json:"retries"`
	Uptime  string `json:"uptime,omitempty"`
}

// StatusReporter exposes process states over HTTP as JSON.
type StatusReporter struct {
	mu        sync.RWMutex
	trackers  map[string]*process.LifecycleTracker
	startedAt map[string]time.Time
}

// NewStatusReporter creates a reporter with no registered processes.
func NewStatusReporter() *StatusReporter {
	return &StatusReporter{
		trackers:  make(map[string]*process.LifecycleTracker),
		startedAt: make(map[string]time.Time),
	}
}

// Register adds a lifecycle tracker for the named process.
func (sr *StatusReporter) Register(name string, lt *process.LifecycleTracker) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.trackers[name] = lt
}

// MarkStarted records the time a process entered the running state.
func (sr *StatusReporter) MarkStarted(name string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.startedAt[name] = time.Now()
}

// Snapshot returns the current status of all registered processes.
func (sr *StatusReporter) Snapshot() []ProcessStatus {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	out := make([]ProcessStatus, 0, len(sr.trackers))
	for name, lt := range sr.trackers {
		ps := ProcessStatus{
			Name:  name,
			State: lt.Current().String(),
		}
		if t, ok := sr.startedAt[name]; ok && lt.Current() == process.StateRunning {
			ps.Uptime = time.Since(t).Round(time.Second).String()
		}
		h := lt.History()
		for _, ch := range h {
			if ch.To == process.StateStarting {
				ps.Retries++
			}
		}
		if ps.Retries > 0 {
			ps.Retries-- // first start is not a retry
		}
		out = append(out, ps)
	}
	return out
}

// ServeHTTP implements http.Handler and writes the JSON status response.
func (sr *StatusReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sr.Snapshot())
}
