package auth

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newSelectCmd(f *cmdutil.Factory) *cobra.Command {
	var activeDeployment string

	cmd := &cobra.Command{
		Use:   "select [<alias>]",
		Short: "Select an Axiom configuration",
		Long: heredoc.Doc(`
			Select an Axiom configuration to use by default and persist the choice
			in the configuration file.
		`),

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &activeDeployment),
		ValidArgsFunction: deploymentCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Select the deployment to use by default:
			$ axiom auth select
			
			# Specify the deployment to use by default:
			$ axiom auth select axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, NewLoginCmd(f)),
			cmdutil.NeedsDeployments(f),
			cmdutil.NeedsValidDeployment(f, &activeDeployment),
		),

		RunE: func(*cobra.Command, []string) error {
			if activeDeployment == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Which deployment to select?",
					Options: f.Config.DeploymentAliases(),
				}, &f.Config.ActiveDeployment, f.IO.SurveyIO()); err != nil {
					return err
				}
			} else {
				f.Config.ActiveDeployment = activeDeployment
			}

			if err := f.Config.Write(); err != nil {
				return err
			}

			cs := f.IO.ColorScheme()
			fmt.Fprintf(f.IO.ErrOut(), "%s Now using deployment %s by default\n",
				cs.SuccessIcon(), cs.Bold(f.Config.ActiveDeployment))

			return nil
		},
	}

	return cmd
}
