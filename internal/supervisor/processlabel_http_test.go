package supervisor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newLabelMux() (*http.ServeMux, *ProcessLabelStore) {
	mux := http.NewServeMux()
	store := NewProcessLabelStore()
	RegisterProcessLabelRoutes(mux, store)
	return mux, store
}

func TestProcessLabelHTTP_PutAndGet(t *testing.T) {
	mux, _ := newLabelMux()

	req := httptest.NewRequest(http.MethodPut, "/labels/web/env",
		strings.NewReader(`{"value":"prod"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/labels/web/env", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "prod") {
		t.Fatalf("expected 'prod' in body, got %s", w2.Body.String())
	}
}

func TestProcessLabelHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newLabelMux()
	req := httptest.NewRequest(http.MethodGet, "/labels/ghost/env", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProcessLabelHTTP_ListAll(t *testing.T) {
	mux, store := newLabelMux()
	_ = store.Set("api", "region", "eu")
	_ = store.Set("api", "tier", "core")

	req := httptest.NewRequest(http.MethodGet, "/labels/api", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "eu") || !strings.Contains(body, "core") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProcessLabelHTTP_Delete(t *testing.T) {
	mux, store := newLabelMux()
	_ = store.Set("svc", "color", "blue")

	req := httptest.NewRequest(http.MethodDelete, "/labels/svc/color", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	_, ok := store.Get("svc", "color")
	if ok {
		t.Fatal("expected label deleted")
	}
}
