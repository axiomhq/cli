package annotation

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

	// ID of the annotation to delete. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	ID string `survey:"id"`
	// Force the deletion and skip the confirmation prompt.
	Force bool
}

func newDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &deleteOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an annotation",

		Aliases: []string{"remove"},

		Args: cmdutil.PopulateFromArgs(f, &opts.ID),

		Example: heredoc.Doc(`
			# Interactively delete an annotation:
			$ axiom annotation delete
			
			# Delete an annotation and provide the annotation id as an argument:
			$ axiom annotation delete ann_1234567890
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeDelete(opts); err != nil {
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

func completeDelete(opts *deleteOptions) error {
	if opts.ID != "" {
		return nil
	}

	return survey.Ask([]*survey.Question{{
		Name: "id",
		Prompt: &survey.Input{
			Message: "What is the ID of the annotation to delete?",
		},
		Validate: survey.Required,
	}}, &opts.ID, opts.IO.SurveyIO())
}

func runDelete(ctx context.Context, opts *deleteOptions) error {
	// Deleting must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Delete annotation %s?", opts.ID)
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

	if err = client.Annotations.Delete(ctx, opts.ID); err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Deleted annotation %s\n",
			cs.Red("✓"), cs.Bold(opts.ID))
	}

	return nil
}
