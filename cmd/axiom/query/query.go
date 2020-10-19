package query

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	axiomdb "axicode.axiom.co/watchmakers/axiomdb/client"
	swagger "axicode.axiom.co/watchmakers/axiomdb/client/swagger/datasets"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/printer"
	"github.com/axiomhq/cli/pkg/terminal"
	"github.com/axiomhq/cli/pkg/utils"
)

const (
	formatJSON = "json"
)

var validFormats = []string{formatJSON}

type options struct {
	*cmdutil.Factory

	// Dataset to query. If not supplied as an argument, which is optional, the
	// user will be asked for it.
	Dataset string `survey:"dataset"`
	// Query to execute. If not supplied as an argument, which is optional, the
	// user will be asked for it.
	Query string `survey:"query"`
	// StreamingDuration to apply to the query.
	StreamingDuration time.Duration
	// NoCache disables the query cache.
	NoCache bool
	// Filename is the path to a file which contains a query.
	Filename string
	// Format to output data in. Defaults to tabular output.
	Format string
}

// NewQueryCmd creates and returns the query command.
func NewQueryCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "query [(-d|--dataset) <dataset-name>] [(-q|--query) <query> | (-f|--file) <filename>] [--streaming-duration <duration>] [--no-cache] [(-F|--format=)JSON]",
		Short: "Query data",
		Long:  `Query data from an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively query a dataset:
			$ axiom query
						
			# Query the "logs" dataset for the servers having the best
			# latencies using the axiom query language:
			$ axiom query -d logs -q "select server, min(ping) group by server"
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Needs(
				cmdutil.NeedsActiveBackend(f),
				cmdutil.NeedsDatasets(f),
			)(cmd, args); err != nil {
				return err
			}

			if opts.Filename != "" && opts.Query != "" {
				return cmdutil.NewFlagErrorf("cannot use (-f|--file) and (-q|--query), choose one")
			} else if opts.IO.IsStdinTTY() {
				return complete(cmd.Context(), opts)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Dataset, "dataset", "d", "", "Dataset to query")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Query to execute")
	cmd.Flags().DurationVar(&opts.StreamingDuration, "streaming-duration", 0, "Streaming duration to apply")
	cmd.Flags().BoolVar(&opts.NoCache, "no-cache", false, "Disable query cache")
	cmd.Flags().StringVarP(&opts.Filename, "file", "f", "", "File to load query from")
	cmd.Flags().StringVarP(&opts.Format, "format", "F", "", "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("dataset", cmdutil.DatasetCompletionFunc(f))
	_ = cmd.RegisterFlagCompletionFunc("query", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("format", formatCompletion)

	_ = cmd.MarkFlagFilename("file", "json")

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("dataset")
		_ = cmd.MarkFlagRequired("query")
	}

	return cmd
}

func complete(ctx context.Context, opts *options) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	datasets, err := client.Datasets.List(ctx, axiomdb.ListOptions{})
	if err != nil {
		return err
	}

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	stop()

	questions := make([]*survey.Question, 0, 2)

	if opts.Dataset == "" {
		questions = append(questions, &survey.Question{
			Name: "dataset",
			Prompt: &survey.Select{
				Message: "Which dataset to query?",
				Options: datasetNames,
			},
		})
	}

	if opts.Query == "" && opts.Filename == "" {
		questions = append(questions, &survey.Question{
			Name:   "query",
			Prompt: &survey.Editor{Message: "Which query to run on the dataset?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(10),
			),
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	if opts.Filename != "" {
		b, err := ioutil.ReadFile(opts.Filename)
		if err != nil {
			return err
		}
		opts.Query = string(b)
	}

	// Sanitize.
	opts.Query = strings.TrimSpace(opts.Query)

	var query swagger.QueryRequest

	// TODO(lukasmalkmus): Unmarshal query.
	query.StartTime = time.Now().Add(-time.Hour)
	query.EndTime = time.Now()

	// if err := json.Unmarshal([]byte(opts.Query), &query); err != nil {
	// 	return err
	// }

	client, err := opts.Client()
	if err != nil {
		return err
	}

	progStop := opts.IO.StartProgressIndicator()
	defer progStop()

	res, err := client.Datasets.Query(ctx, opts.Dataset, query, axiomdb.QueryOptions{
		StreamingDuration: opts.StreamingDuration,
		NoCache:           opts.NoCache,
		// NoEstimates:       opts.NoEstimates,
		// NoFunctions:       opts.NoFunctions,
		// FunctionCount:     opts.FunctionCount,
	})
	if err != nil {
		return err
	}

	progStop()

	if opts.Format != "" {
		writeInfo(opts.IO, res)

		switch opts.Format {
		case formatJSON:
			return json.NewEncoder(opts.IO.Out()).Encode(res)
		}
	}

	pagerStop, err := opts.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	writeInfo(opts.IO, res)

	return printer.Table(opts.IO, res.Matches, true)
}

func formatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(validFormats))
	for _, validFormat := range validFormats {
		if strings.HasPrefix(validFormat, toComplete) {
			res = append(res, validFormat)
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}

func writeInfo(io *terminal.IO, res *swagger.Result) {
	if io.IsStdoutTTY() {
		cs := io.ColorScheme()
		ts := time.Duration(res.Status.ElapsedTime) * time.Microsecond
		fmt.Fprintf(io.Out(), "\nQueried %s in %s:\n\n",
			utils.Pluralize(cs, "result", len(res.Matches)),
			cs.Bold(ts.String()),
		)
	}
}
