package supervisor

import (
	"fmt"
	"testing"
)

func TestOutputBuffer_RecordAndAll(t *testing.T) {
	b := NewOutputBuffer(10)
	b.Record("web", "stdout", "hello")
	b.Record("web", "stderr", "error")

	all := b.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Line != "hello" || all[0].Stream != "stdout" {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
}

func TestOutputBuffer_TimestampIsSet(t *testing.T) {
	b := NewOutputBuffer(10)
	b.Record("svc", "stdout", "line")
	all := b.All()
	if all[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestOutputBuffer_EvictsOldestWhenFull(t *testing.T) {
	b := NewOutputBuffer(3)
	b.Record("p", "stdout", "first")
	b.Record("p", "stdout", "second")
	b.Record("p", "stdout", "third")
	b.Record("p", "stdout", "fourth")

	all := b.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Line != "second" {
		t.Errorf("expected oldest evicted, got %q", all[0].Line)
	}
	if all[2].Line != "fourth" {
		t.Errorf("expected newest last, got %q", all[2].Line)
	}
}

func TestOutputBuffer_ForProcess(t *testing.T) {
	b := NewOutputBuffer(20)
	b.Record("web", "stdout", "w1")
	b.Record("db", "stdout", "d1")
	b.Record("web", "stderr", "w2")

	web := b.ForProcess("web")
	if len(web) != 2 {
		t.Fatalf("expected 2 web entries, got %d", len(web))
	}
	db := b.ForProcess("db")
	if len(db) != 1 {
		t.Fatalf("expected 1 db entry, got %d", len(db))
	}
}

func TestOutputBuffer_Len(t *testing.T) {
	b := NewOutputBuffer(10)
	if b.Len() != 0 {
		t.Errorf("expected 0, got %d", b.Len())
	}
	b.Record("p", "stdout", "x")
	if b.Len() != 1 {
		t.Errorf("expected 1, got %d", b.Len())
	}
}

func TestOutputBuffer_Clear(t *testing.T) {
	b := NewOutputBuffer(10)
	for i := 0; i < 5; i++ {
		b.Record("p", "stdout", fmt.Sprintf("line%d", i))
	}
	b.Clear()
	if b.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", b.Len())
	}
}

func TestOutputBuffer_DefaultMaxSize(t *testing.T) {
	b := NewOutputBuffer(0)
	for i := 0; i < 201; i++ {
		b.Record("p", "stdout", fmt.Sprintf("line%d", i))
	}
	if b.Len() != 200 {
		t.Errorf("expected 200 (default cap), got %d", b.Len())
	}
}
