package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterPinboardRoutes attaches Pinboard HTTP handlers to the given mux.
//
//	GET  /pinboard/{process}          — list all annotations
//	PUT  /pinboard/{process}/{key}     — set annotation (body: {"value":"..."})
//	DELETE /pinboard/{process}/{key}  — remove annotation
func RegisterPinboardRoutes(mux *http.ServeMux, pb *Pinboard) {
	mux.HandleFunc("/pinboard/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/pinboard/"), "/", 2)
		process := parts[0]
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}

		if len(parts) == 1 {
			// /pinboard/{process}
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			all := pb.All(process)
			if all == nil {
				all = map[string]string{}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(all)
			return
		}

		// /pinboard/{process}/{key}
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
			if err := pb.Set(process, key, body.Value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			pb.Delete(process, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
