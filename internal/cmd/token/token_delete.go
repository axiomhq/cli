package token

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

	// Name of the token to delete. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Name string
	// Force the deletion and skip the confirmation prompt.
	Force bool

	tokenType string
}

func newDeleteCmd(f *cmdutil.Factory, tokenType string) *cobra.Command {
	opts := &deleteOptions{
		Factory:   f,
		tokenType: tokenType,
	}

	cmd := &cobra.Command{
		Use:   "delete [<token-name>] [-f|--force]",
		Short: "Delete a token",

		Aliases: []string{"remove"},

		Args:              cmdutil.PopulateFromArgs(f, &opts.Name),
		ValidArgsFunction: tokenCompletionFunc(f, tokenType),

		Example: heredoc.Doc(`
			# Interactively delete a token:
			$ axiom token ` + tokenType + ` delete
			
			# Delete a token and provide the token name as an argument:
			$ axiom token ` + tokenType + ` delete nginx-logs
		`),

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

	tokenNames, err := getTokenNames(ctx, opts.Factory, opts.tokenType)
	if err != nil {
		return err
	}

	return survey.AskOne(&survey.Select{
		Message: "Which token to delete?",
		Options: tokenNames,
	}, &opts.Name, opts.IO.SurveyIO())
}

func runDelete(ctx context.Context, opts *deleteOptions) error {
	// Deleting must be forced if not running interactively.
	if !opts.IO.IsStdinTTY() && !opts.Force {
		return cmdutil.ErrSilent
	}

	if !opts.Force {
		msg := fmt.Sprintf("Delete token %q?", opts.Name)
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

	var deleteFunc func(context.Context, string) error
	switch opts.tokenType {
	case TypeAPI:
		deleteFunc = client.Tokens.API.Delete
	case TypeIngest:
		deleteFunc = client.Tokens.Ingest.Delete
	case TypePersonal:
		deleteFunc = client.Tokens.Personal.Delete
	default:
		return fmt.Errorf("unknown token type: %s", opts.tokenType)
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	if id, err := getTokenIDFromName(ctx, client, opts.tokenType, opts.Name); err != nil {
		return err
	} else if err = deleteFunc(ctx, id); err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Deleted token %s\n",
			cs.Red("✓"), cs.Bold(opts.Name))
	}

	return nil
}
