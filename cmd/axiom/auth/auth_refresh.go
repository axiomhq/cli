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

	cmd := &cobra.Command{
		Use:   "refresh [<backend-alias>] [-f|--force]",
		Short: "Refresh authentication credentials of an axiom instance",

		DisableFlagsInUseLine: true,

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: backendCompletionFunc(f),

		Example: heredoc.Doc(`
			# Select the backend to refresh the authentication credentials for:
			$ axiom auth refresh

			# Refresh authentication credentials for a specified backend:
			$ axiom auth refresh axiom-eu-west-1
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !opts.IO.IsStdinTTY() && len(args) == 0 {
				return cmdutil.ErrNoPromptArgRequired
			} else if len(args) == 1 {
				opts.Alias = args[0]
			}

			if err := cmdutil.Needs(
				cmdutil.NeedsBackends(f),
				cmdutil.NeedsValidBackend(f, opts.Alias),
			)(cmd, args); err != nil {
				return err
			}

			return completeRefresh(opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRefresh(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeRefresh(opts *refreshOptions) error {
	if opts.Alias != "" {
		return nil
	}

	return survey.AskOne(&survey.Select{
		Message: "Which backend to refresh the authentication credentials for?",
		Options: opts.Config.BackendAliases(),
	}, &opts.Alias, opts.IO.SurveyIO())
}

func runRefresh(ctx context.Context, opts *refreshOptions) error {
	cs := opts.IO.ColorScheme()

	backend, ok := opts.Config.Backends[opts.Alias]
	if !ok {
		return fmt.Errorf("backend %s not configured", cs.Bold(opts.Alias))
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	// TODO: Refresh, I guess we need ctx in the here soon ;)
	_ = ctx

	time.Sleep(time.Second * 2)

	stop()

	if opts.IO.IsStderrTTY() {
		fmt.Fprintf(opts.IO.ErrOut(), "%s Refreshed authentication credentials for %s @ %s (%s)\n",
			cs.SuccessIcon(), cs.Bold(backend.Username), cs.Bold(opts.Alias), backend.URL)
	}

	return opts.Config.Write()
}
