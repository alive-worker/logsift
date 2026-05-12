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

func parseTimestamp(s string) (time.Time, error) {
	// BUG-FOR-PROMPT: this list only covers RFC3339 with explicit offset.
	// Timestamps like "2026-01-02T15:04:05" (no zone) or "...Z" written as
	// "2026-01-02 15:04:05Z" fall through and the entry ends up with a zero
	// Timestamp, causing --since filtering to drop them silently.
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised timestamp %q", s)
}
