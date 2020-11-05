package auth

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newSelectCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select [<backend-alias>]",
		Short: "Select an Axiom instance",
		Long: heredoc.Doc(`
			Select an Axiom instance to use by default and persist the choice in
			the configuration file.
		`),

		DisableFlagsInUseLine: true,

		Args: cmdutil.ChainPositionalArgs(
			cobra.MaximumNArgs(1),
			cmdutil.PopulateFromArgs(f, &f.Config.ActiveBackend),
		),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the backend to use by default:
			$ axiom auth select
			
			# Specify the backend to use by default:
			$ axiom auth select axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsBackends(f),
			cmdutil.NeedsValidBackend(f, &f.Config.ActiveBackend),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			if f.Config.ActiveBackend == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Which backend to select?",
					Options: f.Config.BackendAliases(),
				}, &f.Config.ActiveBackend, f.IO.SurveyIO()); err != nil {
					return err
				}
			}

			if err := f.Config.Write(); err != nil {
				return err
			}

			cs := f.IO.ColorScheme()
			fmt.Fprintf(f.IO.ErrOut(), "%s Now using backend %s by default\n",
				cs.SuccessIcon(), cs.Bold(f.Config.ActiveBackend))

			return nil
		},
	}

	return cmd
}
