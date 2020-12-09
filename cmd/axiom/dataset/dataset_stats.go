package dataset

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func newStatsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
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
			return runStats(cmd.Context(), f)
		},
	}

	return cmd
}

func runStats(ctx context.Context, f *cmdutil.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	progStop := f.IO.StartActivityIndicator()
	stats, err := client.Datasets.Stats(ctx)
	if err != nil {
		progStop()
		return err
	}
	progStop()

	pagerStop, err := f.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	cs := f.IO.ColorScheme()
	tp := terminal.NewTablePrinter(f.IO)

	if f.IO.IsStdoutTTY() {
		fmt.Fprintf(f.IO.Out(), "Showing statistics of all dataset:\n\n")
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
