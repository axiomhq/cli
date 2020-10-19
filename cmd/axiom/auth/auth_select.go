package auth

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type selectOptions struct {
	*cmdutil.Factory

	// Alias of the backend to select. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Alias string
}

func newSelectCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &selectOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "select [<backend-alias>]",
		Short: "Select an Axiom instance",
		Long: heredoc.Doc(`
			Select an Axiom instance to use by default and persist the choice in
			the configuration file.
		`),

		DisableFlagsInUseLine: true,

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the backend to use by default:
			$ axiom auth select
			
			# Specify the backend to use by default:
			$ axiom auth select axiom-eu-west-1
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !opts.IO.IsStdinTTY() && len(args) == 0 {
				return cmdutil.ErrNoPromptArgRequired
			} else if len(args) == 1 {
				opts.Alias = args[0]
			}

			if err := cmdutil.ChainRunFuncs(
				cmdutil.NeedsBackends(f),
				cmdutil.NeedsValidBackend(f, opts.Alias),
			)(cmd, args); err != nil {
				return err
			}

			return completeSelect(opts)
		},

		RunE: func(_ *cobra.Command, _ []string) error {
			return runSelect(opts)
		},
	}

	return cmd
}

func completeSelect(opts *selectOptions) error {
	return survey.AskOne(&survey.Select{
		Message: "Which backend to select?",
		Options: opts.Config.BackendAliases(),
	}, &opts.Alias, opts.IO.SurveyIO())
}

func runSelect(opts *selectOptions) error {
	opts.Config.ActiveBackend = opts.Alias

	return opts.Config.Write()
}
