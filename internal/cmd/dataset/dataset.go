package dataset

import (
	"context"
	"sort"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the dataset command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset <command>",
		Short: "Manage datasets",
		Long:  "Manage datasets.",

		Example: heredoc.Doc(`
			$ axiom dataset create --name=nginx-logs --description="All Nginx logs"
			$ axiom dataset list
			$ axiom dataset update nginx-logs --description="Some Nginx logs"
			$ axiom dataset delete nginx-logs
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, auth.NewLoginCmd(f)),
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
		),
	}

	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newTrimCmd(f))
	cmd.AddCommand(newUpdateCmd(f))

	return cmd
}

func getDatasetNames(ctx context.Context, f *cmdutil.Factory) ([]string, error) {
	client, err := f.Client(ctx)
	if err != nil {
		return nil, err
	}

	stop := f.IO.StartActivityIndicator()
	defer stop()

	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		return nil, err
	}

	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}
	sort.Strings(datasetNames)

	return datasetNames, nil
}
