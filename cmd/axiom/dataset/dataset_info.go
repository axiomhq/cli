package dataset

import (
	"context"
	"fmt"
	"strconv"

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
		Use:   "info <dataset-name>",
		Short: "Get info of a dataset",

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively get info of a dataset:
			$ axiom dataset info

			# Get info of a dataset and provide the dataset name as an argument:
			$ axiom dataset info nginx-logs
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.NeedsDatasets(f)(cmd, args); err != nil {
				return err
			}

			if !opts.IO.IsStdinTTY() && len(args) == 0 {
				return cmdutil.ErrNoPromptArgRequired
			} else if len(args) == 1 {
				opts.Name = args[0]
				return nil
			}
			return completeInfo(cmd.Context(), opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInfo(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeInfo(ctx context.Context, opts *infoOptions) error {
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

	progStop := opts.IO.StartProgressIndicator()
	defer progStop()

	dataset, err := client.Datasets.Info(ctx, opts.Name)
	if err != nil {
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
		fmt.Fprintf(opts.IO.Out(), "\nShowing info of dataset %s:\n\n", cs.Bold(dataset.DisplayName))
		tp.AddField("Number of Events", cs.Bold)
		tp.AddField("Number of Blocks", cs.Bold)
		tp.AddField("Number of Fields", cs.Bold)
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
	tp.AddField(dataset.MinTime, cs.Gray)
	tp.AddField(dataset.MaxTime, cs.Gray)
	tp.EndRow()

	return tp.Render()
}
