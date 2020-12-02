package completion

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCompletionCmd creates and returns the completion command.
func NewCompletionCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <command>",
		Short: "Generate shell completion scripts",
		Long: heredoc.Doc(`
			Generate shell completion scripts for Axiom CLI commands.
			
			When installing Axiom CLI through a package manager, however, it's
			possible that no additional shell configuration is necessary to gain
			completion support. For Homebrew, see https://docs.brew.sh/Shell-Completion.
		`),

		Example: heredoc.Doc(`
			$ axiom completion bash
			$ axiom completion fish
			$ axiom completion powershell
			$ axiom completion zsh
		`),
	}

	cmd.AddCommand(newBashCmd(f))
	cmd.AddCommand(newFishCmd(f))
	cmd.AddCommand(newPowershellCmd(f))
	cmd.AddCommand(newZshCmd(f))

	return cmd
}
