package query

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		// Stdlib units pass through.
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"1h30m", time.Hour + 30*time.Minute},
		{"500ms", 500 * time.Millisecond},

		// Day and week support.
		{"1d", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2w", 2 * 7 * 24 * time.Hour},

		// Mixed units.
		{"1d12h", 36 * time.Hour},
		{"1w2d", 9 * 24 * time.Hour},
		{"1w2d3h", 9*24*time.Hour + 3*time.Hour},

		// Fractional values.
		{"0.5d", 12 * time.Hour},
		{"1.5w", 252 * time.Hour},

		// Positive prefix.
		{"+1d", 24 * time.Hour},

		// Negative durations.
		{"-1h", -time.Hour},
		{"-7d", -7 * 24 * time.Hour},
		{"-2w", -2 * 7 * 24 * time.Hour},
		{"-1d12h", -36 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if err != nil {
				t.Fatalf("parseDuration(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDurationErrors(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"7",
		"d",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDuration(input)
			if err == nil {
				t.Errorf("parseDuration(%q) expected error, got nil", input)
			}
		})
	}
}
