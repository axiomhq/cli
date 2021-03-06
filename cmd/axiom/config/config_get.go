package config

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Get a configuration value by specifying its key.`,

		DisableFlagsInUseLine: true,

		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: keyCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Get the url of a configured deployment:
			$ axiom config get "deployments.axiom-eu-west-1
			
			# Get the url of a deployment configured in the given configuration
			# file:
			$ axiom config get "deployments.axiom-eu-west-1 -C /etc/axiom/cli.toml
		`),

		RunE: func(_ *cobra.Command, args []string) error {
			val, err := f.Config.Get(args[0])
			if err != nil {
				return err
			} else if val != "" {
				_, err = fmt.Fprintf(f.IO.Out(), "%s\n", val)
				return err
			}

			return nil
		},
	}

	return cmd
}
