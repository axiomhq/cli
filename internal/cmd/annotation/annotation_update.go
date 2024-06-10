package annotation

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type updateOptions struct {
	*cmdutil.Factory

	// Type of the annotation to update.
	ID string `survey:"id"`
	// Type of the annotation to update.
	Type string `survey:"type"`
	// Datasets to attach the updated annotation to.
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

func newUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &updateOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "update [<id>] [(-t|--type) <type>] [(-d|--datasets) <datasets>] [(-T|--title) <title>] [(-D|--description) <description>] [(-U|--url) <url>] [(--time) <time>] [(--end-time) <endTime>",
		Short: "Update an annotation",

		Args:    cmdutil.PopulateFromArgs(f, &opts.ID),
		Aliases: []string{"update"},

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively update an annotation
			$ axiom annotation update
			
			# Update an annotation and provide the parameters on the command-line:
			$ axiom annotation update --type=deploy --datasets=http-logs
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeUpdate(opts); err != nil {
				return err
			}
			return runUpdate(cmd.Context(), opts)
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

func completeUpdate(opts *updateOptions) error {
	questions := make([]*survey.Question, 0, 2)

	if opts.ID == "" {
		questions = append(questions, &survey.Question{
			Name:   "id",
			Prompt: &survey.Input{Message: "What is the ID of the annotation to update?"},
			Validate: survey.ComposeValidators(
				survey.Required,
			),
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runUpdate(ctx context.Context, opts *updateOptions) error {
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

	if opts.Type == "" && opts.Datasets == nil && opts.Title == "" && opts.Description == "" && opts.URL == "" && startTime.IsZero() && endTime.IsZero() {
		return nil // Nothing to update
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	annotation, err := client.Annotations.Update(ctx, opts.ID, &axiom.AnnotationUpdateRequest{
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
		fmt.Fprintf(opts.IO.ErrOut(), "%s Updated annotation %s\n",
			cs.SuccessIcon(), cs.Bold(annotation.ID))
	}

	return nil
}
