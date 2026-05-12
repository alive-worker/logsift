package filter

import (
	"testing"
	"time"

	"github.com/alive-worker/logsift/internal/parser"
)

func mk(level string, ts time.Time, msg string) *parser.Entry {
	return &parser.Entry{Level: level, Timestamp: ts, Message: msg, Extra: map[string]any{}}
}

func TestLevelFilter(t *testing.T) {
	f := NewLevelFilter("error, warn")
	if !f.Keep(mk("warn", time.Now(), "")) {
		t.Fatalf("warn should pass")
	}
	if f.Keep(mk("debug", time.Now(), "")) {
		t.Fatalf("debug should fail")
	}
}

func TestSinceFilter_dropsZeroTimestamp(t *testing.T) {
	f := &SinceFilter{Cutoff: time.Now().Add(-time.Hour)}
	if f.Keep(mk("info", time.Time{}, "")) {
		t.Fatalf("zero timestamp must be dropped")
	}
}

func TestSinceFilter_keepsRecent(t *testing.T) {
	now := time.Now()
	f := &SinceFilter{Cutoff: now.Add(-time.Hour)}
	if !f.Keep(mk("info", now.Add(-30*time.Minute), "")) {
		t.Fatalf("30m ago should be kept")
	}
	if f.Keep(mk("info", now.Add(-2*time.Hour), "")) {
		t.Fatalf("2h ago should be dropped")
	}
}

func TestGrepFilter(t *testing.T) {
	f := &GrepFilter{Needle: "timeout"}
	if !f.Keep(mk("error", time.Now(), "upstream TIMEOUT after 3s")) {
		t.Fatalf("case-insensitive match failed")
	}
	if f.Keep(mk("info", time.Now(), "ok")) {
		t.Fatalf("non-matching kept")
	}
}

func TestChain_shortCircuits(t *testing.T) {
	chain := Chain{
		NewLevelFilter("error"),
		&GrepFilter{Needle: "timeout"},
	}
	if chain.Keep(mk("info", time.Now(), "timeout")) {
		t.Fatalf("level filter should reject")
	}
	if !chain.Keep(mk("error", time.Now(), "redis timeout")) {
		t.Fatalf("both should pass")
	}
}
