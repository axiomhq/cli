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
		Use:   "delete <dataset-name>",
		Short: "Delete a dataset",

		Aliases: []string{"remove"},

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively delete a dataset:
			$ axiom dataset delete
			
			# Delete a dataset and providen the dataset name as an argument:
			$ axiom dataset delete nginx-logs
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.NeedsDatasets(f)(cmd, args); err != nil {
				return err
			}

			if !opts.IO.IsStdinTTY() && len(args) == 0 {
				return cmdutil.ErrNoPromptArgRequired
			} else if len(args) == 1 {
				opts.Name = args[0]
				return nil
			}
			return completeDelete(cmd.Context(), opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDelete(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	return cmd
}

func completeDelete(ctx context.Context, opts *deleteOptions) error {
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
	cs := opts.IO.ColorScheme()

	if !opts.Force {
		if !opts.IO.IsStdinTTY() {
			return cmdutil.ErrSilent
		}

		msg := fmt.Sprintf("Delete dataset %q?", opts.Name)
		if overwrite, err := surveyext.AskConfirm(msg, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	if err = client.Datasets.Delete(ctx, opts.Name); err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		fmt.Fprintf(opts.IO.ErrOut(), "%s Deleted dataset %s\n", cs.Red("✓"), cs.Bold(opts.Name))
	}

	return nil
}
