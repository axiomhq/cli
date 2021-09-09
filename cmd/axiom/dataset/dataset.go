package dataset

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

// NewDatasetCmd creates and returns the dataset command.
func NewDatasetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset <command>",
		Short: "Manage datasets",
		Long:  "Manage datasets.",

		Example: heredoc.Doc(`
			$ axiom dataset create --name=nginx-logs --description="All Nginx logs"
			$ axiom dataset list
			$ axiom dataset info nginx-logs
			$ axiom dataset update nginx-logs --description="Some Nginx logs"
			$ axiom dataset delete nginx-logs
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
		),
	}

	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	cmd.AddCommand(newInfoCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newStatsCmd(f))
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

	return datasetNames, nil
}

func formatCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(iofmt.Formats()))
	for _, validFormat := range iofmt.Formats() {
		if strings.HasPrefix(validFormat.String(), toComplete) {
			res = append(res, validFormat.String())
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}
