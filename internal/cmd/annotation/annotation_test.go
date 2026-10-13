package annotation

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		input   string
		want    time.Time
		wantErr bool
	}{
		{"", time.Time{}, false},
		{"now", now, false},
		{"-1h", now.Add(-time.Hour), false},
		{"2026-09-17T08:00:00+02:00", now.Add(-4 * time.Hour), false},
		{"2026-09-17 10:00:00", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTime(tt.input, now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTime(%q) error = %v, want error %t", tt.input, err, tt.wantErr)
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
