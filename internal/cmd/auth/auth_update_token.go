package auth

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type updateTokenOptions struct {
	*cmdutil.Factory
	// Token of the user who wants to authenticate against the deployment. The
	// user will be asked for it unless the session has no TTY attached, in
	// which case the token is read from stdin.
	Token string
}

func newUpdateTokenCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &updateTokenOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "update-token",
		Short: "Update the token used to authenticate against an Axiom deployment",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively update the token of the current Axiom deployment:
			$ axiom auth update-token
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_PERSONAL_ACCESS_TOKEN | axiom auth update-token
		`),

		PersistentPreRunE: cmdutil.NeedsActiveDeployment(f),

		PreRunE: func(*cobra.Command, []string) error {
			if !opts.IO.IsStdinTTY() {
				return nil
			}
			return completeUpdateToken(opts)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateToken(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeUpdateToken(opts *updateTokenOptions) error {
	if opts.Token != "" {
		return nil
	}

	return survey.AskOne(&survey.Password{
		Message: "What is your personal access or ingest token?",
	}, &opts.Token, survey.WithValidator(survey.ComposeValidators(
		survey.Required,
		surveyext.ValidateToken,
	)), opts.IO.SurveyIO())
}

func runUpdateToken(ctx context.Context, opts *updateTokenOptions) error {
	// Read token from stdin, if no TTY is attached.
	if !opts.IO.IsStdinTTY() {
		var err error
		if opts.Token, err = readTokenFromStdIn(opts.IO.In()); err != nil {
			return err
		}
	}

	// A requirement for this command to execute is the presence of an active
	// deployment, so no need to check for existence.
	activeDeployment, _ := opts.Config.GetActiveDeployment()

	client, err := client.New(ctx, activeDeployment.URL, opts.Token, activeDeployment.OrganizationID, opts.Config.Insecure)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var user *axiom.AuthenticatedUser
	if axiom.IsPersonalToken(opts.Token) {
		if user, err = client.Users.Current(ctx); err != nil {
			return err
		}
	} else if axiom.IsIngestToken(opts.Token) {
		if err = client.Tokens.Ingest.Validate(ctx); err != nil {
			return err
		}
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()

		if user != nil {
			if activeDeployment.URL == axiom.CloudURL {
				organization, err := client.Organizations.Get(ctx, activeDeployment.OrganizationID)
				if err != nil {
					return err
				}

				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s as %s\n",
					cs.SuccessIcon(), cs.Bold(organization.Name), cs.Bold(user.Name))
			} else {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s as %s\n",
					cs.SuccessIcon(), cs.Bold(opts.Config.ActiveDeployment), cs.Bold(user.Name))
			}
		} else {
			if activeDeployment.URL == axiom.CloudURL {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s %s\n",
					cs.SuccessIcon(), cs.Bold(activeDeployment.OrganizationID), cs.Red(cs.Bold("(ingestion only!)")))
			} else {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s %s\n",
					cs.SuccessIcon(), cs.Bold(opts.Config.ActiveDeployment), cs.Red(cs.Bold("(ingestion only!)")))
			}
		}
	}

	opts.Config.Deployments[opts.Config.ActiveDeployment] = config.Deployment{
		URL:            activeDeployment.URL,
		Token:          opts.Token,
		OrganizationID: activeDeployment.OrganizationID,
	}

	return opts.Config.Write()
}
