package supervisor

import (
	"encoding/json"
	"net/http"
	"time"
)

// HTTPReporter serves process status, metrics, and health over HTTP.
type HTTPReporter struct {
	status  *StatusReporter
	metrics *MetricsCollector
	health  *HealthReporter
	server  *http.Server
}

// NewHTTPReporter creates an HTTPReporter bound to the given address.
func NewHTTPReporter(addr string, status *StatusReporter, metrics *MetricsCollector, health *HealthReporter) *HTTPReporter {
	r := &HTTPReporter{
		status:  status,
		metrics: metrics,
		health:  health,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/status", r.handleStatus)
	mux.HandleFunc("/metrics", r.handleMetrics)
	mux.HandleFunc("/health", r.handleHealth)
	r.server = &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
	}
	return r
}

func (r *HTTPReporter) ListenAndServe() error {
	return r.server.ListenAndServe()
}

func (r *HTTPReporter) Close() error {
	return r.server.Close()
}

func (r *HTTPReporter) handleStatus(w http.ResponseWriter, _ *http.Request) {
	all := r.status.All()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}

func (r *HTTPReporter) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	all := r.metrics.All()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}

func (r *HTTPReporter) handleHealth(w http.ResponseWriter, _ *http.Request) {
	all := r.health.All()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}
