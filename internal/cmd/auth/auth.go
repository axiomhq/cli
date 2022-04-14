package auth

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

// NewCmd creates and returns the auth command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Manage authentication state",
		Long:  `Login, logout and refresh authentication state.`,

		Example: heredoc.Doc(`
			$ axiom auth login
			$ axiom auth logout
			$ axiom auth select
			$ axiom auth status
			$ axiom auth switch-org
			$ axiom auth update-token
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},
	}

	cmd.AddCommand(newLoginCmd(f))
	cmd.AddCommand(newLogoutCmd(f))
	cmd.AddCommand(newSelectCmd(f))
	cmd.AddCommand(newStatusCmd(f))
	cmd.AddCommand(newSwitchOrgCmd(f))
	cmd.AddCommand(newUpdateTokenCmd(f))

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

func readTokenFromStdIn(r io.Reader) (string, error) {
	// The token won't be longer.
	r = io.LimitReader(r, 256)

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	token := strings.TrimSuffix(string(contents), "\n")
	token = strings.TrimSuffix(token, "\r")

	return token, nil
}
