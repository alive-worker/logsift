package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

// Entry is one parsed NDJSON log record. Unknown fields stay in Extra so
// --where expressions can reference them.
type Entry struct {
	Timestamp time.Time
	Level     string
	Service   string
	Message   string
	Extra     map[string]any
	Raw       string
}

// ParseLine consumes a single NDJSON line. Empty / whitespace-only lines
// return (nil, nil); structurally invalid lines return an error.
func ParseLine(line string) (*Entry, error) {
	if line == "" {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	e := &Entry{Extra: map[string]any{}, Raw: line}
	for k, v := range m {
		switch k {
		case "ts", "time", "timestamp":
			if s, ok := v.(string); ok {
				if t, err := parseTimestamp(s); err == nil {
					e.Timestamp = t
				}
			}
		case "level":
			if s, ok := v.(string); ok {
				e.Level = s
			}
		case "service":
			if s, ok := v.(string); ok {
				e.Service = s
			}
		case "msg", "message":
			if s, ok := v.(string); ok {
				e.Message = s
			}
		default:
			e.Extra[k] = v
		}
	}
	return e, nil
}

// timestampLayouts is the ordered list of formats parseTimestamp tries.
// Known limitation: layouts without an explicit zone (e.g. plain
// "2006-01-02T15:04:05") are not in this list, so naive timestamps fall
// through to "unrecognised" and the row ends up with a zero Time. The
// trial task #2 is to fix that.
var timestampLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
}

func parseTimestamp(s string) (time.Time, error) {
	for _, l := range timestampLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised timestamp %q", s)
}
