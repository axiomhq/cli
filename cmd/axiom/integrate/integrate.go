package integrate

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewIntegrateCmd creates and returns the integrate command.
func NewIntegrateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrate <command>",
		Short: "Integrate Axiom into a project",
		Long: heredoc.Doc(`
			Integrate Axiom into an existing codebase.
			
			This will spin up the configurations needed to get started with
			Axiom in your existing project as fast as possible.
		`),

		Example: heredoc.Doc(`
			# Get kickstarted for Axiom log shipping on your vercel project
			$ axiom integrate vercel
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},
	}

	return cmd
}
