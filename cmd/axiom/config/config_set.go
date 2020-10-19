package config

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newSetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  `Set a configuration value by specifying its key and value.`,

		DisableFlagsInUseLine: true,

		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: keyCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Set the url for a configured backend:
			$ axiom config set "backends.my-axiom" "https://my-axiom.eu-west-1.aws.com"

			# Set the url for a backend configured in the given configuration
			# file:
			$ axiom config set "backends.my-axiom" "https://my-axiom.eu-west-1.aws.com" -C /etc/axiom/cli.toml
		`),

		RunE: func(_ *cobra.Command, args []string) error {
			return f.Config.Set(args[0], args[1])
		},

		PostRunE: func(_ *cobra.Command, _ []string) error {
			return f.Config.Write()
		},
	}

	return cmd
}
