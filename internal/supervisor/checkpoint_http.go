package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterCheckpointRoutes mounts checkpoint CRUD endpoints onto mux.
func RegisterCheckpointRoutes(mux *http.ServeMux, cs *CheckpointStore) {
	mux.HandleFunc("/checkpoints", func(w http.ResponseWriter, r *http.Request) {
		process := r.URL.Query().Get("process")
		if process == "" {
			http.Error(w, "process query param required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			cps := cs.ForProcess(process)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cps)
		case http.MethodDelete:
			cs.Clear(process)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/checkpoints/", func(w http.ResponseWriter, r *http.Request) {
		// expects /checkpoints/{process}/{name}
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/checkpoints/"), "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.Error(w, "path must be /checkpoints/{process}/{name}", http.StatusBadRequest)
			return
		}
		process, name := parts[0], parts[1]
		switch r.Method {
		case http.MethodGet:
			cp, ok := cs.Get(process, name)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cp)
		case http.MethodPut:
			var meta map[string]string
			if err := json.NewDecoder(r.Body).Decode(&meta); err != nil && err.Error() != "EOF" {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := cs.Save(process, name, meta); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			cs.Delete(process, name)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
