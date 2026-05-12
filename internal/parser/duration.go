package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseSince accepts Go-style "15m", "2h", "90s", "1d". time.ParseDuration
// already handles the first three; we layer day support on top because logs
// older than a day are a common ask.
func ParseSince(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.HasSuffix(raw, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(raw, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid day duration %q", raw)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return time.ParseDuration(raw)
}
