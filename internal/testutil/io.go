package testutil

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

// CommandIO returns IO for testing with the given command already configured to
// use the IO.
func CommandIO(cmd *cobra.Command) *terminal.IO {
	io := terminal.TestIO()
	cmd.SetIn(io.In())
	cmd.SetOut(io.Out())
	cmd.SetErr(io.ErrOut())
	return io
}
