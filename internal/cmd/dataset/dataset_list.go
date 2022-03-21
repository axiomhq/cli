package dataset

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
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
		Use:   "list [(-f|--format)=json|table]",
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

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	datasets, err := client.Datasets.List(ctx)
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
		return iofmt.FormatToJSON(opts.IO.Out(), datasets, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing %s:\n\n", utils.Pluralize(cs, "dataset", len(datasets)))
			trb.AddField("Name", cs.Bold)
			trb.AddField("Description", cs.Bold)
			trb.AddField("Created", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		dataset := datasets[k]

		trb.AddField(dataset.Name, nil)
		trb.AddField(dataset.Description, nil)
		trb.AddField(dataset.CreatedAt.Format(time.RFC1123), cs.Gray)
	}

	return iofmt.FormatToTable(opts.IO, len(datasets), header, nil, contentRow)
}
