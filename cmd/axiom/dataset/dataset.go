package dataset

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewDatasetCmd creates and returns the dataset command.
func NewDatasetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset <command>",
		Short: "Manage datasets",
		Long:  "Create, edit and delete datasets.",

		Example: heredoc.Doc(`
			$ axiom dataset create --name nginx-logs --description "All Nginx logs"
			$ axiom dataset list
			$ axiom dataset info nginx-logs
			$ axiom dataset update nginx-logs --description "Some Nginx logs"
			$ axiom dataset delete nginx-logs
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.NeedsActiveDeployment(f),
	}

	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	cmd.AddCommand(newInfoCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newStatsCmd(f))
	cmd.AddCommand(newUpdateCmd(f))

	return cmd
}

func getDatasetNames(ctx context.Context, f *cmdutil.Factory) ([]string, error) {
	client, err := f.Client()
	if err != nil {
		return nil, err
	}

	stop := f.IO.StartActivityIndicator()
	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		stop()
		return nil, err
	}
	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	return datasetNames, nil
}
