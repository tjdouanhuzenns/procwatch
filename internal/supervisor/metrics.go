package supervisor

import (
	"sync"
	"time"
)

// ProcessMetrics holds runtime counters for a single process.
type ProcessMetrics struct {
	Name        string        `json:"name"`
	Restarts    int           `json:"restarts"`
	LastStarted time.Time     `json:"last_started"`
	LastExited  time.Time     `json:"last_exited"`
	Uptime      time.Duration `json:"uptime_ns"`
	TotalUptime time.Duration `json:"total_uptime_ns"`
}

// MetricsCollector tracks runtime metrics for all supervised processes.
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*ProcessMetrics
}

// NewMetricsCollector creates a new MetricsCollector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*ProcessMetrics),
	}
}

// RecordStart records that a process has started.
func (m *MetricsCollector) RecordStart(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.getOrCreate(name)
	if !entry.LastStarted.IsZero() {
		entry.Restarts++
	}
	entry.LastStarted = time.Now()
}

// RecordExit records that a process has exited.
func (m *MetricsCollector) RecordExit(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.getOrCreate(name)
	now := time.Now()
	entry.LastExited = now
	if !entry.LastStarted.IsZero() {
		session := now.Sub(entry.LastStarted)
		entry.TotalUptime += session
	}
}

// Get returns a copy of metrics for the named process.
func (m *MetricsCollector) Get(name string) (ProcessMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.metrics[name]
	if !ok {
		return ProcessMetrics{}, false
	}
	copy := *entry
	if !entry.LastStarted.IsZero() && entry.LastExited.Before(entry.LastStarted) {
		copy.Uptime = time.Since(entry.LastStarted)
	}
	return copy, true
}

// All returns a snapshot of metrics for all processes.
func (m *MetricsCollector) All() []ProcessMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ProcessMetrics, 0, len(m.metrics))
	for _, entry := range m.metrics {
		copy := *entry
		if !entry.LastStarted.IsZero() && entry.LastExited.Before(entry.LastStarted) {
			copy.Uptime = time.Since(entry.LastStarted)
		}
		out = append(out, copy)
	}
	return out
}

func (m *MetricsCollector) getOrCreate(name string) *ProcessMetrics {
	if _, ok := m.metrics[name]; !ok {
		m.metrics[name] = &ProcessMetrics{Name: name}
	}
	return m.metrics[name]
}
