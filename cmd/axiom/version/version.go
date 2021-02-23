package version

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

// NewVersionCmd creates and returns the version command.
func NewVersionCmd(f *cmdutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long: heredoc.Doc(`
			Print the version and build details.
			
			When an active deployment is configured with a Personal Access Token,
			its version will be fetched and printed as well.
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if activeDeployment, ok := f.Config.GetActiveDeployment(); !ok {
				cmd.Println(version)
				return nil
			} else if activeDeployment.TokenType != config.Personal {
				cmd.Println(version)
				return nil
			}

			client, err := f.Client()
			if err != nil {
				return err
			}

			stop := f.IO.StartActivityIndicator()
			deploymentVersion, err := client.Version.Get(cmd.Context())
			if err != nil {
				stop()
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
