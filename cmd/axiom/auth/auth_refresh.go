package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type refreshOptions struct {
	*cmdutil.Factory

	// Alias of the backend to refresh authentication credentials for. If not
	// supplied as an argument, which is optional, the user will be asked for
	// it.
	Alias string
}

func newRefreshCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &refreshOptions{
		Factory: f,
	}

	cmd := &cobra.Command{ //nolint:dupl
		Use:   "refresh [<backend-alias>] [-f|--force]",
		Short: "Refresh authentication credentials of an Axiom instance",

		DisableFlagsInUseLine: true,

		Args: cmdutil.ChainPositionalArgs(
			cobra.MaximumNArgs(1),
			cmdutil.PopulateFromArgs(f, &opts.Alias),
		),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the backend to refresh the authentication credentials for:
			$ axiom auth refresh
			
			# Refresh authentication credentials for a specified backend:
			$ axiom auth refresh axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsBackends(f),
			cmdutil.NeedsValidBackend(f, &opts.Alias),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Alias == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Which backend to refresh the authentication credentials for?",
					Options: opts.Config.BackendAliases(),
				}, &opts.Alias, opts.IO.SurveyIO()); err != nil {
					return err
				}
			}

			return runRefresh(cmd.Context(), opts)
		},
	}

	return cmd
}

func runRefresh(ctx context.Context, opts *refreshOptions) error {
	stop := opts.IO.StartActivityIndicator()
	defer stop()

	// TODO: Refresh, I guess we need ctx in the here soon ;)
	_ = ctx

	time.Sleep(time.Second * 2)

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		backend := opts.Config.Backends[opts.Alias]
		fmt.Fprintf(opts.IO.ErrOut(), "%s Refreshed authentication credentials for %s @ %s (%s)\n",
			cs.SuccessIcon(), cs.Bold(backend.Username), cs.Bold(opts.Alias), backend.URL)
	}

	return opts.Config.Write()
}
