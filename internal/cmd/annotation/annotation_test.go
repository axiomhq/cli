package annotation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func execute(t *testing.T, handler http.HandlerFunc, args ...string) error {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cmd := NewCmd(&cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test",
			Deployments:      map[string]config.Deployment{"test": {URL: srv.URL, Token: "xaat-test"}},
		},
		IO: terminal.TestIO(),
	})
	cmd.SetArgs(args)

	testutil.SetupCmd(cmd)

	_, err := cmd.ExecuteC()
	return err
}

func TestAPIToken(t *testing.T) {
	err := execute(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}, "list", "--format=json")
	if err != nil {
		t.Errorf("annotation list with an API token returned error: %v", err)
	}
}

func TestUpdateNothing(t *testing.T) {
	err := execute(t, func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	}, "update", "ann_1", "--datasets=")
	if err == nil || !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("annotation update without changes returned error %v, want nothing to update", err)
	}
}

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
