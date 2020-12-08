package dataset

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type deleteOptions struct {
	*cmdutil.Factory

	// Name of the dataset to delete. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Name string
	// Force the deletion and skip the confirmation prompt.
	Force bool
}

func newDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &deleteOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "delete [<dataset-name>]",
		Short: "Delete a dataset",

		Aliases: []string{"remove"},

		Args:              cmdutil.PopulateFromArgs(f, &opts.Name),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively delete a dataset:
			$ axiom dataset delete
			
			# Delete a dataset and provide the dataset name as an argument:
			$ axiom dataset delete nginx-logs
		`),

		PreRunE: cmdutil.NeedsDatasets(f),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeDelete(cmd.Context(), opts); err != nil {
				return err
			}
			return runDelete(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func completeDelete(ctx context.Context, opts *deleteOptions) error {
	if opts.Name != "" {
		return nil
	}

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to delete?",
		Options: datasetNames,
	}, &opts.Name, opts.IO.SurveyIO())
}

func runDelete(ctx context.Context, opts *deleteOptions) error {
	// Deleting must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Delete dataset %q?", opts.Name)
		if overwrite, err := surveyext.AskConfirm(msg, false, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	if err = client.Datasets.Delete(ctx, opts.Name); err != nil {
		stop()
		return err
	}
	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Deleted dataset %s\n",
			cs.Red("✓"), cs.Bold(opts.Name))
	}

	return nil
}
