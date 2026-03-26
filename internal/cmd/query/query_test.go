package query

import (
	"testing"
	"time"

	"github.com/axiomhq/axiom-go/axiom/query"
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

func TestTableRowAtIndex(t *testing.T) {
	table := query.Table{
		Fields: []query.Field{
			{Name: "organizations"},
			{Name: "datasets"},
			{Name: "spans"},
		},
		Columns: []query.Column{
			{"102 (+2)"},
			{"132 (+3)"},
			{"49.3M (+2.5M)"},
		},
	}

	row := tableRowAtIndex(table, 0)

	if got := row["organizations"]; got != "102 (+2)" {
		t.Errorf("organizations = %v, want %q", got, "102 (+2)")
	}
	if got := row["datasets"]; got != "132 (+3)" {
		t.Errorf("datasets = %v, want %q", got, "132 (+3)")
	}
	if got := row["spans"]; got != "49.3M (+2.5M)" {
		t.Errorf("spans = %v, want %q", got, "49.3M (+2.5M)")
	}
}

func TestTableRowAtIndexMultipleRows(t *testing.T) {
	table := query.Table{
		Fields: []query.Field{
			{Name: "status"},
			{Name: "count_", Aggregation: &query.Aggregation{Op: query.OpCount}},
		},
		Columns: []query.Column{
			{"ok", "err"},
			{float64(42), float64(3)},
		},
	}

	row0 := tableRowAtIndex(table, 0)
	if got := row0["status"]; got != "ok" {
		t.Errorf("row0 status = %v, want %q", got, "ok")
	}
	if got := row0["count_"]; got != float64(42) {
		t.Errorf("row0 count_ = %v, want %v", got, 42)
	}

	row1 := tableRowAtIndex(table, 1)
	if got := row1["status"]; got != "err" {
		t.Errorf("row1 status = %v, want %q", got, "err")
	}
	if got := row1["count_"]; got != float64(3) {
		t.Errorf("row1 count_ = %v, want %v", got, 3)
	}
}

func TestTableHasAggregation(t *testing.T) {
	noAgg := query.Table{
		Fields: []query.Field{
			{Name: "org"},
			{Name: "spans"},
		},
	}
	if tableHasAggregation(noAgg) {
		t.Error("expected no aggregation for plain fields")
	}

	withAgg := query.Table{
		Fields: []query.Field{
			{Name: "count_", Aggregation: &query.Aggregation{Op: query.OpCount}},
		},
	}
	if !tableHasAggregation(withAgg) {
		t.Error("expected aggregation when field has Aggregation set")
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
