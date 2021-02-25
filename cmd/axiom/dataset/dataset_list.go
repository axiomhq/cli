package dataset

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/nwidger/jsoncolor"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
	"github.com/axiomhq/cli/pkg/utils"
)

type listOptions struct {
	*cmdutil.Factory

	// Format to output data in. Defaults to tabular output.
	Format string
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list [(-f|--format=)json]",
		Short: "List all datasets",

		Aliases: []string{"ls"},

		Example: heredoc.Doc(`
			# List all datasets:
			$ axiom dataset list
		`),

		PreRunE: cmdutil.NeedsDatasets(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "", "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", formatCompletion)

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	datasets, err := client.Datasets.List(ctx)
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

		return enc.Encode(datasets)
	}

	cs := opts.IO.ColorScheme()
	tp := terminal.NewTablePrinter(opts.IO)

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Showing %s:\n\n", utils.Pluralize(cs, "dataset", len(datasets)))
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
