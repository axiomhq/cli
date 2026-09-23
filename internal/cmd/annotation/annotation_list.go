package annotation

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
	"github.com/axiomhq/cli/pkg/utils"
)

type listOptions struct {
	*cmdutil.Factory

	// Datasets to filter by.
	Datasets []string `survey:"datasets"`

	// Filter by start time.
	Start string `survey:"start-time"`

	// Filter by end time.
	End string `survey:"end-time"`

	// Format to output data in. Defaults to tabular output.
	Format string
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list [(-f|--format)=json|table] [(-d|--datasets) <datasets>] [--start-time <start-time>] [--end-time <end-time>]",
		Short: "List all annotations",

		Aliases: []string{"ls"},

		Example: heredoc.Doc(`
			# List all annotations:
			$ axiom annotation list

			# List the annotations of a dataset from the last 24 hours:
			$ axiom annotation list --datasets=http-logs --start-time=-24h
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Datasets, "datasets", "d", nil, "Filter by datasets")
	cmd.Flags().StringVarP(&opts.Start, "start-time", "", "", "Filter by start time - may also be now or a relative time eg: -24h")
	cmd.Flags().StringVarP(&opts.End, "end-time", "", "", "Filter by end time - may also be now or a relative time eg: -1h (defaults to now if --start-time is set)")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("datasets", cmdutil.DatasetCompletionFunc(f))
	_ = cmd.RegisterFlagCompletionFunc("start-time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("end-time", cmdutil.NoCompletion)

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	if opts.End != "" && opts.Start == "" {
		return cmdutil.NewFlagErrorf("--end-time requires --start-time")
	}

	now := time.Now()
	start, err := parseTime(opts.Start, now)
	if err != nil {
		return fmt.Errorf("invalid start time: %w", err)
	}
	end, err := parseTime(opts.End, now)
	if err != nil {
		return fmt.Errorf("invalid end time: %w", err)
	}
	if end.IsZero() && !start.IsZero() {
		end = now
	}

	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	annotations, err := client.Annotations.List(ctx, &axiom.AnnotationsFilter{
		Datasets: opts.Datasets,
		Start:    start,
		End:      end,
	})
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
		return iofmt.FormatToJSON(opts.IO.Out(), annotations, opts.IO.ColorEnabled())
	}

	if len(annotations) == 0 {
		fmt.Fprintln(opts.IO.Out(), "No annotations found.")
		return nil
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(_ io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing %s:\n\n", utils.Pluralize(cs, "annotation", len(annotations)))
			trb.AddField("ID", cs.Bold)
			trb.AddField("Type", cs.Bold)
			trb.AddField("Datasets", cs.Bold)
			trb.AddField("Title", cs.Bold)
			trb.AddField("Description", cs.Bold)
			trb.AddField("URL", cs.Bold)
			trb.AddField("Time", cs.Bold)
			trb.AddField("End time", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		annotation := annotations[k]

		trb.AddField(annotation.ID, nil)
		trb.AddField(annotation.Type, nil)
		trb.AddField(strings.Join(annotation.Datasets, ", "), nil)
		trb.AddField(annotation.Title, nil)
		trb.AddField(annotation.Description, nil)
		trb.AddField(annotation.URL, nil)
		trb.AddField(annotation.Time.Format(time.RFC1123), cs.Gray)
		if !annotation.EndTime.IsZero() {
			trb.AddField(annotation.EndTime.Format(time.RFC1123), cs.Gray)
		} else {
			trb.AddField("-", cs.Gray)
		}
	}

	return iofmt.FormatToTable(opts.IO, len(annotations), header, nil, contentRow)
}
