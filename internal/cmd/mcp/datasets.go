package mcp

import (
	"context"
	"errors"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDatasetTools(srv *server.MCPServer, client *axiom.Client) {
	registerTool(srv, client, datasetListTool)
	registerTool(srv, client, datasetCreateTool)
	registerTool(srv, client, datasetDeleteTool)
}

func datasetListTool() (mcp.Tool, handler) {
	tool := mcp.NewTool("dataset-list",
		mcp.WithDescription("List datasets"),
	)
	return tool, func(ctx context.Context, client *axiom.Client, _ map[string]any) (any, error) {
		return client.Datasets.List(ctx)
	}
}

func datasetCreateTool() (mcp.Tool, handler) {
	tool := mcp.NewTool("dataset-create",
		mcp.WithDescription("Create a dataset"),
		mcp.WithString("name",
			mcp.Description("Name of the dataset"),
			mcp.Required(),
		),
		mcp.WithString("description",
			mcp.Description("Description of the dataset"),
		),
	)
	return tool, func(ctx context.Context, client *axiom.Client, args map[string]any) (any, error) {
		name, ok := args["name"].(string)
		if !ok || name == "" {
			return nil, errors.New("name must be provided")
		}
		description, _ := args["description"].(string)

		return client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
			Name:        name,
			Description: description,
		})
	}
}

func datasetDeleteTool() (mcp.Tool, handler) {
	tool := mcp.NewTool("dataset-delete",
		mcp.WithDescription("Delete a dataset"),
		mcp.WithString("name",
			mcp.Description("Name of the dataset"),
			mcp.Required(),
		),
	)
	return tool, func(ctx context.Context, client *axiom.Client, args map[string]any) (any, error) {
		name, ok := args["name"].(string)
		if !ok || name == "" {
			return nil, errors.New("name must be provided")
		}

		return nil, client.Datasets.Delete(ctx, name)
	}
}
