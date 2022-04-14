package version

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the version command.
func NewCmd(f *cmdutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long: heredoc.Doc(`
			Print the version and build details.
			
			When an active deployment is configured with a Personal Access Token,
			its version will be fetched and printed as well.
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if dep, ok := f.Config.GetActiveDeployment(); !ok || !axiom.IsPersonalToken(dep.Token) {
				cmd.Println(version)
				return nil
			}

			client, err := f.Client(cmd.Context())
			if err != nil {
				return err
			}

			stop := f.IO.StartActivityIndicator()
			defer stop()

			deploymentVersion, err := client.Version.Get(cmd.Context())
			if err != nil {
				return err
			}

			stop()

			cs := f.IO.ColorScheme()

			cmd.Println(version)

			if f.Config.ActiveDeployment == "" {
				cmd.Printf("\nAxiom, release %s\n",
					deploymentVersion)
			} else {
				cmd.Printf("\nAxiom, release %s (%s)\n",
					deploymentVersion, cs.Bold(f.Config.ActiveDeployment))
			}

			return nil
		},
	}

	return cmd
}
