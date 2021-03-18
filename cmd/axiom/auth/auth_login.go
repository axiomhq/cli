package auth

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	axiomClient "github.com/axiomhq/cli/internal/client"
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

	// Url of the deployment to authenticate with. If not supplied as a flag,
	// which is optional, the user will be asked for it.
	URL string
	// Alias of the deployment for future reference. If not supplied as a flag,
	// which is optional, the user will be asked for it.
	Alias string
	// Token of the user who wants to authenticate against the deployment. The
	// user will be asked for it unless "token-stdin" is set.
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
		Use:   "login [--url <deployment-url>] [(-a|--alias) <deployment-alias>] [(-o|--org-id) <organization-id>] [-f|--force]",
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

	cmd.Flags().StringVar(&opts.URL, "url", "", "Url of the deployment")
	cmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias of the deployment")
	cmd.Flags().StringVarP(&opts.OrganizationID, "org-id", "o", "", "Organization ID (only valid for Axiom Cloud)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("alias", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("token-stdin", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("org-id", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("url")
		_ = cmd.MarkFlagRequired("alias")
		_ = cmd.MarkFlagRequired("token-stdin")
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func completeLogin(ctx context.Context, opts *loginOptions) error {
	// 1. Cloud or Selfhost?
	var deploymentKind string
	if err := survey.AskOne(&survey.Select{
		Message: "Which kind of Axiom deployment are you using?",
		Options: validDeploymentTypes,
	}, &deploymentKind, opts.IO.SurveyIO()); err != nil {
		return err
	}

	// 2. If Cloud, set the correct URL instead of asking the user for it.
	if deploymentKind == typeCloud {
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
		Message: "What is your personal access or ingest token?",
	}, &opts.Token, survey.WithValidator(survey.ComposeValidators(
		survey.Required,
		surveyext.ValidateToken,
	)), opts.IO.SurveyIO()); err != nil {
		return err
	}

	// 4. Try to authenticate and fetch the organizations available to the user
	// in case a Personal Access Token was provided and the deployment is a
	// cloud deployment. If only one organization is available, that one is
	// selected by default, without asking the user for it.
	if axiomClient.IsPersonalToken(opts.Token) && deploymentKind == typeCloud && opts.OrganizationID == "" {
		client, err := axiomClient.New(opts.URL, opts.Token, "", opts.Config.Insecure)
		if err != nil {
			return err
		}

		organizations, err := client.Organizations.List(ctx)
		if err != nil {
			return err
		}

		if len(organizations) == 1 {
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
		)), opts.IO.SurveyIO()); err != nil {
			return err
		}
	}

	return nil
}

func runLogin(ctx context.Context, opts *loginOptions) error {
	// Read token from stdin, if no TTY is attached.
	if !opts.IO.IsStdinTTY() {
		// The token won't be longer.
		r := io.LimitReader(opts.IO.In(), 64)

		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		opts.Token = strings.TrimSuffix(string(contents), "\n")
		opts.Token = strings.TrimSuffix(opts.Token, "\r")
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

	client, err := axiomClient.New(opts.URL, opts.Token, opts.OrganizationID, opts.Config.Insecure)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var user *axiom.AuthenticatedUser
	if axiomClient.IsPersonalToken(opts.Token) {
		if user, err = client.Users.Current(ctx); err != nil {
			return err
		}
	} else if axiomClient.IsIngestToken(opts.Token) {
		if err = client.Tokens.Ingest.Validate(ctx); err != nil {
			return err
		}
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()

		if user != nil {
			if opts.URL == axiom.CloudURL {
				organization, err := client.Organizations.Get(ctx, opts.OrganizationID)
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
			if opts.URL == axiom.CloudURL {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to organization %s %s\n",
					cs.SuccessIcon(), cs.Bold(opts.OrganizationID), cs.Red(cs.Bold("(ingestion only!)")))
			} else {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to deployment %s %s\n",
					cs.SuccessIcon(), cs.Bold(opts.Alias), cs.Red(cs.Bold("(ingestion only!)")))
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
