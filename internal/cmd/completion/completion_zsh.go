package completion

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newZshCmd(f *cmdutil.Factory) *cobra.Command {
	var completionNoDesc bool

	cmd := &cobra.Command{
		Use:   "zsh [--no-descriptions]",
		Short: "Generate shell completion script for zsh",
		Long:  `Generate the autocompletion script for Axiom CLI for the zsh shell.`,

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# To load completions in your current shell session:
			$ source <(axiom completion zsh)
			
			# To load completions for every new session, execute once:
			$ axiom completion zsh > "${fpath[1]}/_axiom"
		`),

		RunE: func(cmd *cobra.Command, args []string) error {
			if completionNoDesc {
				return cmd.Root().GenZshCompletionNoDesc(f.IO.Out())
			}
			return cmd.Root().GenZshCompletion(f.IO.Out())
		},
	}

	cmd.Flags().BoolVar(&completionNoDesc, "no-descriptions", false, "Disable completion descriptions")

	return cmd
}
