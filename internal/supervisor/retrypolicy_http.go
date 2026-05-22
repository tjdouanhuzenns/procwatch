package supervisor

import (
	"encoding/json"
	"net/http"
	"time"
)

// retryPolicyHandler exposes retry policy state over HTTP for the reporter.
type retryPolicyHandler struct {
	manager *RetryPolicyManager
}

type retryPolicyResponse struct {
	Name   string      `json:"name"`
	Policy RetryPolicy `json:"policy"`
}

type retryAllowResponse struct {
	Name    string `json:"name"`
	Allowed bool   `json:"allowed"`
}

// RegisterRetryPolicyRoutes mounts retry policy endpoints on the given mux.
func RegisterRetryPolicyRoutes(mux *http.ServeMux, m *RetryPolicyManager) {
	h := &retryPolicyHandler{manager: m}
	mux.HandleFunc("/retry/policy", h.handleGetPolicy)
	mux.HandleFunc("/retry/allow", h.handleAllow)
	mux.HandleFunc("/retry/reset", h.handleReset)
}

func (h *retryPolicyHandler) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}
	p := h.manager.GetPolicy(name)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(retryPolicyResponse{Name: name, Policy: p})
}

func (h *retryPolicyHandler) handleAllow(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}
	allowed := h.manager.Allow(name, time.Now())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(retryAllowResponse{Name: name, Allowed: allowed})
}

func (h *retryPolicyHandler) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}
	h.manager.Reset(name)
	w.WriteHeader(http.StatusNoContent)
}
