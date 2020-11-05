package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type logoutOptions struct {
	*cmdutil.Factory

	// Alias of the backend to delete. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Alias string
	// Force the deleteion and skip the confirmation prompt.
	Force bool
}

func newLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &logoutOptions{
		Factory: f,
	}

	cmd := &cobra.Command{ //nolint:dupl
		Use:   "logout [<backend-alias>] [-f|--force]",
		Short: "Logout of an Axiom instance",

		DisableFlagsInUseLine: true,

		Args: cmdutil.ChainPositionalArgs(
			cobra.MaximumNArgs(1),
			cmdutil.PopulateFromArgs(f, &opts.Alias),
		),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the backend to log out of:
			$ axiom auth logout
			
			# Log out of a specified backend:
			$ axiom auth logout axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsBackends(f),
			cmdutil.NeedsValidBackend(f, &opts.Alias),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Alias == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Which backend to log out off?",
					Options: opts.Config.BackendAliases(),
				}, &opts.Alias, opts.IO.SurveyIO()); err != nil {
					return err
				}
			}

			return runLogout(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	return cmd
}

func runLogout(ctx context.Context, opts *logoutOptions) error {
	// Logging out must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Are you sure you want to log out of backend %q?", opts.Alias)
		if overwrite, err := surveyext.AskConfirm(msg, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	// TODO: Logout, I guess we need ctx in the here soon ;)
	_ = ctx

	time.Sleep(time.Second * 2)

	stop()

	if opts.IO.IsStdoutTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Logged out of backend %s\n",
			cs.SuccessIcon(), cs.Bold(opts.Alias))
	}

	delete(opts.Config.Backends, opts.Alias)
	if opts.Config.ActiveBackend == opts.Alias {
		opts.Config.ActiveBackend = ""
	}

	return opts.Config.Write()
}
