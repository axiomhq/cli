package org_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/cli/internal/cmd/root"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

const provisionResponse = `{"id":"test-org","name":"Test Org","defaultEdgeDeployment":"cloud.us-east-1.aws","expiresAt":"2027-01-02T12:00:00Z","claimUrl":"https://app.axiom.co/orgs/test-org/claim?token=test-claim","token":"xaat-test-secret"}`

func createCommand(filename, endpoint string, out io.Writer, args ...string) *cobra.Command {
	f := &cmdutil.Factory{
		Config: &config.Config{
			ConfigFilePath: filename,
			// NOTE: existing credentials must not be sent to the public endpoint.
			TokenOverride:          "xapt-test",
			OrganizationIDOverride: "existing-org",
		},
		IO: terminal.TestIO(),
	}
	cmd := root.NewCmd(f)
	testutil.SetupCmd(cmd)
	cmd.SetOut(out)
	cmd.SetArgs(append([]string{"org", "create", "--auth-url", endpoint}, args...))
	return cmd
}

func TestCreateAndUseOrganization(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		assert.Equal(t, http.MethodPost, r.Method)
		switch r.URL.Path {
		case "/v2/orgs/provision":
			assert.Empty(t, r.Header.Get("Authorization"))
			assert.Empty(t, r.Header.Get("X-Axiom-Org-Id"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, `{"name":"Test Org","edgeDeployment":"cloud.us-east-1.aws"}`, string(body))
			_, _ = io.WriteString(w, provisionResponse)
		case "/v2/datasets":
			assert.Equal(t, "Bearer xaat-test-secret", r.Header.Get("Authorization"))
			// NOTE: API tokens identify the organization; the SDK only sends an org header for personal tokens.
			assert.Empty(t, r.Header.Get("X-Axiom-Org-Id"))
			var body map[string]any
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "http-logs", body["name"])
			_, _ = io.WriteString(w, `{"name":"http-logs"}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	filename := filepath.Join(t.TempDir(), "config.toml")
	var out bytes.Buffer
	cmd := createCommand(filename, server.URL, &out, "--temporary", "--name", "Test Org", "--edge-deployment", "cloud.us-east-1.aws")
	require.NoError(t, cmd.Execute())
	require.EqualValues(t, 1, requests.Load())
	require.JSONEq(t, `{"id":"test-org","name":"Test Org","defaultEdgeDeployment":"cloud.us-east-1.aws","expiresAt":"2027-01-02T12:00:00Z","claimUrl":"https://app.axiom.co/orgs/test-org/claim?token=test-claim"}`, out.String())
	tree, err := toml.LoadFile(filename)
	require.NoError(t, err)
	var saved config.Config
	require.NoError(t, tree.Unmarshal(&saved))
	require.Equal(t, "test-org", saved.ActiveDeployment)
	require.Equal(t, config.Deployment{
		URL: server.URL, Token: "xaat-test-secret", OrganizationID: "test-org",
		ClaimURL: "https://app.axiom.co/orgs/test-org/claim?token=test-claim", ExpiresAt: "2027-01-02T12:00:00Z",
	}, saved.Deployments["test-org"])
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filename)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(filename), ".axiom-*"))
	require.NoError(t, err)
	require.Empty(t, files)

	// NOTE: use the saved config through the root command so credential checks run as they do on a later invocation.
	f := &cmdutil.Factory{Config: &saved, IO: terminal.TestIO()}
	datasetCmd := root.NewCmd(f)
	testutil.SetupCmd(datasetCmd)
	datasetCmd.SetArgs([]string{"dataset", "create", "-D", "test-org", "--name", "http-logs", "--description", "HTTP logs", "--type", "events"})
	require.NoError(t, datasetCmd.Execute())
	require.EqualValues(t, 2, requests.Load())
}

func TestCreateDoesNotFollowRedirect(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "/v2/orgs/provision")
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, _ = io.WriteString(w, `{"message":"xaat-do-not-print"}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	var out bytes.Buffer
	err := createCommand(filepath.Join(dir, "config.toml"), server.URL, &out, "--temporary").Execute()
	require.ErrorContains(t, err, "HTTP 307")
	require.ErrorContains(t, err, "not retried")
	require.NotContains(t, err.Error(), "xaat-do-not-print")
	require.EqualValues(t, 1, requests.Load())
	require.Empty(t, out.String())
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestCreateRetainsResponseWhenSaveFails(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "config.toml")
	const original = "active_deployment = 'test-org'\n[deployments.test-org]\ntoken = 'xapt-existing'\n"
	require.NoError(t, os.WriteFile(filename, []byte(original), 0o600))
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{}`, string(body))
		_, _ = io.WriteString(w, provisionResponse)
	}))
	defer server.Close()
	var out bytes.Buffer
	err := createCommand(filename, server.URL, &out, "--temporary").Execute()
	require.ErrorContains(t, err, "organization created but config not saved")
	require.ErrorContains(t, err, `deployment "test-org" already exists`)
	require.EqualValues(t, 1, requests.Load())
	require.Empty(t, out.String())
	files, globErr := filepath.Glob(filepath.Join(dir, ".axiom-org-*.json"))
	require.NoError(t, globErr)
	require.Len(t, files, 1)
	require.ErrorContains(t, err, files[0])
	require.NotContains(t, err.Error(), "xaat-test-secret")
	response, readErr := os.ReadFile(files[0])
	require.NoError(t, readErr)
	require.Equal(t, provisionResponse, string(response))
	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(files[0])
		require.NoError(t, statErr)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
	saved, readErr := os.ReadFile(filename)
	require.NoError(t, readErr)
	require.Equal(t, original, string(saved))
}
