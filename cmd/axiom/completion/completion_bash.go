package completion

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newBashCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Generate shell completion script for bash",
		Long:  `Generate the autocompletion script for Axiom CLI for the bash shell.`,

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# To load completions in your current shell session:
			$ source <(axiom completion bash)
			
			# To load completions for every new session, execute once:
			# Linux:
			$ axiom completion bash > /etc/bash_completion.d/axiom
			# MacOS:
			$ axiom completion bash > /usr/local/etc/bash_completion.d/axiom
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Root().GenBashCompletion(f.IO.Out())
		},
	}

	return cmd
}
