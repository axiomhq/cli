package dataset

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestDatasetTokens(t *testing.T) {
	for _, token := range []struct {
		name, value, orgHeader string
	}{
		{"api", "xaat-test", ""},
		{"personal", "xapt-test", "test-org"},
	} {
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
			for _, status := range []int{http.StatusOK, http.StatusUnauthorized, http.StatusForbidden} {
				t.Run(token.name+"/"+tt.args[0]+"/"+http.StatusText(status), func(t *testing.T) {
					requests := 0
					srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						if got := r.Header.Get("Authorization"); got != "Bearer "+token.value {
							t.Errorf("Authorization = %q, want the configured token", got)
						}
						if r.Method == http.MethodGet && r.URL.Path == "/v2/user" && token.name == "personal" {
							_, _ = w.Write([]byte(`{"name":"Test user"}`))
							return
						}
						if got := r.Header.Get("X-Axiom-Org-Id"); got != token.orgHeader {
							t.Errorf("organization header = %q, want %q", got, token.orgHeader)
						}
						switch {
						case r.Method == tt.method && r.URL.Path == tt.path:
							requests++
							w.WriteHeader(status)
							if status != http.StatusOK {
								_, _ = w.Write([]byte(`{"message":"request denied"}`))
								return
							}
							_, _ = w.Write([]byte(tt.response))
						case token.name == "personal" && r.Method == http.MethodGet && r.URL.Path == "/v2/datasets":
							_, _ = w.Write([]byte(`[{"name":"logs"}]`))
						default:
							t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
							w.WriteHeader(http.StatusNotFound)
						}
					}))
					defer srv.Close()
					cmd := NewCmd(&cmdutil.Factory{
						Config: &config.Config{
							ActiveDeployment: "test",
							Deployments: map[string]config.Deployment{
								"test": {URL: srv.URL, Token: token.value, OrganizationID: "test-org"},
							},
						},
						IO: terminal.TestIO(),
					})
					cmd.SetArgs(tt.args)
					testutil.SetupCmd(cmd)
					_, err := cmd.ExecuteC()
					srv.Close()
					if status == http.StatusOK {
						if err != nil {
							t.Fatalf("dataset command failed: %v", err)
						}
					} else if !errors.Is(err, axiom.HTTPError{Status: status}) {
						t.Fatalf("error = %v, want HTTP %d", err, status)
					}
					if requests == 0 {
						t.Fatal("dataset command did not reach the API")
					}
				})
			}
		}
	}
}
