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
		Use:   "list [(-f|--format)=json|table] [(-d|--datasets) <datasets>] [(-start-time) <start-time>] [(--end-time) <end-time>",
		Short: "List all annotations",

		Aliases: []string{"ls"},

		Example: heredoc.Doc(`
			# List all annotations:
			$ axiom annotations list
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringArrayVarP(&opts.Datasets, "datasets", "d", nil, "Filter by datasets")
	cmd.Flags().StringVarP(&opts.Start, "start-time", "", "", "Filter by start time")
	cmd.Flags().StringVarP(&opts.End, "end-time", "", "", "Filter by end time")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("datasets", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("start-time", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("end-time", cmdutil.FormatCompletion)

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	var start, end *time.Time
	if opts.Start != "" {
		startVal, err := time.Parse(time.RFC3339, opts.Start)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
		start = &startVal
	}
	if opts.End != "" {
		endVal, err := time.Parse(time.RFC3339, opts.End)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
		end = &endVal
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	var filter *axiom.AnnotationsFilter
	if len(opts.Datasets) > 0 || start != nil || end != nil {
		filter = &axiom.AnnotationsFilter{
			Datasets: opts.Datasets,
			Start:    start,
			End:      end,
		}
	}

	annotations, err := client.Annotations.List(ctx, filter)
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
