package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessTracingRoutes mounts tracing HTTP endpoints onto mux.
//
//	GET /tracing          — all trace entries
//	GET /tracing/{process} — entries for a specific process
func RegisterProcessTracingRoutes(mux *http.ServeMux, store *ProcessTracingStore) {
	mux.HandleFunc("/tracing", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/tracing/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		process := strings.TrimPrefix(r.URL.Path, "/tracing/")
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		entries := store.ForProcess(process)
		if len(entries) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
}
