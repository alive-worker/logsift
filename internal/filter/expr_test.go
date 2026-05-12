package filter

import (
	"testing"
	"time"

	"github.com/alive-worker/logsift/internal/parser"
)

func TestParseExpr_ok(t *testing.T) {
	ex, err := ParseExpr("status>=500")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ex.Field != "status" || ex.Op != ">=" || ex.Value != "500" {
		t.Fatalf("bad parse: %+v", ex)
	}
}

func TestParseExpr_noOp(t *testing.T) {
	if _, err := ParseExpr("status"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestExpr_numeric(t *testing.T) {
	ex, _ := ParseExpr("status>=500")
	e := &parser.Entry{Extra: map[string]any{"status": float64(504)}, Timestamp: time.Now()}
	if !ex.Keep(e) {
		t.Fatalf("504 should match >=500")
	}
	e.Extra["status"] = float64(200)
	if ex.Keep(e) {
		t.Fatalf("200 should not match >=500")
	}
}

func TestExpr_string(t *testing.T) {
	ex, _ := ParseExpr("service==api")
	e := &parser.Entry{Service: "api", Extra: map[string]any{}}
	if !ex.Keep(e) {
		t.Fatalf("api should match")
	}
	e.Service = "worker"
	if ex.Keep(e) {
		t.Fatalf("worker should not match")
	}
}

// NOTE: intentionally missing coverage for malformed expressions
// (empty field, dangling op, quoted values) — the "补测试" prompt fills
// this in.
