package config

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type exportOptions struct {
	*cmdutil.Factory

	// Required for the export command to execute.
	Force bool
}

var forceMessage = heredoc.Doc(`
	You can not export the configuration values without the --force flag.
	If you are sure you want to export the configuration values, run the command
	again with the --force flag. Ensure that you eval this command, without
	doing so, the configuration values will not be exported. Be aware that this
	may overwrite the existing environment variables and may also print the
	values to the console.
`)

var successMessage = heredoc.Doc(`
	Environment Variables set. If they don't seem to be set, ensure that this
	command was ran in an eval statement.
`)

func newExportCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &exportOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "export [-f|--force]",
		Short: "Export the configuration values",
		Long: heredoc.Doc(`
			Export the values AXIOM_URL, AXIOM_TOKEN and AXIOM_ORG_ID for the
			active configuration to the current terminal session.
		`),

		DisableFlagsInUseLine: true,

		ValidArgsFunction: keyCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Export the values AXIOM_URL, AXIOM_TOKEN and AXIOM_ORG_ID for the
			# current configuration to the current terminal session:
			$ eval $(axiom config export --force)
		`),

		RunE: func(_ *cobra.Command, args []string) error {
			// check that the force flag was used
			if !opts.Force {
				fmt.Fprintf(opts.IO.ErrOut(), "%s %s\n", opts.IO.ColorScheme().ErrorIcon(), forceMessage)
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

			fmt.Fprintf(opts.IO.ErrOut(), "%s %s\n", opts.IO.ColorScheme().SuccessIcon(), successMessage)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Allow for the export of the configuration values")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	return cmd
}
