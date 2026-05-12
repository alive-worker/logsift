package parser

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"2h", 2 * time.Hour},
		{"90s", 90 * time.Second},
		{"3d", 72 * time.Hour},
	}
	for _, c := range cases {
		got, err := ParseSince(c.in)
		if err != nil {
			t.Errorf("%s: unexpected err %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %v want %v", c.in, got, c.want)
		}
	}
}

func TestParseSince_invalid(t *testing.T) {
	if _, err := ParseSince(""); err == nil {
		t.Fatalf("empty should error")
	}
	if _, err := ParseSince("xd"); err == nil {
		t.Fatalf("xd should error")
	}
}
