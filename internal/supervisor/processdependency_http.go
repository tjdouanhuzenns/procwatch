package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterDependencyGraphRoutes mounts HTTP handlers for the dependency graph onto mux.
func RegisterDependencyGraphRoutes(mux *http.ServeMux, g *DependencyGraph) {
	mux.HandleFunc("/dependencies/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dependencies/")
		if name == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			deps, err := g.Deps(name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string][]string{"deps": deps})
		case http.MethodPut:
			var body struct {
				Deps []string `json:"deps"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := g.Add(name, body.Deps); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			if err := g.Remove(name); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/dependencies/ready", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Running []string `json:"running"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			body.Running = nil
		}
		ready := g.ReadyToStart(body.Running)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"ready": ready})
	})
}
