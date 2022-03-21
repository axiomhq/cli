package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
)

const (
	typeCloud    = "Cloud"
	typeSelfhost = "Selfhost"
)

var validDeploymentTypes = []string{typeCloud, typeSelfhost}

type loginOptions struct {
	*cmdutil.Factory

	// Type of the deployment to authenticate with. Default to "Cloud". Can be
	// overwritten by flag.
	Type string
	// Url of the deployment to authenticate with. Default to the Axiom Cloud
	// URL. Can be overwritten by flag.
	URL string
	// Alias of the deployment for future reference. If not supplied as a flag,
	// which is optional, the user will be asked for it.
	Alias string
	// Token of the user who wants to authenticate against the deployment. The
	// user will be asked for it unless the session has no TTY attached, in
	// which case the token is read from stdin.
	Token string
	// OrganizationID of the organization the supplied token is valid for. If
	// not supplied as a flag, which is optional, the user will be asked for it.
	// Only valid for cloud deployments.
	OrganizationID string
	// Force the creation and skip the confirmation prompt.
	Force bool
}

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login [(-t|--type)=cloud|selfhost] [(-u|--url) <deployment-url>] [(-a|--alias) <deployment-alias>] [(-o|--org-id) <organization-id>] [-f|--force]",
		Short: "Login to an Axiom deployment",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively authenticate against an Axiom deployment:
			$ axiom auth login
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_ACCESS_TOKEN | axiom auth login --alias="axiom-eu-west-1" --url="https://axiom.eu-west-1.aws.com" -f
		`),

		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if !opts.IO.IsStdinTTY() {
				return nil
			}
			return completeLogin(cmd.Context(), opts)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", strings.ToLower(typeCloud), "Type of the deployment")
	cmd.Flags().StringVarP(&opts.URL, "url", "u", axiom.CloudURL, "Url of the deployment")
	cmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias of the deployment")
	cmd.Flags().StringVarP(&opts.OrganizationID, "org-id", "o", "", "Organization ID")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("type", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("alias", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("org-id", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("type")
		_ = cmd.MarkFlagRequired("url")
		_ = cmd.MarkFlagRequired("alias")
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func completeLogin(ctx context.Context, opts *loginOptions) error {
	// 1. Cloud or Selfhost?
	if opts.Type == "" {
		if err := survey.AskOne(&survey.Select{
			Message: "Which kind of Axiom deployment are you using?",
			Options: validDeploymentTypes,
		}, &opts.Type, opts.IO.SurveyIO()); err != nil {
			return err
		}
	}

	opts.Type = strings.ToLower(opts.Type)

	// 2. If Cloud, set the correct URL instead of asking the user for it.
	if opts.Type == strings.ToLower(typeCloud) {
		opts.URL = axiom.CloudURL
	} else if opts.URL == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "What is the url of the deployment?",
		}, &opts.URL, survey.WithValidator(survey.ComposeValidators(
			survey.Required,
			surveyext.ValidateURL,
		)), opts.IO.SurveyIO()); err != nil {
			return err
		}
	}

	// 3. The token to use.
	if err := survey.AskOne(&survey.Password{
		Message: "What is your personal access token?",
	}, &opts.Token, survey.WithValidator(survey.ComposeValidators(
		survey.Required,
		surveyext.ValidateToken,
	)), opts.IO.SurveyIO()); err != nil {
		return err
	}

	// 4. Try to authenticate and fetch the organizations available to the user
	// in case the deployment is a cloud deployment. If only one organization is
	// available, that one is selected by default, without asking the user for
	// it.
	if opts.Type == strings.ToLower(typeCloud) && opts.OrganizationID == "" {
		client, err := client.New(ctx, opts.URL, opts.Token, "", opts.Config.Insecure)
		if err != nil {
			return err
		}

		if organizations, err := client.Organizations.Selfhost.List(ctx); err != nil {
			return err
		} else if len(organizations) == 1 {
			opts.OrganizationID = organizations[0].ID
		} else {
			organizationIDs := make([]string, len(organizations))
			for k, organization := range organizations {
				organizationIDs[k] = organization.ID
			}

			if err := survey.AskOne(&survey.Select{
				Message: "Which organization to use?",
				Options: organizationIDs,
			}, &opts.OrganizationID, opts.IO.SurveyIO()); err != nil {
				return err
			}
		}
	}

	// Make a useful suggestion for the alias to use (subdomain) but omit the
	// sugesstion if a deployment with that alias is already configured.
	hostRef := firstSubDomain(opts.URL)
	if _, ok := opts.Config.Deployments[hostRef]; ok {
		hostRef = ""
	}

	// 5. Ask for an alias to use.
	if opts.Alias == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "Under which name should the deployment be referenced in the future?",
			Default: hostRef,
		}, &opts.Alias, survey.WithValidator(survey.ComposeValidators(
			survey.Required,
			survey.MinLength(3),
			surveyext.NotIn(opts.Config.DeploymentAliases()),
		)), opts.IO.SurveyIO()); err != nil {
			return err
		}
	}

	return nil
}

func runLogin(ctx context.Context, opts *loginOptions) error {
	// Read token from stdin, if no TTY is attached.
	if !opts.IO.IsStdinTTY() {
		var err error
		if opts.Token, err = readTokenFromStdIn(opts.IO.In()); err != nil {
			return err
		}
	}

	// If a deployment with the alias exists in the config, we ask the user if he
	// wants to overwrite it, if "--force" is not set. When no TTY is attached,
	// we abort and return, not overwritting anything.
	if _, ok := opts.Config.Deployments[opts.Alias]; ok && !opts.Force {
		if !opts.IO.IsStdinTTY() {
			return cmdutil.ErrSilent
		}

		msg := fmt.Sprintf("Deployment with alias %q already configured! Overwrite?", opts.Alias)
		if overwrite, err := surveyext.AskConfirm(msg, false, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	axiomClient, err := client.New(ctx, opts.URL, opts.Token, opts.OrganizationID, opts.Config.Insecure)
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

		if user != nil {
			if (client.IsCloudURL(opts.URL) || opts.Config.ForceCloud) && axiom.IsPersonalToken(opts.Token) {
				organization, err := axiomClient.Organizations.Selfhost.Get(ctx, opts.OrganizationID)
				if err != nil {
					return err
				}

				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s as %s\n",
					cs.SuccessIcon(), cs.Bold(organization.Name), cs.Bold(user.Name))
			} else {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s as %s\n",
					cs.SuccessIcon(), cs.Bold(opts.Alias), cs.Bold(user.Name))
			}
		} else {
			if client.IsCloudURL(opts.URL) || opts.Config.ForceCloud {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s %s\n",
					cs.SuccessIcon(), cs.Bold(opts.OrganizationID), cs.Red(cs.Bold("(ingestion/query only!)")))
			} else {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s %s\n",
					cs.SuccessIcon(), cs.Bold(opts.Alias), cs.Red(cs.Bold("(ingestion/query only!)")))
			}
		}
	}

	opts.Config.ActiveDeployment = opts.Alias
	opts.Config.Deployments[opts.Alias] = config.Deployment{
		URL:            opts.URL,
		Token:          opts.Token,
		OrganizationID: opts.OrganizationID,
	}

	return opts.Config.Write()
}

func firstSubDomain(s string) string {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return ""
	}

	var hostRef string
	hostRefParts := strings.Split(u.Host, ".")
	if len(hostRefParts) > 0 {
		hostRef = hostRefParts[0]
	}

	return strings.TrimLeft(hostRef, u.Scheme)
}
