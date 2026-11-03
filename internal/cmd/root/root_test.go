package root

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestNewCmd_FlagDefaultsOmitEnvironment(t *testing.T) {
	const value = "value-from-environment"
	for _, name := range []string{
		"AXIOM_ORG_ID",
		"AXIOM_TOKEN",
		"AXIOM_URL",
		"AXIOM_EDGE_URL",
		"AXIOM_EDGE_REGION",
	} {
		t.Setenv(name, value)
	}

	cmd := NewCmd(&cmdutil.Factory{IO: terminal.TestIO()})

	cmd.PersistentFlags().VisitAll(func(fl *pflag.Flag) {
		assert.NotContains(t, fl.DefValue, value, "flag %q", fl.Name)
	})
}
