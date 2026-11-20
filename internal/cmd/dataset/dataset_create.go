package dataset

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

const (
	datasetTypeEvents  = "events"
	datasetTypeMetrics = "metrics"
)

var datasetKinds = map[string]string{
	datasetTypeEvents:  "axiom:events:v1",
	datasetTypeMetrics: "otel:metrics:v1",
}

type createOptions struct {
	*cmdutil.Factory

	// Name of the dataset to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Name string `survey:"name"`
	// Description of the dataset to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Description string `survey:"description"`
	// Type of the dataset to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Type string `survey:"type"`
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create [(-n|--name) <dataset-name>] [(-d|--description) <dataset-description>] [--type events|metrics]",
		Short: "Create a dataset",

		Aliases: []string{"new"},

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively create a dataset:
			$ axiom dataset create
			
			# Create a dataset and provide the parameters on the command-line:
			$ axiom dataset create --name=http-logs --description="All of my logs" --type=events

			# Create an OpenTelemetry metrics dataset:
			$ axiom dataset create --name=app-metrics --description="Application metrics" --type=metrics
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeCreate(cmd.Context(), opts); err != nil {
				return err
			}
			if _, ok := datasetKinds[opts.Type]; !ok {
				return fmt.Errorf("invalid dataset type %q: choose events or metrics", opts.Type)
			}
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the dataset")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description of the dataset")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Type of dataset: events or metrics")

	_ = cmd.RegisterFlagCompletionFunc("type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{datasetTypeEvents, datasetTypeMetrics}, cobra.ShellCompDirectiveNoFileComp
	})

	_ = cmd.RegisterFlagCompletionFunc("name", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("description", cmdutil.NoCompletion)

	if !opts.IO.IsInteractive() {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("description")
		_ = cmd.MarkFlagRequired("type")
	}

	return cmd
}

func completeCreate(ctx context.Context, opts *createOptions) error {
	if !opts.IO.IsInteractive() {
		return nil
	}

	questions := make([]*survey.Question, 0, 3)

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	}

	if opts.Name == "" {
		questions = append(questions, &survey.Question{
			Name:   "name",
			Prompt: &survey.Input{Message: "What is the name of the dataset?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(3),
				surveyext.NotIn(datasetNames),
			),
		})
	}

	if opts.Description == "" {
		questions = append(questions, &survey.Question{
			Name:   "description",
			Prompt: &survey.Input{Message: "What is the description of the dataset?"},
		})
	}

	if opts.Type == "" {
		questions = append(questions, &survey.Question{
			Name: "type",
			Prompt: &survey.Select{
				Message: "What type of dataset?",
				Options: []string{datasetTypeEvents, datasetTypeMetrics},
				Default: datasetTypeEvents,
				Help:    "Events stores logs, traces, and general events. Metrics stores OpenTelemetry metrics.",
			},
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runCreate(ctx context.Context, opts *createOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	if _, err := client.Datasets.Create(ctx, axiom.DatasetCreateRequest{
		Kind:        datasetKinds[opts.Type],
		Name:        opts.Name,
		Description: opts.Description,
	}); err != nil {
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
