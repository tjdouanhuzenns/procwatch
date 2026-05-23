package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
)

// RegisterProcessNotesRoutes mounts the process-notes HTTP API onto mux.
//
//	GET  /notes              — list all notes
//	GET  /notes/{process}    — list notes for a process
//	POST /notes/{process}    — add a note  (body: {"note":"...","author":"..."})
//	DELETE /notes/{process}  — clear all notes for a process
func RegisterProcessNotesRoutes(mux *http.ServeMux, store *ProcessNotes) {
	mux.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.All())
	})

	mux.HandleFunc("/notes/", func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimPrefix(r.URL.Path, "/notes/")
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(store.ForProcess(process))

		case http.MethodPost:
			var body struct {
				Note   string `json:"note"`
				Author string `json:"author"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := store.Add(process, body.Note, body.Author); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			store.Clear(process)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
