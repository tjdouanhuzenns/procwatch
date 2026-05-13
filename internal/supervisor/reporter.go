package supervisor

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/procwatch/internal/logger"
)

// HTTPReporter exposes process status and metrics over HTTP.
type HTTPReporter struct {
	server  *http.Server
	status  *StatusReporter
	metrics *MetricsCollector
	log     *logger.Logger
}

// NewHTTPReporter creates an HTTPReporter bound to addr.
func NewHTTPReporter(addr string, sr *StatusReporter, mc *MetricsCollector, log *logger.Logger) *HTTPReporter {
	r := &HTTPReporter{status: sr, metrics: mc, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("/status", r.handleStatus)
	mux.HandleFunc("/metrics", r.handleMetrics)
	r.server = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return r
}

// Start begins serving HTTP in a goroutine. It returns when the server is ready or ctx is done.
func (r *HTTPReporter) Start(ctx context.Context) {
	go func() {
		r.log.Info("http reporter listening", "addr", r.server.Addr)
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.log.Error("http reporter error", "err", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = r.server.Shutdown(shutCtx)
	}()
}

func (r *HTTPReporter) handleStatus(w http.ResponseWriter, _ *http.Request) {
	all := r.status.All()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(all)
}

func (r *HTTPReporter) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	all := r.metrics.All()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(all)
}
