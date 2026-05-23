package supervisor

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

// RegisterProcessGroupRoutes mounts process-group management endpoints onto mux.
//
//	GET  /groups              — list all group names
//	GET  /groups/{group}      — list members of a group
//	POST /groups/{group}/{process} — add process to group
//	DELETE /groups/{group}/{process} — remove process from group
func RegisterProcessGroupRoutes(mux *http.ServeMux, pg *ProcessGroup) {
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		groups := pg.Groups()
		sort.Strings(groups)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"groups": groups})
	})

	mux.HandleFunc("/groups/", func(w http.ResponseWriter, r *http.Request) {
		// path: /groups/{group} or /groups/{group}/{process}
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/groups/"), "/", 2)
		group := parts[0]
		if group == "" {
			http.Error(w, "group name required", http.StatusBadRequest)
			return
		}

		if len(parts) == 1 {
			// /groups/{group}
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			members, err := pg.Members(group)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			sort.Strings(members)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string][]string{"members": members})
			return
		}

		// /groups/{group}/{process}
		process := parts[1]
		switch r.Method {
		case http.MethodPost:
			if err := pg.Add(group, process); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			if err := pg.Remove(group, process); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
