package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the web command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run a MCP server",
		Long:  `Run a MCP server that is aware of the current configuration.`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			srv := server.NewMCPServer(
				"Axiom🚀",
				version.Release(),
			)

			client, err := f.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			registerDatasetTools(srv, client)

			iosrv := server.NewStdioServer(srv)
			iosrv.SetErrorLogger(log.New(f.IO.ErrOut(), "", log.LstdFlags))

			if err := iosrv.Listen(cmd.Context(), f.IO.In(), f.IO.Out()); err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}

			return nil
		},
	}
}

type handler func(context.Context, *axiom.Client, map[string]any) (any, error)

func registerTool(srv *server.MCPServer, client *axiom.Client, f func() (mcp.Tool, handler)) {
	tool, handler := f()
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := handler(ctx, client, req.Params.Arguments)
		if err != nil {
			return nil, err
		}

		v, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(v)), nil
	})
}
