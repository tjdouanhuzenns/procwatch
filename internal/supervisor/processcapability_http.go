package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessCapabilityRoutes mounts process capability endpoints onto mux.
//
//	PUT    /capabilities/{process}        — set capabilities
//	GET    /capabilities/{process}        — get capabilities
//	DELETE /capabilities/{process}        — remove capabilities
//	GET    /capabilities                  — list all
func RegisterProcessCapabilityRoutes(mux *http.ServeMux, store *ProcessCapabilityStore) {
	mux.HandleFunc("/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/capabilities/", func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimPrefix(r.URL.Path, "/capabilities/")
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			caps, ok := store.Get(process)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(caps)
		case http.MethodPut:
			var caps []string
			if err := json.NewDecoder(r.Body).Decode(&caps); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			if err := store.Set(process, caps); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			store.Remove(process)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
