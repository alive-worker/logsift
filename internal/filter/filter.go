package filter

import (
	"strings"
	"time"

	"github.com/alive-worker/logsift/internal/parser"
)

// Filter narrows a stream of entries. Implementations should be cheap and
// stateless; the chain short-circuits on the first reject.
type Filter interface {
	Keep(e *parser.Entry) bool
}

type Chain []Filter

func (c Chain) Keep(e *parser.Entry) bool {
	for _, f := range c {
		if !f.Keep(e) {
			return false
		}
	}
	return true
}

// --- concrete filters -------------------------------------------------------

type LevelFilter struct{ Allowed map[string]struct{} }

func NewLevelFilter(csv string) *LevelFilter {
	f := &LevelFilter{Allowed: map[string]struct{}{}}
	for _, raw := range strings.Split(csv, ",") {
		v := strings.TrimSpace(strings.ToLower(raw))
		if v != "" {
			f.Allowed[v] = struct{}{}
		}
	}
	return f
}

func (l *LevelFilter) Keep(e *parser.Entry) bool {
	if len(l.Allowed) == 0 {
		return true
	}
	_, ok := l.Allowed[strings.ToLower(e.Level)]
	return ok
}

type SinceFilter struct {
	Cutoff time.Time
}

func (s *SinceFilter) Keep(e *parser.Entry) bool {
	if e.Timestamp.IsZero() {
		return false
	}
	return !e.Timestamp.Before(s.Cutoff)
}

type GrepFilter struct{ Needle string }

func (g *GrepFilter) Keep(e *parser.Entry) bool {
	if g.Needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(e.Message), strings.ToLower(g.Needle))
}
