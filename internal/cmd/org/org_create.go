package org

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/pkg/version"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

type createOptions struct {
	Name           string `json:"name,omitempty"`
	EdgeDeployment string `json:"edgeDeployment,omitempty"`
}

type provisionedOrg struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	DefaultEdgeDeployment string    `json:"defaultEdgeDeployment"`
	ExpiresAt             time.Time `json:"expiresAt"`
	ClaimURL              string    `json:"claimUrl"`
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var opts createOptions
	var temporary bool
	cmd := &cobra.Command{
		Use:   "create --temporary",
		Short: "Create a temporary organization without logging in",
		Long: heredoc.Doc(`
			Create a temporary organization without logging in. Use the claim URL
			before the returned expiry time to keep the organization and its data.

			The token is saved privately in the CLI config under the organization ID.
			Existing deployments and the default stay unchanged. The first deployment
			becomes the default. Output is JSON with the claim URL, never the API token.

			The request is not retried. If saving fails, the error names a private
			response file for recovery. Do not create another organization to retry
			a failed save. The response file contains credentials; keep it private.
		`),
		Example: heredoc.Doc(`
			$ axiom org create --temporary --name=my-project
			$ axiom dataset create -D <returned-org-id> --name=http-logs --description="HTTP logs" --type=events
		`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !temporary {
				return cmdutil.NewFlagErrorf("--temporary is required")
			}
			return runCreate(cmd, f, opts)
		},
	}
	cmd.Flags().BoolVar(&temporary, "temporary", false, "Create an organization that expires unless claimed")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Organization name (generated if omitted)")
	cmd.Flags().StringVar(&opts.EdgeDeployment, "edge-deployment", "", "Edge deployment (uses the server default if omitted)")
	return cmd
}

func runCreate(cmd *cobra.Command, f *cmdutil.Factory, opts createOptions) error {
	if f.Config.URLOverride == "" {
		if err := cmdutil.NeedsValidDeployment(f, &f.Config.ActiveDeployment)(cmd, nil); err != nil {
			return err
		}
	}
	baseURL := "https://api.axiom.co"
	if deployment, ok := f.Config.GetActiveDeployment(); ok && deployment.URL != "" {
		baseURL = deployment.URL
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "https://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("invalid API URL")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/v2/orgs/provision"
	data, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "axiom-cli/"+version.Release())

	// NOTE: the token and claim URL are returned once, so reserve private recovery storage before creating the org.
	recovery, err := os.CreateTemp(filepath.Dir(f.Config.ConfigFilePath), ".axiom-org-*.json")
	if err != nil {
		return fmt.Errorf("cannot prepare credential storage: %w", err)
	}
	defer recovery.Close()
	keepRecovery := false
	defer func() {
		if !keepRecovery {
			_ = os.Remove(recovery.Name())
		}
	}()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	defer transport.CloseIdleConnections()
	if f.Config.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Explicit CLI --insecure flag.
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Minute,
		// NOTE: a redirect must not replay this non-idempotent POST.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		// NOTE: wrapping cancellation would hide this message behind main's generic cancellation error.
		return fmt.Errorf("organization creation could not be confirmed; the request was not retried: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("organization creation returned HTTP %d; the request was not retried", resp.StatusCode)
	}
	keepRecovery = true
	if _, err := io.Copy(recovery, resp.Body); err != nil {
		return fmt.Errorf("incomplete organization response in %q; do not repeat creation: %v", recovery.Name(), err)
	}
	if err := recovery.Sync(); err != nil {
		return fmt.Errorf("cannot sync organization response in %q; do not repeat creation: %w", recovery.Name(), err)
	}
	if _, err := recovery.Seek(0, io.SeekStart); err != nil {
		return err
	}
	var result struct {
		provisionedOrg
		Token string `json:"token"`
	}
	// NOTE: decoder errors can include response values, so report the recovery path without the decoder error.
	if err := json.NewDecoder(recovery).Decode(&result); err != nil {
		return fmt.Errorf("invalid organization response saved in %q; do not repeat creation", recovery.Name())
	}
	if result.ID == "" || result.Token == "" || result.ClaimURL == "" || result.ExpiresAt.IsZero() {
		return fmt.Errorf("incomplete organization response saved in %q; do not repeat creation", recovery.Name())
	}
	if err := recovery.Close(); err != nil {
		return fmt.Errorf("cannot close organization response in %q; do not repeat creation: %w", recovery.Name(), err)
	}
	if err := config.AddDeployment(cmd.Context(), f.Config.ConfigFilePath, result.ID, config.Deployment{
		URL:            baseURL,
		Token:          result.Token,
		OrganizationID: result.ID,
		ClaimURL:       result.ClaimURL,
		ExpiresAt:      result.ExpiresAt.Format(time.RFC3339Nano),
	}); err != nil {
		// NOTE: keep the recovery path visible even when the user cancels while waiting for the config lock.
		return fmt.Errorf("organization created but config not saved; recover credentials from %q instead of repeating creation: %v", recovery.Name(), err)
	}
	if err := os.Remove(recovery.Name()); err != nil {
		return fmt.Errorf("organization %q saved; remove the extra credential file %q: %w", result.ID, recovery.Name(), err)
	}
	if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result.provisionedOrg); err != nil {
		return fmt.Errorf("organization %q saved in %q but output failed; do not repeat creation: %w", result.ID, f.Config.ConfigFilePath, err)
	}
	return nil
}
