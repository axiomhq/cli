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

func executeWithAPIToken(t *testing.T, url string, args ...string) error {
	t.Helper()
	cmd := NewCmd(&cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test",
			Deployments:      map[string]config.Deployment{"test": {URL: url, Token: "xaat-test"}},
		},
		IO: terminal.TestIO(),
	})
	cmd.SetArgs(args)
	testutil.SetupCmd(cmd)
	_, err := cmd.ExecuteC()
	return err
}

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

			err := executeWithAPIToken(t, srv.URL, tt.args...)
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

func TestDatasetCreatePreservesAPIErrors(t *testing.T) {
	const message = "token cannot create datasets"
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			requests := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != http.MethodPost || r.URL.Path != "/v2/datasets" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"message":"` + message + `"}`))
			}))
			defer srv.Close()

			err := executeWithAPIToken(t, srv.URL, "create", "--name=logs", "--description=Logs", "--type=events")
			srv.Close()
			if apiErr, ok := errors.AsType[axiom.HTTPError](err); !ok || apiErr.Status != status || apiErr.Message != message {
				t.Fatalf("error = %v, want HTTP %d with message %q", err, status, message)
			}
			if requests != 1 {
				t.Fatalf("got %d requests, want exactly one; a rejected write must not be retried", requests)
			}
		})
	}
}
