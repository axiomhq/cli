package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestStatusCmd_Alias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer xapt-bad" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","name":"test"}`))
	}))
	t.Cleanup(srv.Close)

	tests := []struct {
		name    string
		args    []string
		wantErr error
	}{
		{name: "no alias checks all deployments", args: []string{}, wantErr: cmdutil.ErrSilent},
		{name: "alias checks only that deployment", args: []string{"good"}},
		{name: "unknown alias", args: []string{"unknown"}, wantErr: cmdutil.ErrSilent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newStatusCmd(&cmdutil.Factory{
				Config: &config.Config{
					Deployments: map[string]config.Deployment{
						"good": {URL: srv.URL, Token: "xapt-good"},
						"bad":  {URL: srv.URL, Token: "xapt-bad"},
					},
				},
				IO: terminal.TestIO(),
			})
			cmd.SetArgs(tt.args)

			testutil.SetupCmd(cmd)

			_, err := cmd.ExecuteC()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
