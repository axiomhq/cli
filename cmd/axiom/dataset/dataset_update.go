package dataset

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type updateOptions struct {
	*cmdutil.Factory

	// Name of the dataset to update. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Name string `survey:"name"`
	// Description of the dataset to update. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Description string `survey:"description"`
}

func newUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &updateOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "update [<dataset-name>] [(-d|--description) <dataset-description>]",
		Short: "Update a dataset",

		Args:              cmdutil.PopulateFromArgs(f, &opts.Name),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively update a dataset:
			$ axiom dataset update

			# Interactively update dataset "nginx-logs":
			$ axiom dataset update nginx-logs
			
			# Update a dataset and provide the parameters on the command-line:
			$ axiom dataset update nginx-logs --description "All Nginx logs"
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeUpdate(cmd.Context(), opts); err != nil {
				return err
			}
			return runUpdate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description of the deployment")

	_ = cmd.RegisterFlagCompletionFunc("description", cmdutil.NoCompletion)

	return cmd
}

func completeUpdate(ctx context.Context, opts *updateOptions) error {
	questions := make([]*survey.Question, 0, 2)

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	}

	if opts.Name == "" {
		questions = append(questions, &survey.Question{
			Name: "name",
			Prompt: &survey.Select{
				Message: "Which dataset to update?",
				Options: datasetNames,
			},
		})
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	dataset, err := client.Datasets.Get(ctx, opts.Name)
	if err != nil {
		progStop()
		return err
	}
	progStop()

	if opts.Description == "" {
		questions = append(questions, &survey.Question{
			Name: "description",
			Prompt: &survey.Input{
				Message: "What is the description of the dataset?",
				Default: dataset.Description,
			},
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runUpdate(ctx context.Context, opts *updateOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	if _, err := client.Datasets.Update(ctx, opts.Name, axiom.DatasetUpdateRequest{
		Description: opts.Description,
	}); err != nil {
		stop()
		return err
	}
	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Updated dataset %s\n",
			cs.SuccessIcon(), cs.Bold(opts.Name))
	}

	return nil
}
