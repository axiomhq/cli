package auth

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/auth"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
)

const (
	oAuth2ClientID = "13c885a8-f46a-4424-82d2-883cf7ccfe49"

	typeCloud    = "Cloud"
	typeSelfhost = "Selfhost"
)

var validDeploymentTypes = []string{typeCloud, typeSelfhost}

type loginOptions struct {
	*cmdutil.Factory

	// AutoLogin specifies if the CLI redirects to the Axiom UI for
	// authentication.
	AutoLogin bool
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

// NewLoginCmd creates ans returns the login command.
func NewLoginCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login [(-t|--type)=cloud|selfhost] [(-u|--url) <url>] [(-a|--alias) <alias>] [(-o|--org-id) <organization-id>] [-f|--force]",
		Short: "Login to Axiom",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively authenticate against Axiom:
			$ axiom auth login
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_ACCESS_TOKEN | axiom auth login --alias="axiom-eu-west-1" --url="https://axiom.eu-west-1.aws.com" -f
		`),

		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if !opts.IO.IsStdinTTY() || opts.AutoLogin {
				return nil
			}

			// If the user specifies the url, we assume he wants to authenticate
			// against a selfhost deployment unless he explicitly specifies the
			// hidden type flag that specifies the type of the deployment.
			if cmd.Flag("url").Changed && !cmd.Flag("type").Changed {
				opts.Type = typeSelfhost
			}

			return completeLogin(cmd.Context(), opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.IO.IsStdinTTY() && opts.AutoLogin {
				return autoLogin(cmd.Context(), opts)
			}
			return runLogin(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.AutoLogin, "auto-login", true, "Login through the Axiom UI")
	cmd.Flags().StringVarP(&opts.Type, "type", "t", strings.ToLower(typeCloud), "Type of the deployment")
	cmd.Flags().StringVarP(&opts.URL, "url", "u", axiom.CloudURL, "Url of the deployment")
	cmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias of the deployment")
	cmd.Flags().StringVarP(&opts.OrganizationID, "org-id", "o", "", "Organization ID")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("auto-login", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("type", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("alias", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("org-id", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("alias")
	}

	_ = cmd.PersistentFlags().MarkHidden("type")

	return cmd
}

func completeLogin(ctx context.Context, opts *loginOptions) error {
	// 1. Cloud or Selfhost?
	if opts.Type == "" {
		if err := survey.AskOne(&survey.Select{
			Message: "Which kind of Axiom deployment are you using?",
			Default: validDeploymentTypes[0],
			Options: validDeploymentTypes,
		}, &opts.Type, opts.IO.SurveyIO()); err != nil {
			return err
		}
	}

	opts.Type = strings.ToLower(opts.Type)

	// 2. If Cloud mode but no URL, set the correct URL instead of asking the
	// user for it.
	if opts.Type == strings.ToLower(typeCloud) && opts.URL == "" {
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

	if opts.URL != "" && !strings.HasPrefix(opts.URL, "http://") && !strings.HasPrefix(opts.URL, "https://") {
		opts.URL = "https://" + opts.URL
	}

	// Suggest this URL to the user for creating a personal token.
	u, err := url.ParseRequestURI(opts.URL)
	if err != nil {
		return err
	}
	u.Path = "/profile"

	// 3. Wheather to open the browser or not.
	askTokenMsg := "What is your personal access token?"
	if ok, err := surveyext.AskConfirm("You need to retrieve a personal access token from your profile page. Should I open that page in your default browser?",
		true, opts.IO.SurveyIO()); err != nil {
		return err
	} else if !ok {
		askTokenMsg = fmt.Sprintf("What is your personal access token (create one over at %s)?", u.String())
	} else if err = browser.OpenURL(u.String()); err != nil {
		return err
	}

	// 3. The token to use.
	if err := survey.AskOne(&survey.Password{
		Message: askTokenMsg,
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
		axiomClient, err := client.New(ctx, opts.URL, opts.Token, "axiom", opts.Config.Insecure)
		if err != nil {
			return err
		}

		if organizations, err := axiomClient.Organizations.List(ctx); err != nil {
			return err
		} else if len(organizations) == 1 {
			opts.OrganizationID = organizations[0].ID
		} else {
			sort.Slice(organizations, func(i, j int) bool {
				return strings.ToLower(organizations[i].Name) < strings.ToLower(organizations[j].Name)
			})

			organizationNames := make([]string, len(organizations))
			for i, organization := range organizations {
				organizationNames[i] = organization.Name
			}

			var organizationName string
			if err := survey.AskOne(&survey.Select{
				Message: "Which organization to use?",
				Options: organizationNames,
				Default: organizationNames[0],
				Description: func(_ string, idx int) string {
					return organizations[idx].ID
				},
			}, &organizationName, opts.IO.SurveyIO()); err != nil {
				return err
			}

			for i, organization := range organizations {
				if organization.Name == organizationName {
					opts.OrganizationID = organizations[i].ID
					break
				}
			}
		}
	}

	// Make a useful suggestion for the alias to use (subdomain) but omit the
	// sugesstion if a deployment with that alias is already configured. Cut the
	// port, if present.
	hostRef := firstSubDomain(opts.URL)
	if _, ok := opts.Config.Deployments[hostRef]; ok {
		hostRef = ""
	}

	// Just use "cloud" as the alias if this is their first deployment and they
	// are authenticating against Axiom Cloud.
	if hostRef == strings.ToLower(typeCloud) {
		opts.Alias = hostRef
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

func autoLogin(ctx context.Context, opts *loginOptions) error {
	opts.Type = strings.ToLower(opts.Type)

	// 1. If Cloud mode but no URL, set the correct URL instead of asking the
	// user for it.
	if opts.Type == strings.ToLower(typeCloud) && opts.URL == "" {
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

	if opts.URL != "" && !strings.HasPrefix(opts.URL, "http://") && !strings.HasPrefix(opts.URL, "https://") {
		opts.URL = "https://" + opts.URL
	}

	// 2. Wheather to open the browser or not. But the URL to open and have the
	// user login is presented nonetheless.
	stop := func() {}
	loginFunc := func(_ context.Context, loginURL string) error {
		if ok, err := surveyext.AskConfirm("You need to login to Axiom. Should I open your default browser?",
			true, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !ok {
			fmt.Fprintf(opts.IO.ErrOut(), "Please open %s in your browser, manually.\n", loginURL)
		} else if err = browser.OpenURL(loginURL); err != nil {
			return err
		}

		fmt.Fprintln(opts.IO.ErrOut(), "Waiting for authentication...")

		stop = opts.IO.StartActivityIndicator()

		return nil
	}
	defer stop()

	// Wait five minutes before timing out.
	authContext, authCancel := context.WithTimeout(ctx, time.Minute*5)
	defer authCancel()

	var err error
	if opts.Token, err = auth.Login(authContext, oAuth2ClientID, opts.URL, loginFunc); err != nil {
		return err
	}

	// 3. Try to authenticate and fetch the organizations available to the user
	// in case the deployment is a cloud deployment. If at least one
	// organization is available, the first one is selected.
	if opts.Type == strings.ToLower(typeCloud) && opts.OrganizationID == "" {
		axiomClient, err := client.New(ctx, opts.URL, opts.Token, "axiom", opts.Config.Insecure)
		if err != nil {
			return err
		}

		organizations, err := axiomClient.Organizations.List(ctx)
		if err != nil {
			return err
		}

		if len(organizations) > 0 {
			opts.OrganizationID = organizations[0].ID
		}
	}

	stop()

	// Make a useful suggestion for the alias to use (subdomain) but omit the
	// sugesstion if a deployment with that alias is already configured. Cut the
	// port, if present.
	hostRef := firstSubDomain(opts.URL)
	if _, ok := opts.Config.Deployments[hostRef]; ok {
		hostRef = ""
	}

	// Just use "cloud" as the alias if this is their first deployment and they
	// are authenticating against Axiom Cloud.
	if hostRef == strings.ToLower(typeCloud) {
		opts.Alias = hostRef
	}

	// 4. Ask for an alias to use.
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

	// 5. Try to login with the retrieved credentials.
	return runLogin(ctx, opts)
}

func runLogin(ctx context.Context, opts *loginOptions) error {
	// Read token from stdin, if no TTY is attached.
	if !opts.IO.IsStdinTTY() {
		var err error
		if opts.Token, err = readTokenFromStdIn(opts.IO.In()); err != nil {
			return err
		}
	}

	// If a deployment with the alias exists in the config, we ask the user if
	// he wants to overwrite it, if "--force" is not set. When no TTY is
	// attached, we abort and return, not overwritting anything.
	if _, ok := opts.Config.Deployments[opts.Alias]; ok && !opts.Force {
		if !opts.IO.IsStdinTTY() {
			return fmt.Errorf("deployment with alias %q already configured, overwrite with '-f|--force' flag", opts.Alias)
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
				organization, err := axiomClient.Organizations.Get(ctx, opts.OrganizationID)
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
	hostRefParts := strings.Split(u.Hostname(), ".")
	if len(hostRefParts) > 0 {
		hostRef = hostRefParts[0]
	}

	return strings.TrimLeft(hostRef, u.Scheme)
}
