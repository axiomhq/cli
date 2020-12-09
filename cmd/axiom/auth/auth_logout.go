package auth

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type logoutOptions struct {
	*cmdutil.Factory

	// Alias of the deployment to delete. If not supplied as a flag, which is
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
		Use:   "logout [<deployment-alias>] [-f|--force]",
		Short: "Logout of an Axiom deployment",

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Alias),
		ValidArgsFunction: deploymentCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the deployment to log out of:
			$ axiom auth logout
			
			# Log out of a specified deployment:
			$ axiom auth logout axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsDeployments(f),
			cmdutil.NeedsValidDeployment(f, &opts.Alias),
		),

		RunE: func(_ *cobra.Command, _ []string) error {
			if opts.Alias == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Which deployment to log out off?",
					Options: opts.Config.DeploymentAliases(),
				}, &opts.Alias, opts.IO.SurveyIO()); err != nil {
					return err
				}
			}

			return runLogout(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func runLogout(opts *logoutOptions) error {
	// Logging out must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Are you sure you want to log out of deployment %q?", opts.Alias)
		if overwrite, err := surveyext.AskConfirm(msg, true, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	delete(opts.Config.Deployments, opts.Alias)
	if opts.Config.ActiveDeployment == opts.Alias {
		opts.Config.ActiveDeployment = ""
	}

	if opts.IO.IsStdoutTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Logged out of deployment %s\n",
			cs.SuccessIcon(), cs.Bold(opts.Alias))
	}

	return opts.Config.Write()
}
