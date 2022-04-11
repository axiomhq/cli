package config

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type exportOptons struct {
	*cmdutil.Factory

	// Login of the user to export the configuration values for.
	Login string
	// Required for the export command to execute.
	Force bool
}

func newExportCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &exportOptons{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "export [-f|--force]",
		Short: "Export the configuration values for the current deployment.",
		Long:  `Export the configuration values AXIOM_URL, AXIOM_TOKEN and AXIOM_ORG_ID from the current deployment to the current terminal session.`,

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Login),
		ValidArgsFunction: keyCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Export the configuration values AXIOM_URL, AXIOM_TOKEN and AXIOM_ORG_ID from the current deployment to the current terminal session:
			$ axiom config export
		`),

		RunE: func(_ *cobra.Command, args []string) error {
			// check that the force flag was used
			if !opts.Force {
				_, err := f.IO.ErrOut().Write([]byte(`You can not export the configuration values without the --force flag.
If you are sure you want to export the configuration values, run the command again with the --force flag.
Ensure that you eval this command, without doing so, the configuration values will not be exported.
Be aware that this may overwrite the existing environment variables and may also print the values to the console.` + "\n"))
				if err != nil {
					return err
				}
				return nil
			}

			// fetch current deployment
			deployment, ok := f.Config.GetActiveDeployment()
			if !ok {
				return fmt.Errorf("no active deployment found")
			}

			// export environment variables
			fmt.Fprintf(f.IO.Out(), `export AXIOM_URL="`+deployment.URL+`"`+"\n")
			fmt.Fprintf(f.IO.Out(), `export AXIOM_TOKEN="`+deployment.Token+`"`+"\n")
			fmt.Fprintf(f.IO.Out(), `export AXIOM_ORG_ID="`+deployment.OrganizationID+`"`+"\n")

			if _, err := f.IO.ErrOut().Write([]byte("Environment Variables set. If they don't seem to be set, ensure that this command was ran in an eval statement.\n")); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Allow for the export of the configuration values")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	return cmd
}
