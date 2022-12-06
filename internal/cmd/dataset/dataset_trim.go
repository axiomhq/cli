package dataset

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type trimOptions struct {
	*cmdutil.Factory

	// Name of the dataset to update. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Name string `survey:"name"`
	// Duration of the dataset to update. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Duration time.Duration `survey:"duration"`
	// Force the creation and skip the confirmation prompt.
	Force bool
}

func newTrimCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &trimOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "trim [<dataset-name>] [(-d|--duration) <max-duration>] [-f|--force]",
		Short: "Trim a dataset to a given size",

		Args:              cmdutil.PopulateFromArgs(f, &opts.Name),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively trim a dataset:
			$ axiom dataset trim

			# Interactively trim dataset "http-logs":
			$ axiom dataset trim http-logs
			
			# Trim a dataset and provide the parameters on the command-line.
			# This trims the dataset down to the last 12 hours:
			$ axiom dataset trim http-logs --duration="12h"
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeTrim(cmd.Context(), opts); err != nil {
				return err
			}
			return runTrim(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Duration, "duration", "d", 0, "Duration to trim the dataset to")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("duration", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("duration")
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func completeTrim(ctx context.Context, opts *trimOptions) error {
	questions := make([]*survey.Question, 0, 2)

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	} else if len(datasetNames) == 1 {
		opts.Name = datasetNames[0]
		return nil
	}

	if opts.Name == "" {
		questions = append(questions, &survey.Question{
			Name: "name",
			Prompt: &survey.Select{
				Message: "Which dataset to trim?",
				Default: datasetNames[0],
				Options: datasetNames,
			},
		})
	}

	if opts.Duration == 0 {
		questions = append(questions, &survey.Question{
			Name: "duration",
			Prompt: &survey.Input{
				Message: "Which duration to trim the dataset to?",
			},
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runTrim(ctx context.Context, opts *trimOptions) error {
	// Trimming must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Are you sure you want to trim dataset %q down to the last %s?", opts.Name, opts.Duration)
		if overwrite, err := surveyext.AskConfirm(msg, false, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	res, err := client.Datasets.Trim(ctx, opts.Name, opts.Duration)
	if err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Trimmed dataset %s (dropped %s blocks)\n",
			cs.SuccessIcon(), cs.Bold(opts.Name), cs.Bold(strconv.Itoa(res.BlocksDeleted)))
	}

	return nil
}
