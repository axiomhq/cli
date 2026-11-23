package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"

	"github.com/axiomhq/cli/internal/config"
)

func TestAddDeploymentPreservesStoredConfig(t *testing.T) {
	for _, tc := range []struct{ name, prefix, extra, active string }{
		{name: "implicit default", active: "existing"},
		{name: "explicit default", prefix: "active_deployment = 'other'\n", extra: "[deployments.other]\ntoken = 'xapt-other'\n", active: "other"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "config.toml")
			const existing = "future_setting = true\n[deployments.existing]\nurl = 'https://api.axiom.co'\ntoken = 'xapt-existing'\norg_id = 'existing-org'\nfuture_field = 'preserve-me'\n"
			require.NoError(t, os.WriteFile(filename, []byte(tc.prefix+existing+tc.extra), 0o600))
			t.Setenv("AXIOM_DEPLOYMENT", "ignore-env-default")
			t.Setenv("AXIOM_TOKEN", "ignore-env-token")
			dep := config.Deployment{URL: "https://api.axiom.co", Token: "xaat-new", OrganizationID: "new-org", ClaimURL: "https://app.axiom.co/claim", ExpiresAt: "2027-01-02T12:00:00Z"}
			require.NoError(t, config.AddDeployment(filename, "new.org", dep))
			tree, err := toml.LoadFile(filename)
			require.NoError(t, err)
			require.Equal(t, tc.active, tree.Get("active_deployment"))
			require.Equal(t, true, tree.Get("future_setting"))
			require.Equal(t, "preserve-me", tree.Get("deployments.existing.future_field"))
			require.Equal(t, "xapt-existing", tree.Get("deployments.existing.token"))
			require.Equal(t, dep.Token, tree.GetPath([]string{"deployments", "new.org", "token"}))

			// Later config writes must not discard the claim URL and deadline.
			var cfg config.Config
			require.NoError(t, tree.Unmarshal(&cfg))
			cfg.ConfigFilePath = filename
			require.Equal(t, dep, cfg.Deployments["new.org"])
			if tc.extra != "" {
				require.Equal(t, "xapt-other", cfg.Deployments["other"].Token)
			}
			require.NoError(t, cfg.Write())
			tree, err = toml.LoadFile(filename)
			require.NoError(t, err)
			require.Equal(t, dep.ClaimURL, tree.GetPath([]string{"deployments", "new.org", "claim_url"}))
			require.Equal(t, dep.ExpiresAt, tree.GetPath([]string{"deployments", "new.org", "expires_at"}))
		})
	}
}

func TestAddDeploymentRefusesOverwrite(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.toml")
	const original = "[deployments.existing]\ntoken = 'xapt-existing'\n"
	require.NoError(t, os.WriteFile(filename, []byte(original), 0o600))
	err := config.AddDeployment(filename, "existing", config.Deployment{Token: "xaat-new"})
	require.ErrorContains(t, err, "already exists")
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, original, string(data))
}

func TestAddDeploymentConcurrentSaves(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(filename, []byte("[deployments.existing]\ntoken = 'xapt-existing'\n"), 0o600))
	const count = 8
	start := make(chan struct{})
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for i := range count {
		wg.Go(func() {
			<-start
			alias := fmt.Sprintf("org-%d", i)
			errs <- config.AddDeployment(filename, alias, config.Deployment{Token: "xaat-" + alias})
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	tree, err := toml.LoadFile(filename)
	require.NoError(t, err)
	var cfg config.Config
	require.NoError(t, tree.Unmarshal(&cfg))
	require.Len(t, cfg.Deployments, count+1)
	require.Equal(t, "existing", cfg.ActiveDeployment)
	require.Equal(t, "xapt-existing", cfg.Deployments["existing"].Token)
	for i := range count {
		alias := fmt.Sprintf("org-%d", i)
		require.Equal(t, "xaat-"+alias, cfg.Deployments[alias].Token)
	}
}

func TestAddDeploymentThroughSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks requires extra privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target.toml")
	link := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(target, nil, 0o600))
	require.NoError(t, os.Symlink(target, link))
	require.NoError(t, config.AddDeployment(link, "test-org", config.Deployment{Token: "xaat-new"}))
	actual, err := os.Readlink(link)
	require.NoError(t, err)
	require.Equal(t, target, actual)
	tree, err := toml.LoadFile(target)
	require.NoError(t, err)
	require.Equal(t, "xaat-new", tree.Get("deployments.test-org.token"))
}
