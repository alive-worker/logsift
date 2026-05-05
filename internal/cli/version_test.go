package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRun_versionPrintsAndExits(t *testing.T) {
	opts, err := ParseArgs([]string{"--version"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out, errBuf bytes.Buffer
	if err := Run(opts, strings.NewReader("anything"), &out, &errBuf, time.Now()); err != nil {
		t.Fatalf("run: %v", err)
	}
	got := strings.TrimSpace(out.String())
	if got != "logsift "+Version {
		t.Fatalf("want %q got %q", "logsift "+Version, got)
	}
}
