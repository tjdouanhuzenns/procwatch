package supervisor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNotifierLog_RecordAndAll(t *testing.T) {
	l := NewNotifierLog(10)
	l.Record(NotifyEvent{Process: "svc", State: "running", Timestamp: time.Now().UTC()})
	l.Record(NotifyEvent{Process: "svc", State: "failed", Timestamp: time.Now().UTC()})

	all := l.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 records, got %d", len(all))
	}
	if all[0].State != "running" || all[1].State != "failed" {
		t.Errorf("unexpected records: %+v", all)
	}
}

func TestNotifierLog_EvictsOldest(t *testing.T) {
	l := NewNotifierLog(3)
	for i := 0; i < 5; i++ {
		l.Record(NotifyEvent{Process: "p", State: "running"})
	}
	if len(l.All()) != 3 {
		t.Errorf("expected max 3 records, got %d", len(l.All()))
	}
}

func TestNotifierHTTP_Endpoint(t *testing.T) {
	l := NewNotifierLog(10)
	l.Record(NotifyEvent{Process: "web", State: "stopped", Message: "exit 2", Timestamp: time.Now().UTC()})

	mux := http.NewServeMux()
	RegisterNotifierRoutes(mux, l)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var records []notifyRecord
	if err := json.NewDecoder(rec.Body).Decode(&records); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Process != "web" || records[0].State != "stopped" {
		t.Errorf("unexpected record: %+v", records[0])
	}
}

func TestNotifierLog_DefaultMaxSize(t *testing.T) {
	l := NewNotifierLog(0) // should default to 100
	for i := 0; i < 105; i++ {
		l.Record(NotifyEvent{Process: "p", State: "running"})
	}
	if len(l.All()) != 100 {
		t.Errorf("expected 100 records, got %d", len(l.All()))
	}
}
