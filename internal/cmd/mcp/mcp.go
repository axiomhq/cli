package mcp

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/version"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/mcptools"
)

// NewCmd creates and returns the mcp command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the Axiom MCP server",
		Long: heredoc.Doc(`
			Run the Axiom MCP server that provides access to Axiom's query and
			dataset capabilities via the Model Context Protocol.

			The server communicates over stdin/stdout.
		`),

		Example: heredoc.Doc(`
			# Configure your client to run the command
			$ axiom mcp

			# Some clients might need an absolute path to the binary
			$ /opt/homebrew/bin/axiom mcp
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServer(cmd.Context(), f)
		},
	}
}

func runServer(ctx context.Context, f *cmdutil.Factory) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "axiom-cli",
		Title:   "Axiom CLI",
		Version: version.Release(),
	}, &mcp.ServerOptions{
		// TODO: Tell clients about envs.
		// Instructions: "TODO: Env info",
	})

	client, err := f.Client(ctx)
	if err != nil {
		return err
	}

	registerTools(server, client.Datasets)

	// TODO: Use f.IO.

	return server.Run(ctx, mcp.NewStdioTransport())
}

func registerTools(s *mcp.Server, client *axiom.DatasetsService) {
	// Dataset
	registerTool(s, mcptools.DatasetsListTool(client))

	// Query
	registerTool(s, mcptools.QueryTool(client))
}

func registerTool[TIn, TOut any](s *mcp.Server, t *mcptools.Tool[TIn, TOut]) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        t.Name,
		Title:       t.Title,
		Description: t.Description,
		// InputSchema populated by [AddTool].
		// OutputSchema populated by [AddTool].
	}, func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[TIn]) (*mcp.CallToolResultFor[TOut], error) {
		res, err := t.Handler(ctx, params.Arguments)
		if err != nil {
			return nil, err
		}
		return &mcp.CallToolResultFor[TOut]{
			StructuredContent: res,
		}, nil
	})
}
