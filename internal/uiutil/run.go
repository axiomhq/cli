package uiutil

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

// RunUIOrHelp runs the given model as a bubbletea program if a TTY is attached
// and it has color enabled. If not, the commands help is invoked.
func RunUIOrHelp(io *terminal.IO, model tea.Model, cmd *cobra.Command) error {
	if io.IsStdinTTY() && io.IsStdoutTTY() && io.IsStderrTTY() && io.ColorEnabled() {
		return tea.NewProgram(model).Start()
	}
	return cmd.Help()
}
