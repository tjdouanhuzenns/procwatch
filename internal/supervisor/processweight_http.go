package supervisor

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// RegisterProcessWeightRoutes mounts CRUD endpoints for process weights.
//
//	PUT    /weights/{name}        — set weight (body: {"weight": N})
//	GET    /weights/{name}        — get weight
//	DELETE /weights/{name}        — remove weight
//	GET    /weights               — list all
func RegisterProcessWeightRoutes(mux *http.ServeMux, pw *ProcessWeight) {
	mux.HandleFunc("/weights", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pw.All())
	})

	mux.HandleFunc("/weights/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/weights/")
		if name == "" {
			http.Error(w, "missing process name", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			var body struct {
				Weight int `json:"weight"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := pw.Set(name, body.Weight); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			val, ok := pw.Get(name)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"weight": strconv.Itoa(val)})
		case http.MethodDelete:
			pw.Remove(name)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
