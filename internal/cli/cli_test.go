package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const fixture = `{"ts":"2026-05-12T10:00:00Z","level":"info","service":"api","msg":"server started"}
{"ts":"2026-05-12T10:00:05Z","level":"warn","service":"api","msg":"slow query","duration_ms":820}
{"ts":"2026-05-12T10:00:10Z","level":"error","service":"worker","msg":"redis timeout","status":500}
{"ts":"2026-05-12T10:00:20Z","level":"error","service":"api","msg":"upstream timeout","status":504}
`

func runWith(t *testing.T, args []string) string {
	t.Helper()
	opts, err := ParseArgs(args, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	var out, errBuf bytes.Buffer
	now := time.Date(2026, 5, 12, 10, 0, 30, 0, time.UTC)
	if err := Run(opts, strings.NewReader(fixture), &out, &errBuf, now); err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, errBuf.String())
	}
	return out.String()
}

func TestRun_levelFilter(t *testing.T) {
	got := runWith(t, []string{"--level=error", "--output=json"})
	if strings.Count(got, "\n") != 2 {
		t.Fatalf("want 2 error lines, got:\n%s", got)
	}
	if !strings.Contains(got, "redis timeout") || !strings.Contains(got, "upstream timeout") {
		t.Fatalf("missing expected errors:\n%s", got)
	}
}

func TestRun_grepAndWhere(t *testing.T) {
	got := runWith(t, []string{"--grep=timeout", "--where=status>=500", "--output=tsv"})
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d:\n%s", len(lines), got)
	}
}

func TestRun_outputTSVColumns(t *testing.T) {
	got := runWith(t, []string{"--level=info", "--output=tsv"})
	line := strings.TrimRight(got, "\n")
	cols := strings.Split(line, "\t")
	if len(cols) != 4 {
		t.Fatalf("want 4 tsv cols, got %d: %q", len(cols), line)
	}
}
