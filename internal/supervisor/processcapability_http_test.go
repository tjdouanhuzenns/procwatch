package supervisor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCapabilityMux() (*http.ServeMux, *ProcessCapabilityStore) {
	mux := http.NewServeMux()
	store := NewProcessCapabilityStore()
	RegisterProcessCapabilityRoutes(mux, store)
	return mux, store
}

func TestProcessCapabilityHTTP_PutAndGet(t *testing.T) {
	mux, _ := newCapabilityMux()
	body, _ := json.Marshal([]string{"net_bind", "read_logs"})
	req := httptest.NewRequest(http.MethodPut, "/capabilities/web", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/capabilities/web", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var caps []string
	json.NewDecoder(rr2.Body).Decode(&caps)
	if len(caps) != 2 || caps[0] != "net_bind" {
		t.Errorf("unexpected caps: %v", caps)
	}
}

func TestProcessCapabilityHTTP_GetMissingReturns404(t *testing.T) {
	mux, _ := newCapabilityMux()
	req := httptest.NewRequest(http.MethodGet, "/capabilities/ghost", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestProcessCapabilityHTTP_Delete(t *testing.T) {
	mux, store := newCapabilityMux()
	_ = store.Set("api", []string{"write_db"})
	req := httptest.NewRequest(http.MethodDelete, "/capabilities/api", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	_, ok := store.Get("api")
	if ok {
		t.Fatal("expected capability to be removed")
	}
}

func TestProcessCapabilityHTTP_ListAll(t *testing.T) {
	mux, store := newCapabilityMux()
	_ = store.Set("a", []string{"x"})
	_ = store.Set("b", []string{"y"})
	req := httptest.NewRequest(http.MethodGet, "/capabilities", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var all []ProcessCapability
	json.NewDecoder(rr.Body).Decode(&all)
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
