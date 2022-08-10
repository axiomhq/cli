package version

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the version command.
func NewCmd(f *cmdutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  `Print the version and build details.`,

		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(version)
		},
	}

	return cmd
}
