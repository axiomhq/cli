package completion

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// NewCompletionCmd creates and returns the completion command.
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <command>",
		Short: "Generate shell completion scripts",
		Long: heredoc.Doc(`
			Generate shell completion scripts for Axiom CLI commands.

			When installing Axiom CLI through a package manager, however, it's
			possible that no additional shell configuration is necessary to gain
			completion support. For Homebrew, see https://docs.brew.sh/Shell-Completion
		`),
	}

	cmd.AddCommand(newBashCmd())
	cmd.AddCommand(newZshCmd())
	cmd.AddCommand(newFishCmd())
	cmd.AddCommand(newPowershellCmd())

	return cmd
}
