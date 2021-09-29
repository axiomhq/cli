package completion

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newPowershellCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generate shell completion script for powershell",
		Long:  `Generate the autocompletion script for Axiom CLI for powershell.`,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Root().GenPowerShellCompletion(f.IO.Out())
		},
	}

	return cmd
}
