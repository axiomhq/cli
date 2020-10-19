package dataset

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type createOptions struct {
	*cmdutil.Factory

	// Name of the dataset to create. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Name string
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create <dataset-name>",
		Short: "Create a dataset",

		Aliases: []string{"new"},

		Args: cobra.MaximumNArgs(1),

		Example: heredoc.Doc(`
			# Interactively create a dataset:
			$ axiom dataset create
			
			# Create a dataset and providen the dataset name as an argument:
			$ axiom dataset create nginx-logs
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !opts.IO.IsStdinTTY() && len(args) == 0 {
				return cmdutil.ErrNoPromptArgRequired
			} else if len(args) == 1 {
				opts.Name = args[0]
				return nil
			}
			return completeCreate(opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreate(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeCreate(opts *createOptions) error {
	v := survey.ComposeValidators(
		survey.Required,
		survey.MinLength(3),
	)
	return survey.AskOne(&survey.Input{
		Message: "What is the name of the dataset?",
	}, &opts.Name, opts.IO.SurveyIO(), survey.WithValidator(v))
}

func runCreate(ctx context.Context, opts *createOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	if _, err := client.Datasets.CreateDataset(ctx, opts.Name); err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Created dataset %s\n", cs.SuccessIcon(), cs.Bold(opts.Name))
	}

	return nil
}
