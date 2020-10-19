package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/cli/internal/config"
)

var configFile = `
active_backend = "axiom-eu-west-1"

[backends]

[backends.axiom-eu-west-1]
url = "axiom-eu-west-1.aws.com"
username = "lukas@axiom.co"
password = "this-is-obviously-stupid"

[backends.axiom-eu-west-2]
url = "axiom-eu-west-2.aws.com"
username = "seif@axiom.co"
password = "this-is-obviously-more-stupid"
`

type TestConfigSuite struct {
	suite.Suite
}

func TestFileSystem(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}

func (s *TestConfigSuite) SetupTest() {
	os.Clearenv()
}

// Make sure TOML configuration is properly loaded and the active backend is
// overwritten from the environment.
func (s *TestConfigSuite) Load() {
	s.Require().NoError(os.Setenv("AXM_BACKEND", "axiom-eu-west-2"))

	cfg, err := config.LoadFromReader(strings.NewReader(configFile))
	s.Require().NoError(err)
	s.Require().NotEmpty(cfg)

	s.Equal("axiom-eu-west-2", cfg.ActiveBackend)
	s.Len(cfg.Backends, 2)
}

// Make sure the active backend is configured, if specified by the user.
func (s *TestConfigSuite) LoadActiveBackendNotConfigured() {
	s.Require().NoError(os.Setenv("AXM_BACKEND", "axiom-eu-west-3"))

	cfg, err := config.LoadFromReader(strings.NewReader(configFile))
	s.Require().Error(err)
	s.Require().NotEmpty(cfg)
}
