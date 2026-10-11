package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestStatus_EmptyConfig(t *testing.T) {
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Deployments: make(map[string]config.Deployment),
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	assert.ErrorIs(t, err, cmdutil.ErrSilent)
	assert.Contains(t, errBuf.String(), "No deployments configured!")
	assert.Contains(t, errBuf.String(), "axiom auth login")
	assert.Empty(t, outBuf.String())
}

func TestStatusCmd_EmptyConfig(t *testing.T) {
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Deployments: make(map[string]config.Deployment),
		},
		IO: testIO,
	}

	cmd := newStatusCmd(f)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err := cmd.ExecuteContext(context.Background())
	assert.ErrorIs(t, err, cmdutil.ErrSilent)
	assert.Contains(t, errBuf.String(), "No deployments configured!")
}

func TestStatus_NilConfig(t *testing.T) {
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: nil,
		IO:     testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration provided")
}

func TestStatus_MissingToken(t *testing.T) {
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test-dep",
			Deployments: map[string]config.Deployment{
				"test-dep": {
					URL:   "https://api.axiom.co",
					Token: "",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	assert.ErrorIs(t, err, cmdutil.ErrSilent)
	assert.Contains(t, outBuf.String(), "test-dep")
	assert.Contains(t, outBuf.String(), "missing token")
}

func TestStatus_Unauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "unauthenticated",
		})
	}))
	defer srv.Close()

	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test-dep",
			Deployments: map[string]config.Deployment{
				"test-dep": {
					URL:   srv.URL,
					Token: "xapt-invalid-token",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	assert.ErrorIs(t, err, cmdutil.ErrSilent)
	assert.Contains(t, outBuf.String(), "test-dep")
	assert.Contains(t, outBuf.String(), "Invalid credentials")
}

func TestStatus_SuccessWithoutOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/user", "/v2/user":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "usr_123",
				"name":  "Alice Smith",
				"email": "alice@example.com",
			})
		case "/v1/datasets", "/v2/datasets":
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test-dep",
			Deployments: map[string]config.Deployment{
				"test-dep": {
					URL:   srv.URL,
					Token: "xapt-valid-token",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	t.Logf("outBuf: %q, errBuf: %q", outBuf.String(), errBuf.String())
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "test-dep")
	assert.Contains(t, outBuf.String(), "Logged in as Alice Smith")
}

func TestStatus_SuccessWithOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/user" || r.URL.Path == "/v2/user":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "usr_123",
				"name":  "Alice Smith",
				"email": "alice@example.com",
			})
		case strings.HasPrefix(r.URL.Path, "/v1/orgs/") || strings.HasPrefix(r.URL.Path, "/v2/orgs/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "org_456",
				"name": "Acme Corporation",
			})
		case r.URL.Path == "/v1/datasets" || r.URL.Path == "/v2/datasets":
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test-dep",
			Deployments: map[string]config.Deployment{
				"test-dep": {
					URL:            srv.URL,
					Token:          "xapt-valid-token",
					OrganizationID: "org_456",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "test-dep")
	assert.Contains(t, outBuf.String(), "Logged in to Acme Corporation as Alice Smith")
}

func TestStatus_UserWithoutName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/user", "/v2/user":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "usr_123",
				"name":  "",
				"email": "alice@example.com",
			})
		case "/v1/datasets", "/v2/datasets":
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			ActiveDeployment: "test-dep",
			Deployments: map[string]config.Deployment{
				"test-dep": {
					URL:   srv.URL,
					Token: "xapt-valid-token",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
	}

	err := runStatus(context.Background(), opts)
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "test-dep")
	assert.Contains(t, outBuf.String(), "Logged in as alice@example.com")
}

func TestStatus_UnknownAlias(t *testing.T) {
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	testIO := terminal.NewTestIO(nil, &outBuf, &errBuf)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Deployments: map[string]config.Deployment{
				"existing-dep": {
					URL:   "https://api.axiom.co",
					Token: "xapt-token",
				},
			},
		},
		IO: testIO,
	}

	opts := &statusOptions{
		Factory: f,
		Alias:   "unknown-dep",
	}

	err := runStatus(context.Background(), opts)
	assert.ErrorIs(t, err, cmdutil.ErrSilent)
	assert.Contains(t, errBuf.String(), "Chosen deployment unknown-dep is not configured!")
	assert.Empty(t, outBuf.String())
}
