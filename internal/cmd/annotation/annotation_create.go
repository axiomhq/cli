package annotation

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type createOptions struct {
	*cmdutil.Factory

	// Type of the annotation to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Type string `survey:"type"`
	// Datasets to attache the annotation to. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Datasets []string `survey:"datasets"`

	// Title of the annotation to create.
	Title string `survey:"title"`
	// Description of the annotation to create.
	Description string `survey:"description"`
	// URL of the annotation to create.
	URL string `survey:"url"`
	// Time of the annotation to create.
	Time string `survey:"time"`
	// EndTime of the annotation to create.
	EndTime string `survey:"end-time"`
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create (-t|--type) <type> (-d|--datasets) <datasets> [(-T|--title) <title>] [(-D|--description) <description>] [(-U|--url) <url>] [(--time) <time>] [(--end-time) <endTime>",
		Short: "Create an annotation",

		Aliases: []string{"new"},

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively create an annotation
			$ axiom annotation create
			
			# Create an annotation and provide the parameters on the command-line:
			$ axiom annotation create --type=deploy --datasets=http-logs
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeCreate(cmd.Context(), opts); err != nil {
				return err
			}
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "", "Type of the annotation (must be lowercase alphanumeric and hyphens)")
	cmd.Flags().StringArrayVarP(&opts.Datasets, "datasets", "d", nil, "Datasets to attach the annotation to")
	cmd.Flags().StringVarP(&opts.Title, "title", "", "", "Title of the annotation")
	cmd.Flags().StringVarP(&opts.Description, "description", "", "", "Description of the annotation")
	cmd.Flags().StringVarP(&opts.URL, "url", "u", "", "URL of the annotation")
	cmd.Flags().StringVarP(&opts.Time, "time", "", "", "Time of the annotation")
	cmd.Flags().StringVarP(&opts.EndTime, "end-time", "", "", "End time of the annotation")

	_ = cmd.RegisterFlagCompletionFunc("type", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("datasets", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("title", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("description", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("time", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("end-time", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("type")
		_ = cmd.MarkFlagRequired("datasets")
	}

	return cmd
}

func completeCreate(ctx context.Context, opts *createOptions) error {
	questions := make([]*survey.Question, 0, 2)

	datasetNames, err := getDatasetNames(ctx, opts.Factory)
	if err != nil {
		return err
	}

	if len(opts.Datasets) == 0 {
		questions = append(questions, &survey.Question{
			Name: "datasets",
			Prompt: &survey.MultiSelect{
				Message: "Which datasets should this annotation be attached to?",
				Options: datasetNames,
			},
			Validate: survey.ComposeValidators(
				survey.Required,
			),
		})
	}

	if opts.Type == "" {
		questions = append(questions, &survey.Question{
			Name:   "type",
			Prompt: &survey.Input{Message: "What is the type of the annotation?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				validateType,
			),
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

var typeRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func validateType(ans any) error {
	typ, ok := ans.(string)
	if !ok {
		return fmt.Errorf("expected a string, got %T", ans)
	}

	if !typeRegex.MatchString(typ) {
		return fmt.Errorf("type can only be lowercase letters, numbers, and hyphens")
	}

	return nil
}

func runCreate(ctx context.Context, opts *createOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	if opts.Type != "" {
		if err := validateType(opts.Type); err != nil {
			return err
		}
	}

	var startTime, endTime time.Time
	if opts.Time != "" {
		startTime, err = time.Parse(time.RFC3339, opts.Time)
		if err != nil {
			return fmt.Errorf("invalid time: %w", err)
		}
	}
	if opts.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, opts.EndTime)
		if err != nil {
			return fmt.Errorf("invalid end-time: %w", err)
		}
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	annotation, err := client.Annotations.Create(ctx, &axiom.AnnotationCreateRequest{
		Type:        opts.Type,
		Datasets:    opts.Datasets,
		Title:       opts.Title,
		Description: opts.Description,
		URL:         opts.URL,
		Time:        startTime,
		EndTime:     endTime,
	})
	if err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Created annotation %s\n",
			cs.SuccessIcon(), cs.Bold(annotation.ID))
	}

	return nil
}
