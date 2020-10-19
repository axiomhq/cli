package auth

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
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
			$ axiom auth logout
		`),
	}

	cmd.AddCommand(newLoginCmd(f))
	cmd.AddCommand(newLogoutCmd(f))
	cmd.AddCommand(newRefreshCmd(f))
	cmd.AddCommand(newStatusCmd(f))

	return cmd
}

func backendCompletionFunc(f *cmdutil.Factory) cmdutil.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		res := make([]string, 0, len(f.Config.Backends))
		for _, backend := range f.Config.BackendAliases() {
			if strings.HasPrefix(backend, toComplete) {
				res = append(res, backend)
			}
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	}
}
