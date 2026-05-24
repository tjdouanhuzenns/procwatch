package supervisor

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// RegisterProcessHeartbeatRoutes mounts heartbeat endpoints onto mux.
//
//	POST /heartbeat/{process}        — record a beat (optional JSON body: {"threshold_ms": 5000})
//	GET  /heartbeat/{process}        — return entry + alive status
//	DELETE /heartbeat/{process}      — remove heartbeat record
//	GET  /heartbeat                  — list all entries with alive status
func RegisterProcessHeartbeatRoutes(mux *http.ServeMux, hb *ProcessHeartbeat) {
	mux.HandleFunc("/heartbeat/", func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimPrefix(r.URL.Path, "/heartbeat/")
		if process == "" {
			http.Error(w, "process name required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var body struct {
				ThresholdMS int64 `json:"threshold_ms"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			threshold := time.Duration(body.ThresholdMS) * time.Millisecond
			if err := hb.Beat(process, threshold); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			e, ok := hb.Get(process)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"process":      e.Process,
				"last_beat":    e.LastBeat,
				"threshold_ms": e.Threshold.Milliseconds(),
				"alive":        hb.IsAlive(process),
			})
		case http.MethodDelete:
			hb.Remove(process)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		all := hb.All()
		type row struct {
			Process     string    `json:"process"`
			LastBeat    time.Time `json:"last_beat"`
			ThresholdMS int64     `json:"threshold_ms"`
			Alive       bool      `json:"alive"`
		}
		rows := make([]row, len(all))
		for i, e := range all {
			rows[i] = row{e.Process, e.LastBeat, e.Threshold.Milliseconds(), hb.IsAlive(e.Process)}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rows)
	})
}
