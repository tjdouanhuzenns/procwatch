package supervisor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newEnvMux() (*http.ServeMux, *ProcessEnvStore) {
	mux := http.NewServeMux()
	store := NewProcessEnvStore()
	RegisterProcessEnvRoutes(mux, store)
	return mux, store
}

func TestProcessEnvHTTP_PutAndGet(t *testing.T) {
	mux, _ := newEnvMux()
	body, _ := json.Marshal(map[string]string{"PORT": "8080"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPut, "/env/web", bytes.NewReader(body)))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/env/web", nil))
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var got map[string]string
	_ = json.NewDecoder(rr2.Body).Decode(&got)
	if got["PORT"] != "8080" {
		t.Errorf("unexpected env: %v", got)
	}
}

func TestProcessEnvHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newEnvMux()
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/env/ghost", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestProcessEnvHTTP_Delete(t *testing.T) {
	mux, store := newEnvMux()
	_ = store.Set("web", map[string]string{"A": "1"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/env/web", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, err := store.Get("web"); err == nil {
		t.Fatal("expected env to be removed")
	}
}

func TestProcessEnvHTTP_ListAll(t *testing.T) {
	mux, store := newEnvMux()
	_ = store.Set("web", map[string]string{"A": "1"})
	_ = store.Set("worker", map[string]string{"B": "2"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/env", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var all map[string]map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&all)
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
