package version

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewVersionCmd creates and returns the version command.
func NewVersionCmd(f *cmdutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  `Print the version and build details.`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			if _, ok := f.Config.Deployments[f.Config.ActiveDeployment]; !ok {
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
			cmd.Printf("\nAxiom UI, release %s (%s)\n",
				deploymentVersion, cs.Bold(f.Config.ActiveDeployment))

			return nil
		},
	}

	return cmd
}
