// Derived from https://github.com/cli/cli/blob/trunk/pkg/cmd/root/help_topic.go
package root

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

var topics = map[string]string{
	"environment": heredoc.Doc(`
		AXM_BACKEND: The backend to use by default. Overwrittes the choice
		loaded from the configuration file.
		
		AXM_PAGER, PAGER (in order of precedence): A terminal paging program to
		send standard output to, e.g. "less".

		VISUAL, EDITOR (in order of precedence): The editor to use for authoring
		text.

		NO_COLOR: Set to any value to avoid printing ANSI escape sequences for
		color output.

		CLICOLOR: Set to "0" to disable printing ANSI colors in output.

		CLICOLOR_FORCE: Set to a value other than "0" to keep ANSI colors in
		output even when the output is piped.
	`),
}

func newHelpTopic(topic string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  topic,
		Long: topics[topic],

		Hidden: true,

		Run: helpTopicHelpFunc,
	}

	cmd.SetHelpFunc(helpTopicHelpFunc)
	cmd.SetUsageFunc(helpTopicUsageFunc)

	return cmd
}

func helpTopicHelpFunc(cmd *cobra.Command, args []string) {
	cmd.Print(cmd.Long)
}

func helpTopicUsageFunc(cmd *cobra.Command) error {
	cmd.Printf("Usage: axiom help %s", cmd.Use)
	return nil
}
