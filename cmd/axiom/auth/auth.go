package auth

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

// NewAuthCmd creates and returns the auth command.
func NewAuthCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Manage authentication state",
		Long:  `Login, logout and refresh authentication state.`,

		Example: heredoc.Doc(`
			$ axiom auth login
			$ axiom auth status
			$ axiom auth select
			$ axiom auth logout
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},
	}

	cmd.AddCommand(newLoginCmd(f))
	cmd.AddCommand(newLogoutCmd(f))
	cmd.AddCommand(newSelectCmd(f))
	cmd.AddCommand(newStatusCmd(f))

	return cmd
}

func deploymentCompletionFunc(config *config.Config) cmdutil.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Just complete the first argument.
		if len(args) > 0 {
			return cmdutil.NoCompletion(cmd, args, toComplete)
		}

		aliases := config.DeploymentAliases()
		res := make([]string, 0, len(aliases))
		for _, alias := range aliases {
			if strings.HasPrefix(alias, toComplete) {
				res = append(res, alias)
			}
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	}
}
