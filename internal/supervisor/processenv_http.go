package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessEnvRoutes registers HTTP handlers for process environment management.
//
//	PUT    /env/{process}        — set env vars
//	GET    /env/{process}        — get env vars
//	DELETE /env/{process}        — remove env vars
//	GET    /env                  — list all
func RegisterProcessEnvRoutes(mux *http.ServeMux, store *ProcessEnvStore) {
	mux.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/env/", func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimPrefix(r.URL.Path, "/env/")
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			var env map[string]string
			if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := store.Set(process, env); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			env, err := store.Get(process)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(env)
		case http.MethodDelete:
			store.Remove(process)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
