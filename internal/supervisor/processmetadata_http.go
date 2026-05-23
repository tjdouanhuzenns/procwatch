package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessMetadataRoutes attaches metadata CRUD endpoints to mux.
//
//	GET    /metadata/{process}          — list all metadata for a process
//	PUT    /metadata/{process}/{key}     — set a metadata value
//	DELETE /metadata/{process}/{key}    — delete a metadata entry
func RegisterProcessMetadataRoutes(mux *http.ServeMux, store *ProcessMetadataStore) {
	mux.HandleFunc("/metadata/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/metadata/"), "/", 2)
		process := parts[0]
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}

		if len(parts) == 1 {
			// List all metadata for process
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			entries := store.ForProcess(process)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
			return
		}

		key := parts[1]
		switch r.Method {
		case http.MethodPut:
			var body struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if err := store.Set(process, key, body.Value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodGet:
			entry, ok := store.Get(process, key)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entry)

		case http.MethodDelete:
			store.Delete(process, key)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
