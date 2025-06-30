package mcptools

import (
	"context"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
)

// QueryToolParams represents the parameters for the query tool.
type QueryToolParams struct {
	// APL is the APL query to execute.
	APL string `json:"apl" jsonschema:"The APL query to execute"`
}

// QueryTool returns a tool that executes an APL query.
func QueryTool(client *axiom.DatasetsService) *Tool[QueryToolParams, *query.Result] {
	return &Tool[QueryToolParams, *query.Result]{
		Name:        "query",
		Title:       "Query",
		Description: "Execute an APL query",
		Handler: func(ctx context.Context, params QueryToolParams) (*query.Result, error) {
			return client.Query(ctx, params.APL)
		},
	}
}
