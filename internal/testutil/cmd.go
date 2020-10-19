package testutil

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// SetupCmd applies default completion and the test IO to a command. The
// completion defaults are applied recursively to all commands.
func SetupCmd(cmd *cobra.Command) {
	_ = CommandIO(cmd)
	cmdutil.DefaultCompletion(cmd)
	cmdutil.InheritRootPersistenPreRun(cmd)

	// Do what cmdutil.DefaultCompletion() does but apply to the cmd as well
	// because the method treats cmd as the root command.
	if cmd.Args == nil {
		cmd.Args = cobra.NoArgs
	}
}
