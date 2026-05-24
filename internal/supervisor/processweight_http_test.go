package supervisor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newWeightMux() (*http.ServeMux, *ProcessWeight) {
	mux := http.NewServeMux()
	pw := NewProcessWeight()
	RegisterProcessWeightRoutes(mux, pw)
	return mux, pw
}

func TestProcessWeightHTTP_PutAndGet(t *testing.T) {
	mux, _ := newWeightMux()
	body, _ := json.Marshal(map[string]int{"weight": 42})
	req := httptest.NewRequest(http.MethodPut, "/weights/worker", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/weights/worker", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr2.Body).Decode(&resp)
	if resp["weight"] != "42" {
		t.Fatalf("expected weight 42, got %s", resp["weight"])
	}
}

func TestProcessWeightHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newWeightMux()
	req := httptest.NewRequest(http.MethodGet, "/weights/ghost", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestProcessWeightHTTP_Delete(t *testing.T) {
	mux, pw := newWeightMux()
	_ = pw.Set("worker", 7)
	req := httptest.NewRequest(http.MethodDelete, "/weights/worker", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	_, ok := pw.Get("worker")
	if ok {
		t.Fatal("expected weight to be removed")
	}
}

func TestProcessWeightHTTP_ListAll(t *testing.T) {
	mux, pw := newWeightMux()
	_ = pw.Set("a", 1)
	_ = pw.Set("b", 2)
	req := httptest.NewRequest(http.MethodGet, "/weights", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]int
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp))
	}
}
