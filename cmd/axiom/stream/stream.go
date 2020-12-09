package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/nwidger/jsoncolor"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

const (
	formatJSON = "json"
)

var validFormats = []string{formatJSON}

const (
	streamingDuration = time.Second * 2
)

type options struct {
	*cmdutil.Factory

	// Dataset to stream from. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Dataset string
	// Format to output data in. Defaults to tabular output.
	Format string
}

// NewStreamCmd creates and returns the stream command.
func NewStreamCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "stream [<dataset-name>] [(-f|--format=)json]",
		Short: "Livestream data",
		Long:  `Livestream data from an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Dataset),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively stream a dataset:
			$ axiom stream
			
			# Stream the "nginx-logs" dataset:
			$ axiom stream nginx-logs
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsDatasets(f),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := complete(cmd.Context(), opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "", "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", formatCompletion)

	return cmd
}

func complete(ctx context.Context, opts *options) error {
	if opts.Dataset != "" {
		return nil
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		stop()
		return err
	}
	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to stream from?",
		Options: datasetNames,
	}, &opts.Dataset, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	cs := opts.IO.ColorScheme()

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Streaming events from dataset %s:\n\n", cs.Bold(opts.Dataset))
	}

	var enc interface {
		Encode(interface{}) error
	}
	if opts.IO.ColorEnabled() {
		enc = jsoncolor.NewEncoder(opts.IO.Out())
	} else {
		enc = json.NewEncoder(opts.IO.Out())
	}

	t := time.NewTicker(streamingDuration)
	defer t.Stop()

	lastRequest := time.Now().Add(-time.Nanosecond)
	for {
		queryCtx, queryCancel := context.WithTimeout(ctx, streamingDuration)

		res, err := client.Datasets.Query(queryCtx, opts.Dataset, query.Query{
			StartTime: lastRequest,
			EndTime:   time.Now(),
		}, query.Options{
			StreamingDuration: streamingDuration,
		})
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			queryCancel()
			return err
		}

		queryCancel()

		if res != nil && len(res.Matches) > 0 {
			lastRequest = res.Matches[len(res.Matches)-1].Time.Add(time.Nanosecond)

			for _, entry := range res.Matches {
				switch opts.Format {
				case formatJSON:
					_ = enc.Encode(entry)
				default:
					fmt.Fprintf(opts.IO.Out(), "%s\t",
						cs.Gray(entry.Time.Format(time.RFC1123)))
					_ = enc.Encode(entry.Data)
				}
				fmt.Fprintln(opts.IO.Out())
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

func formatCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(validFormats))
	for _, validFormat := range validFormats {
		if strings.HasPrefix(validFormat, toComplete) {
			res = append(res, validFormat)
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}
