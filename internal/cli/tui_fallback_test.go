package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// When stdout is a *bytes.Buffer (not a TTY), --tui should warn on stderr
// and emit the same output the normal path would have produced.
func TestRun_tuiFallsBackWhenNotATTY(t *testing.T) {
	opts, err := ParseArgs([]string{"--tui", "--output=json"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out, errBuf bytes.Buffer
	now := time.Date(2026, 5, 12, 10, 0, 30, 0, time.UTC)
	if err := Run(opts, strings.NewReader(fixture), &out, &errBuf, now); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(errBuf.String(), "not a TTY") {
		t.Fatalf("expected TTY warning on stderr; got %q", errBuf.String())
	}
	// All 4 fixture lines should appear in stdout (no level/since filter set).
	if strings.Count(out.String(), "\n") != 4 {
		t.Fatalf("want 4 output lines, got:\n%s", out.String())
	}
}
