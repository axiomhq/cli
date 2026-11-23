package root

// Derived from https://github.com/cli/cli/blob/trunk/pkg/cmd/root/help_topic.go

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

var topics = map[string]string{
	"credentials": `
		Personal access tokens can be retrieved from the user's profile page
		(Settings -> Profile).

		API tokens can be provided by flag or environment variable for ingestion,
		querying, annotations and dataset management, subject to their permissions.
		Login, auth status, token replacement and organization switching still
		require a personal access token.

		EDGE OPERATIONS: When using edge endpoints (--edge-url or --edge-region),
		ingest operations require an API token (xaat-*). Personal tokens are
		only supported for edge queries, not edge ingest.
	`,

	"environment": `
		AXIOM_DEPLOYMENT: The deployment to use. Overwrites the choice loaded
		from the configuration file.

		AXIOM_EDGE_REGION: Regional edge domain for ingest and query operations
		(e.g., "eu-west-1.aws.edge.axiom.co").

		AXIOM_EDGE_URL: Explicit edge URL for ingest and query operations. Takes
		precedence over AXIOM_EDGE_REGION if both are set.

		AXIOM_ORG_ID: The organization id of the organization the access token
		is valid for.

		AXIOM_PAGER, PAGER (in order of precedence): A terminal paging program
		to send standard output to, e.g. "less".

		AXIOM_TOKEN: Token The access token to use. Overwrites the choice loaded
		from the configuration file.

		AXIOM_URL: The deployment url to use. Overwrites the choice loaded from
		the configuration file.

		VISUAL, EDITOR (in order of precedence): The editor to use for authoring
		text.

		NO_COLOR: Set to any value to avoid printing ANSI escape sequences for
		color output.

		CLICOLOR: Set to "0" to disable printing ANSI colors in output.

		CLICOLOR_FORCE: Set to a value other than "0" to keep ANSI colors in
		output even when the output is piped.

		HTTP_PROXY, HTTPS_PROXY (in order of precedence): The URL of the proxy
		server to use for HTTP or HTTPS requests respectively. The values may be
		either a complete URL or a "host[:port]", with the "http" scheme assumed.
		Supported schemes are "http", "https", and "socks5".

		NO_PROXY: A comma-separated list of hostnames, domains, IP addresses, or CIDR
		notations that should bypass the proxy specified by HTTP_PROXY or HTTPS_PROXY.
		Entries may optionally include a port number, formatted as 'domain:port' or 'IP:port'.
	`,

	"exit-codes": `
		gh follows normal conventions regarding exit codes.
		
		- If a command completes successfully, the exit code will be 0
		
		- If a command fails for any reason, the exit code will be 1
		
		- If configuration fails to load, the exit code will be 2
		
		NOTE: It is possible that a particular command may have more exit codes,
		so it is a good practice to check documentation for the command if you
		are relying on exit codes to control some behavior. 
	`,
}

func newHelpTopic(io *terminal.IO, topic string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  topic,
		Long: io.Doc(topics[topic]),

		Hidden: true,

		Run: helpTopicHelpFunc,
	}

	cmd.SetHelpFunc(helpTopicHelpFunc)
	cmd.SetUsageFunc(helpTopicUsageFunc)

	return cmd
}

func helpTopicHelpFunc(cmd *cobra.Command, _ []string) {
	cmd.Print(cmd.Long)
}

func helpTopicUsageFunc(cmd *cobra.Command) error {
	cmd.Printf("Usage: axiom help %s", cmd.Use)
	return nil
}
