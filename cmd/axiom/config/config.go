package config

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

// NewConfigCmd creates and returns the config command.
func NewConfigCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage configuration",
		Long:  `Get or set values in the configuration or edit the file in an editor.`,

		Example: heredoc.Doc(`
			# Get the url of a configured deployment:
			$ axiom config get "deployments.axiom-eu-west-1"

			# Set the url for a configured deployment:
			$ axiom config set "deployments.axiom-eu-west-1" "https://axiom.eu-west-1.aws.com"

			# Open the configuration file in the configured editor:
			$ axiom config edit
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},
	}

	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newGetCmd(f))
	cmd.AddCommand(newSetCmd(f))

	return cmd
}

func keyCompletionFunc(config *config.Config) cmdutil.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Just complete the first argument.
		if len(args) > 0 {
			return cmdutil.NoCompletion(cmd, args, toComplete)
		}

		keys := config.Keys()
		res := make([]string, 0, len(keys))
		for _, key := range keys {
			if strings.HasPrefix(key, toComplete) {
				res = append(res, key)
			}
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	}
}
