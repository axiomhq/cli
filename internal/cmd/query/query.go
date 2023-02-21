package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/araddon/dateparse"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/nwidger/jsoncolor"
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

		RunE: func(cmd *cobra.Command, args []string) error {
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
		// parse the timestamp as absoute because we have a definitive format
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

	res, err := client.Datasets.Query(ctx, opts.Query,
		query.SetStartTime(opts.startTime),
		query.SetEndTime(opts.endTime),
	)
	if err != nil {
		return err
	} else if len(res.Matches) == 0 && len(res.Buckets.Totals) == 0 {
		return errors.New("query returned no results")
	}

	progStop()

	pagerStop, err := opts.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	// Deal with JSON output format.
	if opts.Format == iofmt.JSON.String() {
		var data any
		if len(res.Matches) > 0 {
			data = res.Matches
		} else if len(res.Buckets.Totals) > 0 {
			// For ungrouped buckets we just return the aggregated total.
			if len(res.Buckets.Totals[0].Group) == 0 {
				data = res.Buckets.Totals[0].Aggregations[0].Value
			} else {
				data = res.Buckets.Totals
			}
		}
		return iofmt.FormatToJSON(opts.IO.Out(), data, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	headerText := cs.Bold(opts.Query)
	headerText += fmt.Sprintf(" processed in %s", cs.Gray(res.Status.ElapsedTime.String()))
	headerText = fmt.Sprintf("Result of query %s:\n\n", headerText)

	// Deal with table output format for matches.
	if len(res.Matches) > 0 {
		if opts.IO.IsStdoutTTY() {
			fmt.Fprint(opts.IO.Out(), headerText)
		}

		var enc interface {
			Encode(any) error
		}
		if opts.IO.ColorEnabled() {
			enc = jsoncolor.NewEncoder(opts.IO.Out())
		} else {
			enc = json.NewEncoder(opts.IO.Out())
		}

		var data any
		for _, entry := range res.Matches {
			switch opts.Format {
			case iofmt.JSON.String():
				data = entry
			default:
				fmt.Fprintf(opts.IO.Out(), "%s\t", cs.Gray(entry.Time.Format(time.RFC1123)))
				data = entry.Data
			}
			if err = enc.Encode(data); err != nil {
				return err
			}
			fmt.Fprintln(opts.IO.Out())
		}

		return nil
	}

	// Deal with table output format for grouped results.

	// If we have no groups, we just print the aggregated total.
	if len(res.Buckets.Totals[0].Group) == 0 {
		if opts.IO.IsStdoutTTY() {
			fmt.Fprint(opts.IO.Out(), headerText)
		}

		if err = iofmt.FormatToJSON(opts.IO.Out(),
			res.Buckets.Totals[0].Aggregations[0].Value, opts.IO.ColorEnabled()); err != nil {
			return err
		}

		fmt.Fprintln(opts.IO.Out())
		return nil
	}

	// If we have groups, we print a table with the groups as columns and the
	// aggregated totals as rows.
	var (
		header      iofmt.HeaderBuilderFunc
		columnNames = res.GroupBy
	)
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprint(opts.IO.Out(), headerText)
			for _, name := range columnNames {
				trb.AddField(name, cs.Bold)
			}
			trb.AddField(res.Buckets.Totals[0].Aggregations[0].Alias, cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		total := res.Buckets.Totals[k]
		aggValue, _ := json.Marshal(total.Aggregations[0].Value)

		for _, name := range columnNames {
			trb.AddField(fmt.Sprint(total.Group[name]), nil)
		}
		trb.AddField(string(aggValue), nil)
	}

	return iofmt.FormatToTable(opts.IO, len(res.Buckets.Totals), header, nil, contentRow)
}
