package version

import (
	"github.com/spf13/cobra"
)

// NewVersionCmd creates and returns the version command.
func NewVersionCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  `Print the version and build details.`,

		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version)
		},
	}

	return cmd
}
