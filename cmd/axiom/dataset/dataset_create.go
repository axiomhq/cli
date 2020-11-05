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
		Use:   "create [<dataset-name>]",
		Short: "Create a dataset",

		Aliases: []string{"new"},

		Args: cmdutil.ChainPositionalArgs(
			cobra.MaximumNArgs(1),
			cmdutil.PopulateFromArgs(f, &opts.Name),
		),

		Example: heredoc.Doc(`
			# Interactively create a dataset:
			$ axiom dataset create
			
			# Create a dataset and provide the dataset name as an argument:
			$ axiom dataset create nginx-logs
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeCreate(opts); err != nil {
				return err
			}
			return runCreate(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeCreate(opts *createOptions) error {
	if opts.Name != "" {
		return nil
	}

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

	stop := opts.IO.StartActivityIndicator()
	if _, err := client.Datasets.CreateDataset(ctx, opts.Name); err != nil {
		stop()
		return err
	}
	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Created dataset %s\n",
			cs.SuccessIcon(), cs.Bold(opts.Name))
	}

	return nil
}
