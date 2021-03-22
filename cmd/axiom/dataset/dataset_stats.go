package dataset

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type statsOptions struct {
	*cmdutil.Factory

	// Format to output data in. Defaults to tabular output.
	Format string
}

func newStatsCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &statsOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "stats [(-f|--format=)json]",
		Short: "Get statistics about all datasets",
		Long: heredoc.Doc(`
			Get statistics about all datasets.

			This operation is more expensive compared to calling "list". When
			just the dataset name, description or creation time are needed,
			"list" is preferred.
		`),

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Get statstics about all datasets:
			$ axiom dataset stats
		`),

		PreRunE: cmdutil.NeedsDatasets(f),

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStats(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "", "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", formatCompletion)

	return cmd
}

func runStats(ctx context.Context, opts *statsOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	stats, err := client.Datasets.Stats(ctx)
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
		return iofmt.FormatToJSON(opts.IO.Out(), stats, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing statistics of all dataset:\n\n")
			trb.AddField("Name", cs.Bold)
			trb.AddField("Events", cs.Bold)
			trb.AddField("Blocks", cs.Bold)
			trb.AddField("Fields", cs.Bold)
			trb.AddField("Ingested", cs.Bold)
			trb.AddField("Compressed", cs.Bold)
			trb.AddField("Min Time", cs.Bold)
			trb.AddField("Max Time", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		dataset := stats.Datasets[k]

		trb.AddField(dataset.Name, nil)
		trb.AddField(strconv.Itoa(int(dataset.NumEvents)), nil)
		trb.AddField(strconv.Itoa(int(dataset.NumBlocks)), nil)
		trb.AddField(strconv.Itoa(int(dataset.NumFields)), nil)
		trb.AddField(dataset.InputBytesHuman, cs.Green)
		trb.AddField(dataset.CompressedBytesHuman, cs.Green)
		trb.AddField(dataset.MinTime.Format(time.RFC1123), cs.Gray)
		trb.AddField(dataset.MaxTime.Format(time.RFC1123), cs.Gray)
	}

	footer := func(_ io.Writer, trb iofmt.TableRowBuilder) {
		trb.AddField("Sum", cs.Bold)
		trb.AddField(strconv.Itoa(int(stats.NumEvents)), cs.Bold)
		trb.AddField(strconv.Itoa(int(stats.NumBlocks)), cs.Bold)
		trb.AddField("", nil)
		trb.AddField(stats.InputBytesHuman, func(s string) string { return cs.Bold(cs.Green(s)) })
		trb.AddField(stats.CompressedBytesHuman, func(s string) string { return cs.Bold(cs.Green(s)) })
	}

	return iofmt.FormatToTable(opts.IO, len(stats.Datasets), header, footer, contentRow)
}
