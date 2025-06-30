package mcptools

import (
	"context"

	"github.com/axiomhq/axiom-go/axiom"
)

// DatasetsListTool returns a tool that lists all datasets.
func DatasetsListTool(client *axiom.DatasetsService) *Tool[any, []*axiom.Dataset] {
	return &Tool[any, []*axiom.Dataset]{
		Name:        "datasets_list",
		Title:       "List Datasets",
		Description: "List all datasets",
		Handler: func(ctx context.Context, _ any) ([]*axiom.Dataset, error) {
			return client.List(ctx)
		},
	}
}
