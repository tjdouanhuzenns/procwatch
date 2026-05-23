package supervisor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newDepMux(t *testing.T) (*http.ServeMux, *DependencyGraph) {
	t.Helper()
	g := NewDependencyGraph()
	mux := http.NewServeMux()
	RegisterDependencyGraphRoutes(mux, g)
	return mux, g
}

func TestDepHTTP_PutAndGet(t *testing.T) {
	mux, _ := newDepMux(t)

	body, _ := json.Marshal(map[string][]string{"deps": {"db", "cache"}})
	req := httptest.NewRequest(http.MethodPut, "/dependencies/api", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/dependencies/api", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp map[string][]string
	json.NewDecoder(rec2.Body).Decode(&resp)
	if len(resp["deps"]) != 2 {
		t.Fatalf("expected 2 deps, got %v", resp["deps"])
	}
}

func TestDepHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newDepMux(t)
	req := httptest.NewRequest(http.MethodGet, "/dependencies/ghost", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDepHTTP_Delete(t *testing.T) {
	mux, g := newDepMux(t)
	if err := g.Add("worker", []string{"db"}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/dependencies/worker", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDepHTTP_ReadyEndpoint(t *testing.T) {
	mux, g := newDepMux(t)
	if err := g.Add("api", []string{"db"}); err != nil {
		t.Fatal(err)
	}
	if err := g.Add("db", nil); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string][]string{"running": {"db"}})
	req := httptest.NewRequest(http.MethodGet, "/dependencies/ready", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string][]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp["ready"]) == 0 {
		t.Fatal("expected at least one ready process")
	}
}
