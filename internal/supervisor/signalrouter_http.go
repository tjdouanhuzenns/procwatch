package supervisor

import (
	"encoding/json"
	"net/http"
)

// RegisterSignalRouterRoutes registers HTTP endpoints for the SignalRouter.
//
//	GET  /signals         — list all routes
//	POST /signals/register — register a new route
//	DELETE /signals/{signal} — remove all routes for a signal
func RegisterSignalRouterRoutes(mux *http.ServeMux, r *SignalRouter) {
	mux.HandleFunc("/signals", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(r.All())
	})

	mux.HandleFunc("/signals/register", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Signal  string `json:"signal"`
			Process string `json:"process"`
			Action  string `json:"action"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := r.Register(body.Signal, body.Process, body.Action); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("/signals/remove", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Signal string `json:"signal"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		r.Remove(body.Signal)
		w.WriteHeader(http.StatusNoContent)
	})
}
