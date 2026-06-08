package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/cli/internal/config"
)

var configFile = `
active_deployment = "axiom-eu-west-1"

[deployments]

[deployments.app]
url = "axiom-app.aws.com"
token = "this-is-obviously-stupid"
org_id = "axiomers-dh71"

[deployments.axiom-eu-west-1]
url = "axiom-eu-west-1.aws.com"
token = "this-is-obviously-stupid"
org_id = ""

[deployments.axiom-eu-west-2]
url = "axiom-eu-west-2.aws.com"
token = "this-is-obviously-more-stupid"
org_id = ""
`

type TestConfigSuite struct {
	suite.Suite
}

func TestFileSystem(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}

func (s *TestConfigSuite) SetupTest() {
	s.Require().NoError(os.Unsetenv("AXIOM_DEPLOYMENT"))
}

// Make sure TOML configuration is properly loaded and the active deployment is
// overwritten from the environment.
func (s *TestConfigSuite) TestLoad() {
	s.Require().NoError(os.Setenv("AXIOM_DEPLOYMENT", "axiom-eu-west-2"))

	cfg, err := config.LoadFromReader(strings.NewReader(configFile))
	s.Require().NoError(err)
	s.Require().NotEmpty(cfg)

	s.Equal("axiom-eu-west-2", cfg.ActiveDeployment)
	s.Len(cfg.Deployments, 3)
}

// TestWriteFileMode pins the file mode of the on-disk configuration file at
// 0o600 so the embedded per-deployment tokens are not exposed to other local
// users. Covers both a freshly created file and a pre-existing wider-mode file
// left behind by older CLI versions.
func TestWriteFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file mode semantics do not apply on Windows")
	}

	newConfig := func(path string) *config.Config {
		return &config.Config{
			ActiveDeployment: "test",
			ConfigFilePath:   path,
			Deployments: map[string]config.Deployment{
				"test": {URL: "https://api.axiom.co", Token: "xaat-secret"},
			},
		}
	}

	t.Run("new file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".axiom.toml")
		if err := newConfig(path).Write(); err != nil {
			t.Fatalf("Write: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
			t.Errorf("mode = %o, want %o", got, want)
		}
	})

	t.Run("existing wider file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ".axiom.toml")
		if err := os.WriteFile(path, []byte("active_deployment = \"\"\n"), 0o600); err != nil {
			t.Fatalf("seed: %v", err)
		}
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatalf("seed chmod: %v", err)
		}
		if err := newConfig(path).Write(); err != nil {
			t.Fatalf("Write: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
			t.Errorf("mode = %o, want %o", got, want)
		}
	})
}
