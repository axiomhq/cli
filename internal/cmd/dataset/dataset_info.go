package dataset

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type infoOptions struct {
	*cmdutil.Factory

	// Name of the dataset to fetch info of. If not supplied as an argument,
	// which is optional, the user will be asked for it.
	Name string
	// Format to output data in. Defaults to tabular output.
	Format string
}

func newInfoCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &infoOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "info [<dataset-name>] [(-f|--format)=json|table]",
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

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func completeInfo(ctx context.Context, opts *infoOptions) error {
	if opts.Name != "" {
		return nil
	}

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	} else if len(datasetNames) == 1 {
		opts.Name = datasetNames[0]
		return nil
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to get info for?",
		Options: datasetNames,
	}, &opts.Name, opts.IO.SurveyIO())
}

func runInfo(ctx context.Context, opts *infoOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
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

	if opts.Format == iofmt.JSON.String() {
		return iofmt.FormatToJSON(opts.IO.Out(), dataset, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing info of dataset %s:\n\n", cs.Bold(dataset.Name))
			trb.AddField("Events", cs.Bold)
			trb.AddField("Blocks", cs.Bold)
			trb.AddField("Fields", cs.Bold)
			trb.AddField("Ingested Bytes", cs.Bold)
			trb.AddField("Compressed Bytes", cs.Bold)
			trb.AddField("Min Time", cs.Bold)
			trb.AddField("Max Time", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, _ int) {
		trb.AddField(strconv.Itoa(int(dataset.NumEvents)), nil)
		trb.AddField(strconv.Itoa(int(dataset.NumBlocks)), nil)
		trb.AddField(strconv.Itoa(int(dataset.NumFields)), nil)
		trb.AddField(dataset.InputBytesHuman, nil)
		trb.AddField(dataset.CompressedBytesHuman, nil)
		trb.AddField(dataset.MinTime.Format(time.RFC1123), cs.Gray)
		trb.AddField(dataset.MaxTime.Format(time.RFC1123), cs.Gray)
	}

	return iofmt.FormatToTable(opts.IO, 1, header, nil, contentRow)
}
