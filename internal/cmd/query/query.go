package query

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/araddon/dateparse"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type options struct {
	*cmdutil.Factory

	// Query to run. If not supplied as an argument, which is optional, the user
	// will be asked for it.
	Query string
	// StartTime of the query.
	StartTime string
	// EndTime of the query.
	EndTime string
	// TimestampFormat the timestamp is formatted in.
	TimestampFormat string
	// Format to output data in. Defaults to tabular output.
	Format string

	startTime time.Time
	endTime   time.Time
}

// NewCmd creates and returns the query command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "query [<apl-query>] [(-f|--format)=json|table] [--start-time <start-time>] [--end-time <end-time>] [--timestamp-format <timestamp-format>]",
		Short: "Query data using APL",
		Long: heredoc.Doc(`
			Query data from an Axiom dataset using APL, the Axiom Processing
			Language.

			The query range can be specified by specifying start and end time.
			The timestamp format can be configured by specifying a pattern with
			the reference date:

				Mon Jan 2 15:04:05 -0700 MST 2006

			Omitted elements in the pattern are treated as zero or one as
			applicable. See the Go reference documentation for examples:
			https://pkg.go.dev/time#pkg-constants
		`),

		DisableFlagsInUseLine: true,

		Args: cmdutil.PopulateFromArgs(f, &opts.Query),

		Example: heredoc.Doc(`
			# Query the "http-logs" dataset for logs with a 304 status code:
			$ axiom query "['http-logs'] | where response == 304"
			
			# Count all events in the "http-logs" dataset with a 404 status code:
			$ axiom query "['http-logs'] | where response == 404 | count"
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, auth.NewLoginCmd(f)),
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsDatasets(f),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := complete(opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")
	cmd.Flags().StringVar(&opts.StartTime, "start-time", "", "Start time of the query - may also be a relative time eg: -24h, -20m")
	cmd.Flags().StringVar(&opts.EndTime, "end-time", "", "End time of the query - may also be a relative time eg: -24h, -20m")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format used in the the timestamp field. Default uses a heuristic parser. Must be expressed using the reference time 'Mon Jan 2 15:04:05 -0700 MST 2006'")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("start-time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("end-time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("timestamp-format", cmdutil.NoCompletion)

	return cmd
}

func timeStrToTime(timeStr string, timestampFormat string) (time.Time, error) {
	if timestampFormat != "" {
		// parse the timestamp as absolute because we have a definitive format
		return time.Parse(timestampFormat, timeStr)
	}

	// try relative dates first
	duration, err := time.ParseDuration(timeStr)
	if err == nil {
		return time.Now().Add(duration), nil
	}

	// try absolute dates without format
	return dateparse.ParseAny(timeStr)
}

func complete(opts *options) (err error) {
	if ts := opts.StartTime; ts != "" {
		opts.startTime, err = timeStrToTime(ts, opts.TimestampFormat)
		if err != nil {
			return err
		}
	}

	if ts := opts.EndTime; ts != "" {
		opts.endTime, err = timeStrToTime(ts, opts.TimestampFormat)
		if err != nil {
			return err
		}
	}

	if opts.Query != "" {
		return nil
	}

	return survey.AskOne(&survey.Input{
		Message: "Which query to run?",
	}, &opts.Query, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	res, err := client.Query(ctx, opts.Query,
		query.SetStartTime(opts.startTime),
		query.SetEndTime(opts.endTime),
	)
	if err != nil {
		return err
	} else if res.Status.RowsMatched == 0 || len(res.Tables) == 0 || len(res.Tables[0].Columns) == 0 {
		return errors.New("query returned no results")
	}

	progStop()

	pagerStop, err := opts.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	cs := opts.IO.ColorScheme()

	headerText := cs.Bold(opts.Query)
	headerText += fmt.Sprintf(" processed in %s", cs.Gray(res.Status.ElapsedTime.String()))
	headerText = fmt.Sprintf("Result of query %s:\n\n", headerText)

	table := res.Tables[0]

	// Deal with JSON output format. It only works for non-aggregated results OR
	// an aggregated result which produces a single value (it has no groups).
	if opts.Format == iofmt.JSON.String() {
		if tableHasAggregation(table) {
			if len(table.Groups) > 1 || (len(table.Columns) > 1 && len(table.Columns[0]) > 1) {
				return errors.New("JSON output format is not supported for aggregated results with groups")
			}
			if opts.IO.IsStdoutTTY() {
				fmt.Fprint(opts.IO.Out(), headerText)
			}
			return iofmt.FormatToJSON(opts.IO.Out(), table.Columns[0][0], opts.IO.ColorEnabled())
		}

		if opts.IO.IsStdoutTTY() {
			fmt.Fprint(opts.IO.Out(), headerText)
		}

		for i := range len(table.Columns[0]) {
			if err = iofmt.FormatToJSON(opts.IO.Out(), tableRowAtIndex(table, i), opts.IO.ColorEnabled()); err != nil {
				return err
			}
		}

		return nil
	}

	// Deal with table output for a result with more than ten fields as it
	// wouldn't be readable in a table.
	if !tableHasAggregation(table) && len(table.Fields) > 10 {
		if opts.IO.IsStdoutTTY() {
			fmt.Fprint(opts.IO.Out(), headerText)
		}

		hasTimeField := slices.ContainsFunc(table.Fields, func(field query.Field) bool {
			return field.Name == "_time"
		})
		hasSysTimeField := slices.ContainsFunc(table.Fields, func(field query.Field) bool {
			return field.Name == "_sysTime"
		})

		for i := range len(table.Columns[0]) {
			row := tableRowAtIndex(table, i)
			if hasTimeField {
				ts, err := time.Parse(time.RFC3339Nano, fmt.Sprint(row["_time"]))
				if err != nil {
					return err
				}
				fmt.Fprintf(opts.IO.Out(), "%s\t", cs.Gray(ts.Format(time.RFC1123)))
				delete(row, "_time")
			}
			if hasSysTimeField {
				delete(row, "_sysTime")
			}
			if err = iofmt.FormatToJSON(opts.IO.Out(), row, opts.IO.ColorEnabled()); err != nil {
				return err
			}
		}

		return nil
	}

	// Deal with table output format for non-aggregated results.
	if !tableHasAggregation(table) {
		var header iofmt.HeaderBuilderFunc
		if opts.IO.IsStdoutTTY() {
			header = func(_ io.Writer, trb iofmt.TableRowBuilder) {
				fmt.Fprint(opts.IO.Out(), headerText)
				for _, field := range table.Fields {
					trb.AddField(field.Name, cs.Bold)
				}
			}
		}

		contentRow := func(trb iofmt.TableRowBuilder, k int) {
			for _, column := range table.Columns {
				trb.AddField(fmt.Sprint(column[k]), nil)
			}
		}

		return iofmt.FormatToTable(opts.IO, len(table.Columns[0]), header, nil, contentRow)
	}

	// Deal with table output format for aggregated results.
	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(_ io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprint(opts.IO.Out(), headerText)
			for _, field := range table.Fields {
				trb.AddField(field.Name, cs.Bold)
			}
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		for _, column := range table.Columns {
			trb.AddField(fmt.Sprint(column[k]), nil)
		}
	}

	var footer iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() && len(res.Tables) > 1 && res.Tables[1].Name == "_totals" {
		totalsTable := res.Tables[1]
		footer = func(_ io.Writer, trb iofmt.TableRowBuilder) {
			trb.AddField("Total", cs.Bold) // Account for the _time field present in the former table.
			for i, field := range totalsTable.Fields {
				var total string
				if field.Aggregation != nil {
					total = fmt.Sprint(totalsTable.Columns[i][0])
				}
				trb.AddField(total, nil)
			}
		}
	}

	return iofmt.FormatToTable(opts.IO, len(table.Columns[0]), header, footer, contentRow)
}

func tableHasAggregation(table query.Table) bool {
	for _, field := range table.Fields {
		if field.Aggregation != nil {
			return true
		}
	}
	return false
}

func tableRowAtIndex(table query.Table, rowIdx int) map[string]any {
	row := make(map[string]any, len(table.Fields))
	for i, field := range table.Fields {
		row[field.Name] = table.Columns[i][rowIdx]
	}
	return row
}
