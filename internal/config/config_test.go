package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/axiomhq/cli/internal/config"
)

var configFile = `
active_deployment = "axiom-eu-west-1"

[deployments]

[deployments.axiom-eu-west-1]
url = "axiom-eu-west-1.aws.com"
token = "this-is-obviously-stupid"
token_type = "personal"

[deployments.axiom-eu-west-2]
url = "axiom-eu-west-2.aws.com"
token = "this-is-obviously-more-stupid"
token_type = "ingest"
`

type TestConfigSuite struct {
	suite.Suite
}

func TestFileSystem(t *testing.T) {
	suite.Run(t, new(TestConfigSuite))
}

func (s *TestConfigSuite) SetupTest() {
	s.Require().NoError(os.Unsetenv("AXM_DEPLOYMENT"))
}

// Make sure TOML configuration is properly loaded and the active deployment is
// overwritten from the environment.
func (s *TestConfigSuite) TestLoad() {
	s.Require().NoError(os.Setenv("AXM_DEPLOYMENT", "axiom-eu-west-2"))

	cfg, err := config.LoadFromReader(strings.NewReader(configFile))
	s.Require().NoError(err)
	s.Require().NotEmpty(cfg)

	s.Equal("axiom-eu-west-2", cfg.ActiveDeployment)
	s.Len(cfg.Deployments, 2)
}
