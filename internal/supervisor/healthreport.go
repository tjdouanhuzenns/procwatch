package supervisor

import (
	"sync"
	"time"
)

// HealthStatus represents the health state of a process.
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthDegraded
	HealthUnhealthy
)

func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "healthy"
	case HealthDegraded:
		return "degraded"
	case HealthUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthReport holds the health state and metadata for a single process.
type HealthReport struct {
	Process     string
	Status      HealthStatus
	Consecutive int
	LastChecked time.Time
	Message     string
}

// HealthReporter tracks per-process health reports.
type HealthReporter struct {
	mu      sync.RWMutex
	reports map[string]*HealthReport
}

// NewHealthReporter creates a new HealthReporter.
func NewHealthReporter() *HealthReporter {
	return &HealthReporter{
		reports: make(map[string]*HealthReport),
	}
}

// Record updates the health status for a named process.
func (r *HealthReporter) Record(name string, status HealthStatus, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.reports[name]
	if !ok {
		existing = &HealthReport{Process: name}
		r.reports[name] = existing
	}

	if existing.Status == status {
		existing.Consecutive++
	} else {
		existing.Consecutive = 1
	}
	existing.Status = status
	existing.LastChecked = time.Now()
	existing.Message = msg
}

// Get returns the health report for a process, and whether it exists.
func (r *HealthReporter) Get(name string) (HealthReport, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rep, ok := r.reports[name]
	if !ok {
		return HealthReport{}, false
	}
	return *rep, true
}

// All returns a snapshot of all current health reports.
func (r *HealthReporter) All() []HealthReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]HealthReport, 0, len(r.reports))
	for _, rep := range r.reports {
		out = append(out, *rep)
	}
	return out
}

// Remove deletes the health report for a process.
func (r *HealthReporter) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.reports, name)
}
