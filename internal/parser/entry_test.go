package parser

import (
	"testing"
	"time"
)

func TestParseLine_full(t *testing.T) {
	e, err := ParseLine(`{"ts":"2026-05-12T10:00:00Z","level":"info","service":"api","msg":"hello","port":8080}`)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if e.Level != "info" || e.Service != "api" || e.Message != "hello" {
		t.Fatalf("fields wrong: %+v", e)
	}
	if e.Timestamp.IsZero() {
		t.Fatalf("timestamp not parsed")
	}
	if e.Extra["port"].(float64) != 8080 {
		t.Fatalf("extra missing or wrong: %+v", e.Extra)
	}
}

func TestParseLine_empty(t *testing.T) {
	e, err := ParseLine("")
	if err != nil || e != nil {
		t.Fatalf("expected (nil,nil); got (%v,%v)", e, err)
	}
}

func TestParseLine_invalidJSON(t *testing.T) {
	_, err := ParseLine(`not json`)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseTimestamp_rfc3339(t *testing.T) {
	got, err := parseTimestamp("2026-05-12T10:00:00Z")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
