package dataset

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

type infoOptions struct {
	*cmdutil.Factory

	// Name of the dataset to fetch info of. If not supplied as an argument,
	// which is optional, the user will be asked for it.
	Name string
}

func newInfoCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &infoOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "info [<dataset-name>]",
		Short: "Get info about a dataset",

		Args:              cmdutil.PopulateFromArgs(f, &opts.Name),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively get info of a dataset:
			$ axiom dataset info
			
			# Get info of a dataset and provide the dataset name as an argument:
			$ axiom dataset info nginx-logs
		`),

		PreRunE: cmdutil.NeedsDatasets(f),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeInfo(cmd.Context(), opts); err != nil {
				return err
			}
			return runInfo(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeInfo(ctx context.Context, opts *infoOptions) error {
	if opts.Name != "" {
		return nil
	}

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to get info for?",
		Options: datasetNames,
	}, &opts.Name, opts.IO.SurveyIO())
}

func runInfo(ctx context.Context, opts *infoOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	dataset, err := client.Datasets.Info(ctx, opts.Name)
	if err != nil {
		progStop()
		return err
	}
	progStop()

	pagerStop, err := opts.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	cs := opts.IO.ColorScheme()
	tp := terminal.NewTablePrinter(opts.IO)

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Showing info of dataset %s:\n\n", cs.Bold(dataset.Name))
		tp.AddField("Events", cs.Bold)
		tp.AddField("Blocks", cs.Bold)
		tp.AddField("Fields", cs.Bold)
		tp.AddField("Ingested Bytes", cs.Bold)
		tp.AddField("Compressed Bytes", cs.Bold)
		tp.AddField("Min Time", cs.Bold)
		tp.AddField("Max Time", cs.Bold)
		tp.EndRow()
		tp.AddField("", nil)
		tp.EndRow()
	}

	tp.AddField(strconv.Itoa(int(dataset.NumEvents)), nil)
	tp.AddField(strconv.Itoa(int(dataset.NumBlocks)), nil)
	tp.AddField(strconv.Itoa(int(dataset.NumFields)), nil)
	tp.AddField(dataset.InputBytesHuman, nil)
	tp.AddField(dataset.CompressedBytesHuman, nil)
	tp.AddField(dataset.MinTime.Format(time.RFC1123), cs.Gray)
	tp.AddField(dataset.MaxTime.Format(time.RFC1123), cs.Gray)
	tp.EndRow()

	return tp.Render()
}
