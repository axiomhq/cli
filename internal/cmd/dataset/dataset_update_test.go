package dataset

import (
	"strings"
	"testing"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func TestUpdateRequiresDescriptionWhenNotInteractive(t *testing.T) {
	cmd := newUpdateCmd(&cmdutil.Factory{Config: &config.Config{}, IO: terminal.TestIO()})
	cmd.SetArgs([]string{"logs"})

	testutil.SetupCmd(cmd)

	_, err := cmd.ExecuteC()
	if err == nil || !strings.Contains(err.Error(), `required flag(s) "description"`) {
		t.Errorf("error = %v, want the required description flag", err)
	}
}
