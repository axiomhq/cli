package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom/querylegacy"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

const streamingDuration = time.Second * 2

type options struct {
	*cmdutil.Factory

	// Dataset to stream from. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Dataset string
	// Format to output data in. Defaults to tabular output.
	Format string
}

// NewCmd creates and returns the stream command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "stream [<dataset-name>] [(-f|--format)=json|table]",
		Short: "Livestream data",
		Long:  `Livestream data from an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Dataset),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively stream a dataset:
			$ axiom stream
			
			# Stream the "http-logs" dataset:
			$ axiom stream http-logs
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
			if err := complete(cmd.Context(), opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func complete(ctx context.Context, opts *options) error {
	if opts.Dataset != "" {
		return nil
	}

	// Just fetch a list of available datasets if a Personal Access Token is
	// used.
	var datasetNames []string
	if dep, ok := opts.Config.GetActiveDeployment(); ok && client.IsPersonalToken(dep.Token) {
		client, err := opts.Client(ctx)
		if err != nil {
			return err
		}

		stop := opts.IO.StartActivityIndicator()
		defer stop()

		datasets, err := client.Datasets.List(ctx)
		if err != nil {
			return err
		}

		stop()

		datasetNames = make([]string, len(datasets))
		for i, dataset := range datasets {
			datasetNames[i] = dataset.Name
		}
	}

	if len(datasetNames) == 0 {
		return errors.New("missing dataset")
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to stream from?",
		Default: datasetNames[0],
		Options: datasetNames,
	}, &opts.Dataset, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	cs := opts.IO.ColorScheme()

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Streaming events from dataset %s:\n\n", cs.Bold(opts.Dataset))
	}

	t := time.NewTicker(streamingDuration)
	defer t.Stop()

	lastRequest := time.Now().Add(-time.Nanosecond)
	for {
		queryCtx, queryCancel := context.WithTimeout(ctx, streamingDuration)

		res, err := client.Datasets.QueryLegacy(queryCtx, opts.Dataset, querylegacy.Query{
			StartTime: lastRequest,
			EndTime:   time.Now(),
		}, querylegacy.Options{
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
				var data any
				switch opts.Format {
				case iofmt.JSON.String():
					data = entry
				default:
					fmt.Fprintf(opts.IO.Out(), "%s\t", cs.Gray(entry.Time.Format(time.RFC1123)))
					data = entry.Data
				}
				if err = iofmt.FormatToJSON(opts.IO.Out(), data, opts.IO.ColorEnabled()); err != nil {
					return err
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
