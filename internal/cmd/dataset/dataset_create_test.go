package dataset

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestCreateDatasetType(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    []string
		kind    string
		wantErr bool
	}{
		{name: "default events", kind: "axiom:events:v1"},
		{name: "explicit events", args: []string{"--type", "events"}, kind: "axiom:events:v1"},
		{name: "metrics", args: []string{"--type", "metrics"}, kind: "otel:metrics:v1"},
		{name: "invalid", args: []string{"--type", "logs"}, wantErr: true},
		{name: "empty", args: []string{"--type="}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var requests []axiom.DatasetCreateRequest
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost && r.URL.Path == "/v2/datasets" {
					var req axiom.DatasetCreateRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Error(err)
					}
					requests = append(requests, req)
					_, _ = w.Write([]byte(`{"name":"app-data"}`))
					return
				}
				if tt.wantErr {
					t.Errorf("invalid type made a request: %s %s", r.Method, r.URL.Path)
				}
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()
			cmd := newCreateCmd(&cmdutil.Factory{
				Config: &config.Config{ActiveDeployment: "test", Deployments: map[string]config.Deployment{"test": {URL: srv.URL, Token: "xapt-test", OrganizationID: "test-org"}}},
				IO:     terminal.TestIO(),
			})
			cmd.SetArgs(append([]string{"--name", "app-data", "--description", "App telemetry"}, tt.args...))
			testutil.SetupCmd(cmd)
			_, err := cmd.ExecuteC()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "invalid dataset type") {
					t.Fatalf("error = %v, want invalid dataset type", err)
				}
				if len(requests) != 0 {
					t.Fatalf("invalid type created a dataset")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) != 1 {
				t.Fatalf("got %d create requests, want 1", len(requests))
			}
			req := requests[0]
			if req.Kind != tt.kind || req.Name != "app-data" || req.Description != "App telemetry" {
				t.Errorf("unexpected create request: %+v", req)
			}
		})
	}
}
