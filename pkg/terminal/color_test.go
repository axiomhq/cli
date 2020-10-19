// Derived from https://github.com/cli/cli/blob/trunk/pkg/iostreams/color_test.go.
package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestColorSchemeSuite struct {
	suite.Suite

	origEnv map[string]string
}

func TestColorScheme(t *testing.T) {
	suite.Run(t, new(TestColorSchemeSuite))
}

func (s *TestColorSchemeSuite) SetupTest() {
	s.origEnv = map[string]string{
		"NO_COLOR":       os.Getenv("NO_COLOR"),
		"CLICOLOR":       os.Getenv("CLICOLOR"),
		"CLICOLOR_FORCE": os.Getenv("CLICOLOR_FORCE"),
	}
}

func (s *TestColorSchemeSuite) TeardownTest() {
	for k, v := range s.origEnv {
		os.Setenv(k, v)
	}
}

func (s *TestColorSchemeSuite) TestEnvColorDisabled() {
	tests := []struct {
		name             string
		envNoColor       string
		envClicolor      string
		envClicolorForce string
		want             bool
	}{
		{
			name:             "pristine env",
			envNoColor:       "",
			envClicolor:      "",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "NO_COLOR enabled",
			envNoColor:       "1",
			envClicolor:      "",
			envClicolorForce: "",
			want:             true,
		},
		{
			name:             "CLICOLOR disabled",
			envNoColor:       "",
			envClicolor:      "0",
			envClicolorForce: "",
			want:             true,
		},
		{
			name:             "CLICOLOR enabled",
			envNoColor:       "",
			envClicolor:      "1",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "CLICOLOR_FORCE has no effect",
			envNoColor:       "",
			envClicolor:      "",
			envClicolorForce: "1",
			want:             false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			os.Setenv("NO_COLOR", tt.envNoColor)
			os.Setenv("CLICOLOR", tt.envClicolor)
			os.Setenv("CLICOLOR_FORCE", tt.envClicolorForce)

			s.Equal(tt.want, envColorDisabled())
		})
	}
}

func (s *TestColorSchemeSuite) TestEnvColorForced() {
	tests := []struct {
		name             string
		envNoColor       string
		envClicolor      string
		envClicolorForce string
		want             bool
	}{
		{
			name:             "pristine env",
			envNoColor:       "",
			envClicolor:      "",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "NO_COLOR enabled",
			envNoColor:       "1",
			envClicolor:      "",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "CLICOLOR disabled",
			envNoColor:       "",
			envClicolor:      "0",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "CLICOLOR enabled",
			envNoColor:       "",
			envClicolor:      "1",
			envClicolorForce: "",
			want:             false,
		},
		{
			name:             "CLICOLOR_FORCE enabled",
			envNoColor:       "",
			envClicolor:      "",
			envClicolorForce: "1",
			want:             true,
		},
		{
			name:             "CLICOLOR_FORCE disabled",
			envNoColor:       "",
			envClicolor:      "",
			envClicolorForce: "0",
			want:             false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			os.Setenv("NO_COLOR", tt.envNoColor)
			os.Setenv("CLICOLOR", tt.envClicolor)
			os.Setenv("CLICOLOR_FORCE", tt.envClicolorForce)

			s.Equal(tt.want, envColorForced())
		})
	}
}
