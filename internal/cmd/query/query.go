package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/araddon/dateparse"
	"github.com/axiomhq/axiom-go/axiom/apl"
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
	// NoCache disables cache usage for the query.
	NoCache bool
	// Save the query on the server.
	Save bool

	startTime time.Time
	endTime   time.Time
}

// NewCmd creates and returns the query command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "query [<apl-query>] [(-f|--format)=json|table] [--start-time <start-time>] [--end-time <end-time>] [--timestamp-format <timestamp-format>] [-c|--no-cache] [-s|--save]",
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
			# Query the "nginx-logs" dataset for logs with a 304 status code:
			$ axiom query "['nginx-logs'] | where response == 304"
			
			# Query all logs of the "http" dataset and save the query in the
			# history. The histories entry ID is returned with the result:
			$ axiom query -s "['http']"
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
	cmd.Flags().StringVar(&opts.StartTime, "start-time", "", "Start time of the query")
	cmd.Flags().StringVar(&opts.EndTime, "end-time", "", "End time of the query")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format used in the the timestamp field. Default uses a heuristic parser. Must be expressed using the reference time 'Mon Jan 2 15:04:05 -0700 MST 2006'")
	cmd.Flags().BoolVarP(&opts.NoCache, "no-cache", "c", false, "Disable cache usage")
	cmd.Flags().BoolVarP(&opts.Save, "save", "s", false, "Save query on the server side")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)
	_ = cmd.RegisterFlagCompletionFunc("start-time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("end-time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("timestamp-format", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("no-cache", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("save", cmdutil.NoCompletion)

	return cmd
}

func complete(opts *options) (err error) {
	if ts := opts.StartTime; ts != "" {
		if tsf := opts.TimestampFormat; tsf != "" {
			opts.startTime, err = time.Parse(tsf, ts)
		} else {
			opts.startTime, err = dateparse.ParseAny(ts)
		}
		if err != nil {
			return err
		}
	}

	if ts := opts.EndTime; ts != "" {
		if tsf := opts.TimestampFormat; tsf != "" {
			opts.endTime, err = time.Parse(tsf, ts)
		} else {
			opts.endTime, err = dateparse.ParseAny(ts)
		}
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

	cs := opts.IO.ColorScheme()

	var enc interface {
		Encode(any) error
	}
	if opts.IO.ColorEnabled() {
		enc = jsoncolor.NewEncoder(opts.IO.Out())
	} else {
		enc = json.NewEncoder(opts.IO.Out())
	}

	res, err := client.Datasets.APLQuery(ctx, opts.Query, apl.Options{
		StartTime: opts.startTime,
		EndTime:   opts.endTime,
		NoCache:   opts.NoCache,
		Save:      opts.Save,
	})
	if err != nil {
		return err
	} else if res == nil || len(res.Matches) == 0 {
		return errors.New("query returned no results")
	}

	if opts.IO.IsStdoutTTY() {
		s := cs.Bold(opts.Query)
		if res.SavedQueryID != "" {
			s += fmt.Sprintf(" (saved as %s)", cs.Bold(res.SavedQueryID))
		}
		fmt.Fprintf(opts.IO.Out(), "Result of query %s:\n\n", s)
	}

	for _, entry := range res.Matches {
		switch opts.Format {
		case iofmt.JSON.String():
			_ = enc.Encode(entry)
		default:
			fmt.Fprintf(opts.IO.Out(), "%s\t",
				cs.Gray(entry.Time.Format(time.RFC1123)))
			_ = enc.Encode(entry.Data)
		}
		fmt.Fprintln(opts.IO.Out())
	}

	return nil
}
