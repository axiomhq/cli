package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
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
		Short: "Update the token used to authenticate against Axiom",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively update the token of the current configuration:
			$ axiom auth update-token
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_TOKEN | axiom auth update-token
		`),

		PersistentPreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, NewLoginCmd(f)),
			cmdutil.NeedsActiveDeployment(f),
		),

		PreRunE: func(*cobra.Command, []string) error {
			if !opts.IO.IsStdinTTY() {
				return nil
			}
			return completeUpdateToken(opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUpdateToken(cmd.Context(), opts)
		},
	}

	return cmd
}

func completeUpdateToken(opts *updateTokenOptions) error {
	if opts.Token != "" {
		return nil
	}

	// A requirement for this command to execute is the presence of an active
	// deployment, so no need to check for existence.
	activeDeployment, _ := opts.Config.GetActiveDeployment()

	depURL := activeDeployment.URL
	if depURL != "" && !strings.HasPrefix(depURL, "http://") && !strings.HasPrefix(depURL, "https://") {
		depURL = "https://" + depURL
	}

	u, err := url.ParseRequestURI(depURL)
	if err != nil {
		return err
	}
	u.Path = "/profile"

	return survey.AskOne(&survey.Password{
		Message: fmt.Sprintf("What is your personal access token (create one over at %s)?", u.String()),
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

	axiomClient, err := client.New(ctx, activeDeployment.URL, opts.Token, activeDeployment.OrganizationID, "", "", opts.Config.Insecure)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	user, err := axiomClient.Users.Current(ctx)
	if err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()

		if client.IsPersonalToken(opts.Token) {
			organization, err := axiomClient.Organizations.Get(ctx, activeDeployment.OrganizationID)
			if err != nil {
				return err
			}

			fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s as %s\n",
				cs.SuccessIcon(), cs.Bold(organization.Name), cs.Bold(user.Name))
		} else {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s %s\n",
				cs.SuccessIcon(), cs.Bold(activeDeployment.OrganizationID), cs.Red(cs.Bold("(ingestion/query only!)")))
		}
	}

	opts.Config.Deployments[opts.Config.ActiveDeployment] = config.Deployment{
		URL:            activeDeployment.URL,
		Token:          opts.Token,
		OrganizationID: activeDeployment.OrganizationID,
	}

	return opts.Config.Write()
}
