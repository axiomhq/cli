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

type createOptions struct {
	*cmdutil.Factory

	// Name of the dataset to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Name string `survey:"name"`
	// Description of the dataset to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Description string `survey:"description"`
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create [(-n|--name) <dataset-name>] [(-d|--description) <dataset-description>]",
		Short: "Create a dataset",

		Aliases: []string{"new"},

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively create a dataset:
			$ axiom dataset create
			
			# Create a dataset and provide the parameters on the command-line:
			$ axiom dataset create --name nginx-logs --description "All Nginx logs"
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeCreate(opts); err != nil {
				return err
			}
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the deployment")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description of the deployment")

	_ = cmd.RegisterFlagCompletionFunc("name", cmdutil.DatasetCompletionFunc(f))
	_ = cmd.RegisterFlagCompletionFunc("description", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("description")
	}

	return cmd
}

func completeCreate(opts *createOptions) error {
	questions := make([]*survey.Question, 0, 2)

	if opts.Name == "" {
		questions = append(questions, &survey.Question{
			Name:   "name",
			Prompt: &survey.Input{Message: "What is the name of the dataset?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(3),
			),
		})
	}

	if opts.Description == "" {
		questions = append(questions, &survey.Question{
			Name:   "description",
			Prompt: &survey.Input{Message: "What is the description of the dataset?"},
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runCreate(ctx context.Context, opts *createOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	if _, err := client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Name:        opts.Name,
		Description: opts.Description,
	}); err != nil {
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
