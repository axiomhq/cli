package ingest

import (
	"strings"
	"testing"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_detectContentType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  axiom.ContentType
	}{
		{
			name:  "json",
			input: `[{"a":"b"}, {"c":"d"}]`,
			want:  axiom.JSON,
		},
		{
			// Not a problem that a single line of ndjson is classified as json.
			name:  "json",
			input: `{"a":"b"}`,
			want:  axiom.JSON,
		},
		{
			name: "ndjson",
			input: `{"a":"b"}
				{"c":"d"}`,
			want: axiom.NDJSON,
		},
		{
			name: "csv-comma",
			input: `Year,Make,Model,Length
				1997,Ford,E350,2.35
				2000,Mercury,Cougar,2.38`,
			want: axiom.CSV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := detectContentType(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}
