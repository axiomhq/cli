package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	axiomClient "github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

type updateTokenOptions struct {
	*cmdutil.Factory

	// Token of the user who wants to authenticate against the deployment. The
	// user will be asked for it unless "token-stdin" is set.
	Token string `survey:"token"`
	// TokenType of the supplied token. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	TokenType string `survey:"tokenType"`
	// TokenStdIn reads the token from stdin instead of prompting the user for
	// it.
	TokenStdIn bool
}

func newUpdateTokenCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &updateTokenOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "update-token [--token-stdin] [(-t|--token-type=)personal|ingest]",
		Short: "Update the token used to authenticate against an Axiom deployment",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively update the token of the current Axiom deployment:
			$ axiom auth update-token
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_PERSONAL_ACCESS_TOKEN | axiom auth update-token --token-stdin --token-type="personal"
		`),

		PersistentPreRunE: cmdutil.NeedsActiveDeployment(f),

		PreRunE: func(*cobra.Command, []string) error {
			if !opts.IO.IsStdinTTY() && opts.TokenStdIn {
				return nil
			}
			return completeUpdateToken(opts)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.TokenType != config.Personal && opts.TokenType != config.Ingest {
				return fmt.Errorf("unknown token type %q (choose %q or %q)",
					opts.TokenType, config.Personal, config.Ingest)
			}
			return runUpdateToken(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.TokenStdIn, "token-stdin", false, "Read token from stdin")
	cmd.Flags().StringVarP(&opts.TokenType, "token-type", "t", "", "Type of the token (choose \"personal\" or \"ingest\")")

	_ = cmd.RegisterFlagCompletionFunc("token-stdin", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("token-type", tokenTypeCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("token-stdin")
		_ = cmd.MarkFlagRequired("token-type")
	}

	return cmd
}

func completeUpdateToken(opts *updateTokenOptions) error {
	questions := make([]*survey.Question, 0, 2)

	if opts.TokenType == "" {
		questions = append(questions, &survey.Question{
			Name: "tokenType",
			Prompt: &survey.Select{
				Message: "What kind of token will you provide?",
				Options: validTokenTypes,
			},
		})
	}

	if opts.Token == "" {
		questions = append(questions, &survey.Question{
			Name:   "token",
			Prompt: &survey.Password{Message: "What is your personal access or ingest token?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(36),
				survey.MaxLength(36),
			),
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runUpdateToken(ctx context.Context, opts *updateTokenOptions) error {
	// Read token from stdin, if the appropriate option is set.
	if opts.TokenStdIn {
		contents, err := ioutil.ReadAll(opts.IO.In())
		if err != nil {
			return err
		}

		opts.Token = strings.TrimSuffix(string(contents), "\n")
		opts.Token = strings.TrimSuffix(opts.Token, "\r")
	}

	// A requirement for this command to execute is the presence of an active
	// deployment, so no need to check for existence.
	activeDeployment, _ := opts.Config.GetActiveDeployment()

	if opts.TokenType == config.Personal {
		client, err := axiomClient.New(activeDeployment.URL, opts.Token, activeDeployment.OrganizationID, opts.Config.Insecure)
		if err != nil {
			return err
		}

		stop := opts.IO.StartActivityIndicator()
		defer stop()

		user, err := client.Users.Current(ctx)
		if err != nil {
			return err
		}

		stop()

		if opts.IO.IsStderrTTY() {
			cs := opts.IO.ColorScheme()
			fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s (%s) as %s\n",
				cs.SuccessIcon(), cs.Bold(opts.Config.ActiveDeployment), activeDeployment.URL, cs.Bold(user.Name))
		}
	}

	opts.Config.Deployments[opts.Config.ActiveDeployment] = config.Deployment{
		URL:            activeDeployment.URL,
		Token:          opts.Token,
		TokenType:      opts.TokenType,
		OrganizationID: activeDeployment.OrganizationID,
	}

	return opts.Config.Write()
}
