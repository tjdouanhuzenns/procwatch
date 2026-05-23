package supervisor

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// RegisterProcessTimeoutRoutes mounts process timeout HTTP endpoints onto mux.
//
//	PUT  /timeouts/{name}?seconds=N  — set timeout
//	GET  /timeouts/{name}            — get timeout
//	DELETE /timeouts/{name}          — remove timeout
//	GET  /timeouts                   — list all
func RegisterProcessTimeoutRoutes(mux *http.ServeMux, pt *ProcessTimeout) {
	mux.HandleFunc("/timeouts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		all := pt.All()
		out := make(map[string]float64, len(all))
		for k, v := range all {
			out[k] = v.Seconds()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("/timeouts/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/timeouts/"):]
		if name == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			d, ok := pt.Get(name)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]float64{"seconds": d.Seconds()})

		case http.MethodPut:
			sec := r.URL.Query().Get("seconds")
			if sec == "" {
				http.Error(w, "seconds query param required", http.StatusBadRequest)
				return
			}
			n, err := strconv.ParseFloat(sec, 64)
			if err != nil || n <= 0 {
				http.Error(w, "seconds must be a positive number", http.StatusBadRequest)
				return
			}
			if err := pt.Set(name, time.Duration(n*float64(time.Second))); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			pt.Remove(name)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
