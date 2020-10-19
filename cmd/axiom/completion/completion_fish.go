package completion

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newFishCmd(f *cmdutil.Factory) *cobra.Command {
	var completionNoDesc bool

	cmd := &cobra.Command{
		Use:   "fish [--no-descriptions]",
		Short: "Generate shell completion script for fish",
		Long: heredoc.Doc(`
			Generate the autocompletion script for Axiom CLI for the fish shell.

			You will need to start a new shell for this setup to take effect.
		`),

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# To load completions in your current shell session:
			$ axiom completion fish | source
			
			# To load completions for every new session, execute once:
			$ axiom completion fish > ~/.config/fish/completions/axiom.fish
		`),

		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			return cmd.Root().GenFishCompletion(f.IO.Out(), !completionNoDesc)
		},
	}

	cmd.Flags().BoolVar(&completionNoDesc, "no-descriptions", false, "Disable completion descriptions")

	return cmd
}
