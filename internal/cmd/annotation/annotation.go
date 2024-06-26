package annotation

import (
	"context"
	"sort"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the annotation command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "annotation <command>",
		Short: "Manage annotations",
		Long:  "Manage annotations.",

		Example: heredoc.Doc(`
			$ axiom annotation create --type=production-deployment --datasets=http-logs
			$ axiom annotation list
			$ axiom annotation update ann_123456789 --title="Production Deployment"
			$ axiom annotation delete ann_123456789
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
