package dataset

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestDatasetCommandsAcceptAPIToken(t *testing.T) {
	for _, tt := range []struct {
		args         []string
		method, path string
		response     string
	}{
		{[]string{"create", "--name=logs", "--description=Logs", "--type=events"}, http.MethodPost, "/v2/datasets", `{"name":"logs"}`},
		{[]string{"list", "--format=json"}, http.MethodGet, "/v2/datasets", `[{"name":"logs"}]`},
		{[]string{"update", "logs", "--description=Updated"}, http.MethodPut, "/v2/datasets/logs", `{"name":"logs"}`},
		{[]string{"delete", "logs", "--force"}, http.MethodDelete, "/v2/datasets/logs", ""},
		{[]string{"trim", "logs", "--duration=1h", "--force"}, http.MethodPost, "/v2/datasets/logs/trim", ""},
	} {
		t.Run(tt.args[0], func(t *testing.T) {
			requests := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
				}
				if got := r.Header.Get("Authorization"); got != "Bearer xaat-test" {
					t.Errorf("Authorization = %q, want the configured API token", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer srv.Close()

			cmd := NewCmd(&cmdutil.Factory{
				Config: &config.Config{
					ActiveDeployment: "test",
					Deployments:      map[string]config.Deployment{"test": {URL: srv.URL, Token: "xaat-test"}},
				},
				IO: terminal.TestIO(),
			})
			cmd.SetArgs(tt.args)
			testutil.SetupCmd(cmd)
			_, err := cmd.ExecuteC()
			srv.Close()
			if err != nil {
				t.Fatalf("dataset command failed: %v", err)
			}
			if requests != 1 {
				t.Fatalf("got %d requests, want exactly one", requests)
			}
		})
	}
}
