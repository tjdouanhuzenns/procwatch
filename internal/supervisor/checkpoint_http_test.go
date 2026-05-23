package supervisor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCheckpointMux() (*http.ServeMux, *CheckpointStore) {
	cs := NewCheckpointStore()
	mux := http.NewServeMux()
	RegisterCheckpointRoutes(mux, cs)
	return mux, cs
}

func TestCheckpointHTTP_PutAndGet(t *testing.T) {
	mux, _ := newCheckpointMux()
	body, _ := json.Marshal(map[string]string{"step": "init"})
	req := httptest.NewRequest(http.MethodPut, "/checkpoints/svc/boot", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/checkpoints/svc/boot", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var cp Checkpoint
	_ = json.NewDecoder(w2.Body).Decode(&cp)
	if cp.Metadata["step"] != "init" {
		t.Errorf("expected step=init, got %s", cp.Metadata["step"])
	}
}

func TestCheckpointHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newCheckpointMux()
	req := httptest.NewRequest(http.MethodGet, "/checkpoints/svc/missing", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCheckpointHTTP_ListForProcess(t *testing.T) {
	mux, cs := newCheckpointMux()
	_ = cs.Save("svc", "boot", nil)
	_ = cs.Save("svc", "ready", nil)
	req := httptest.NewRequest(http.MethodGet, "/checkpoints?process=svc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var cps []Checkpoint
	_ = json.NewDecoder(w.Body).Decode(&cps)
	if len(cps) != 2 {
		t.Errorf("expected 2 checkpoints, got %d", len(cps))
	}
}

func TestCheckpointHTTP_DeleteSpecific(t *testing.T) {
	mux, cs := newCheckpointMux()
	_ = cs.Save("svc", "boot", nil)
	req := httptest.NewRequest(http.MethodDelete, "/checkpoints/svc/boot", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	_, ok := cs.Get("svc", "boot")
	if ok {
		t.Error("expected checkpoint to be deleted")
	}
}

func TestCheckpointHTTP_DeleteMissingReturns404(t *testing.T) {
	mux, _ := newCheckpointMux()
	req := httptest.NewRequest(http.MethodDelete, "/checkpoints/svc/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCheckpointHTTP_MissingProcessQueryParam(t *testing.T) {
	mux, _ := newCheckpointMux()
	req := httptest.NewRequest(http.MethodGet, "/checkpoints", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
