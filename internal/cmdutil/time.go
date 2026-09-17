package cmdutil

import (
	"fmt"
	"strings"
	"time"
)

// RelativeTime resolves "now" or a duration like "-1h" or "-7d" against now.
// It reports false if s is neither.
func RelativeTime(s string, now time.Time) (time.Time, bool) {
	if s == "now" {
		return now, true
	}
	d, err := parseDuration(s)
	if err != nil {
		return time.Time{}, false
	}
	return now.Add(d), true
}

// parseDuration parses a duration string like time.ParseDuration but also
// supports "d" (day = 24h) and "w" (week = 168h) suffixes. Mixed units like
// "1w2d12h" are supported.
func parseDuration(s string) (time.Duration, error) {
	// Fast path: try stdlib first, which handles everything except d/w.
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Check if the string contains "d" or "w" at all.
	if !strings.ContainsAny(s, "dw") {
		return 0, fmt.Errorf("time: invalid duration %q", s)
	}

	// Replace "d" and "w" with their hour equivalents so the stdlib can
	// parse the result. We accumulate extra hours from d/w conversions and
	// append them to the remaining string.
	neg := strings.HasPrefix(s, "-")
	raw := strings.TrimLeft(s, "+-")

	var extra time.Duration
	var rest strings.Builder
	i := 0
	for i < len(raw) {
		// Scan a number.
		numStart := i
		for i < len(raw) && (raw[i] == '.' || (raw[i] >= '0' && raw[i] <= '9')) {
			i++
		}
		if numStart == i {
			return 0, fmt.Errorf("time: invalid duration %q", s)
		}
		num := raw[numStart:i]

		// Scan the unit suffix.
		unitStart := i
		for i < len(raw) && (raw[i] < '0' || raw[i] > '9') && raw[i] != '.' {
			i++
		}
		if unitStart == i {
			return 0, fmt.Errorf("time: missing unit in duration %q", s)
		}
		unit := raw[unitStart:i]

		switch unit {
		case "d":
			v, err := time.ParseDuration(num + "h")
			if err != nil {
				return 0, fmt.Errorf("time: invalid duration %q", s)
			}
			extra += v * 24
		case "w":
			v, err := time.ParseDuration(num + "h")
			if err != nil {
				return 0, fmt.Errorf("time: invalid duration %q", s)
			}
			extra += v * 168
		default:
			rest.WriteString(num)
			rest.WriteString(unit)
		}
	}

	var d time.Duration
	if rest.Len() > 0 {
		var err error
		d, err = time.ParseDuration(rest.String())
		if err != nil {
			return 0, err
		}
	}
	d += extra
	if neg {
		d = -d
	}
	return d, nil
}
