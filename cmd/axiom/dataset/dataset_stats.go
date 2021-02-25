package dataset

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/nwidger/jsoncolor"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
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
	stats, err := client.Datasets.Stats(ctx)
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

	if opts.Format == formatJSON {
		var enc interface {
			Encode(interface{}) error
		}
		if opts.IO.ColorEnabled() {
			enc = jsoncolor.NewEncoder(opts.IO.Out())
		} else {
			enc = json.NewEncoder(opts.IO.Out())
		}

		return enc.Encode(stats)
	}

	cs := opts.IO.ColorScheme()
	tp := terminal.NewTablePrinter(opts.IO)

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Showing statistics of all dataset:\n\n")
		tp.AddField("Name", cs.Bold)
		tp.AddField("Events", cs.Bold)
		tp.AddField("Blocks", cs.Bold)
		tp.AddField("Fields", cs.Bold)
		tp.AddField("Ingested", cs.Bold)
		tp.AddField("Compressed", cs.Bold)
		tp.AddField("Min Time", cs.Bold)
		tp.AddField("Max Time", cs.Bold)
		tp.EndRow()
		tp.AddField("", nil)
		tp.EndRow()
	}

	for _, dataset := range stats.Datasets {
		tp.AddField(dataset.Name, nil)
		tp.AddField(strconv.Itoa(int(dataset.NumEvents)), nil)
		tp.AddField(strconv.Itoa(int(dataset.NumBlocks)), nil)
		tp.AddField(strconv.Itoa(int(dataset.NumFields)), nil)
		tp.AddField(dataset.InputBytesHuman, cs.Green)
		tp.AddField(dataset.CompressedBytesHuman, cs.Green)
		tp.AddField(dataset.MinTime.Format(time.RFC1123), cs.Gray)
		tp.AddField(dataset.MaxTime.Format(time.RFC1123), cs.Gray)
		tp.EndRow()
	}

	tp.AddField("", nil)
	tp.EndRow()

	tp.AddField("Sum", cs.Bold)
	tp.AddField(strconv.Itoa(int(stats.NumEvents)), cs.Bold)
	tp.AddField(strconv.Itoa(int(stats.NumBlocks)), cs.Bold)
	tp.AddField("", nil)
	tp.AddField(stats.InputBytesHuman, func(s string) string { return cs.Bold(cs.Green(s)) })
	tp.AddField(stats.CompressedBytesHuman, func(s string) string { return cs.Bold(cs.Green(s)) })
	tp.EndRow()

	return tp.Render()
}
