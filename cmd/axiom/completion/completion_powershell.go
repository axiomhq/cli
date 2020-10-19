package completion

import (
	"github.com/spf13/cobra"
)

func newPowershellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generate shell completion script for powershell",
		Long:  `Generate the autocompletion script for Axiom CLI for powershell.`,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
		},
	}

	return cmd
}
