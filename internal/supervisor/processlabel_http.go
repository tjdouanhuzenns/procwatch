package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessLabelRoutes mounts label CRUD endpoints onto mux.
//
//	PUT    /labels/{process}/{key}   – set a label value (body: {"value":"..."} )
//	GET    /labels/{process}/{key}   – get a single label
//	GET    /labels/{process}         – list all labels for a process
//	DELETE /labels/{process}/{key}   – remove a single label
func RegisterProcessLabelRoutes(mux *http.ServeMux, store *ProcessLabelStore) {
	mux.HandleFunc("/labels/", func(w http.ResponseWriter, r *http.Request) {
		// strip leading "/labels/"
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/labels/"), "/", 2)
		process := parts[0]
		key := ""
		if len(parts) == 2 {
			key = parts[1]
		}

		switch r.Method {
		case http.MethodPut:
			if key == "" {
				http.Error(w, "key required", http.StatusBadRequest)
				return
			}
			var body struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if err := store.Set(process, key, body.Value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodGet:
			if key == "" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(store.All(process))
				return
			}
			v, ok := store.Get(process, key)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"value": v})

		case http.MethodDelete:
			if key == "" {
				http.Error(w, "key required", http.StatusBadRequest)
				return
			}
			store.Remove(process, key)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
