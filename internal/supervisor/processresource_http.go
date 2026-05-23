package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessResourceRoutes mounts CRUD endpoints for process resource limits.
//
//	PUT    /resources/{name}  — set limits
//	GET    /resources/{name}  — get limits
//	DELETE /resources/{name}  — remove limits
//	GET    /resources         — list all
func RegisterProcessResourceRoutes(mux *http.ServeMux, store *ProcessResourceStore) {
	mux.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/resources/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/resources/")
		if name == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			l, ok := store.Get(name)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(l)
		case http.MethodPut:
			var limits ResourceLimits
			if err := json.NewDecoder(r.Body).Decode(&limits); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := store.Set(name, limits); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			store.Remove(name)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
