package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
	"github.com/axiomhq/cli/pkg/utils"
)

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all datasets",

		Aliases: []string{"ls"},

		Example: heredoc.Doc(`
			# List all datasets:
			$ axiom dataset list
		`),

		PreRunE: cmdutil.NeedsDatasets(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), f)
		},
	}

	return cmd
}

func runList(ctx context.Context, f *cmdutil.Factory) error {
	client, err := f.Client()
	if err != nil {
		return err
	}

	progStop := f.IO.StartActivityIndicator()
	datasets, err := client.Datasets.List(ctx)
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
		fmt.Fprintf(f.IO.Out(), "Showing %s:\n\n", utils.Pluralize(cs, "dataset", len(datasets)))
		tp.AddField("Name", cs.Bold)
		tp.AddField("Description", cs.Bold)
		tp.AddField("Created", cs.Bold)
		tp.EndRow()
		tp.AddField("", nil)
		tp.EndRow()
	}

	for _, dataset := range datasets {
		tp.AddField(dataset.Name, nil)
		tp.AddField(dataset.Description, nil)
		tp.AddField(dataset.Created.Format(time.RFC1123), cs.Gray)
		tp.EndRow()
	}
	return tp.Render()
}
