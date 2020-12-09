package auth

import (
	"context"
	"fmt"
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

var validTokenTypes = []string{string(config.Personal), string(config.Ingest)}

type loginOptions struct {
	*cmdutil.Factory

	// Url of the deployment to authenticate with. If not supplied as a flag,
	// which is optional, the user will be asked for it.
	URL string
	// Alias of the deployment for future reference. If not supplied as a flag,
	// which is optional, the user will be asked for it.
	Alias string `survey:"alias"`
	// Token of the user who wants to authenticate against the deployment. The
	// user will be asked for it unless "token-stdin" is set.
	Token string `survey:"token"`
	// TokenType of the supplied token. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	TokenType string `survey:"tokenType"`
	// Read token from stdin instead of prompting the user for it.
	TokenStdIn bool
	// Force the creation and skip the confirmation prompt.
	Force bool
}

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login [--url <deployment-url>] [(-a|--alias) <deployment-alias>] [(-t|--token-type=)personal|ingest] [--token-stdin] [-f|--force]",
		Short: "Login to an Axiom deployment",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively authenticate against an Axiom deployment:
			$ axiom auth login
			
			# Provide parameters on the command-line:
			$ echo $AXIOM_ACCESS_TOKEN | axiom auth login --url="https://axiom.eu-west-1.aws.com" --alias="axiom-eu-west-1" --token-stdin
		`),

		PreRunE: func(*cobra.Command, []string) error {
			if !opts.IO.IsStdinTTY() && !opts.TokenStdIn {
				return cmdutil.NewFlagErrorf("--token-stdin required when not running interactively")
			} else if opts.TokenStdIn {
				if opts.URL == "" {
					return cmdutil.NewFlagErrorf("--url required when --token-stdin is set")
				} else if opts.Alias == "" {
					return cmdutil.NewFlagErrorf("--alias required when --token-stdin is set")
				} else if opts.TokenType == "" {
					return cmdutil.NewFlagErrorf("--token-type required when --token-stdin is set")
				}
				return nil
			}
			return completeLogin(opts)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.TokenType != config.Personal && opts.TokenType != config.Ingest {
				return fmt.Errorf("unknown token type %q (choose %q or %q)",
					opts.TokenType, config.Personal, config.Ingest)
			}
			return runLogin(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.URL, "url", "", "Url of the deployment")
	cmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias of the deployment")
	cmd.Flags().BoolVar(&opts.TokenStdIn, "token-stdin", false, "Read token from stdin")
	cmd.Flags().StringVarP(&opts.TokenType, "token-type", "t", "", "Type of the token")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("alias", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("token-stdin", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("token-type", tokenTypeCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("url")
		_ = cmd.MarkFlagRequired("alias")
		_ = cmd.MarkFlagRequired("token-stdin")
		_ = cmd.MarkFlagRequired("token-type")
		_ = cmd.MarkFlagRequired("force")
	}

	return cmd
}

func completeLogin(opts *loginOptions) error {
	var res string
	if err := survey.AskOne(&survey.Select{
		Message: "Which kind of Axiom deployment are you using?",
		Options: []string{"Cloud", "Selfhost"},
	}, &res); err != nil {
		return err
	}
	if res == "Cloud" {
		opts.URL = axiom.CloudURL
	}

	if opts.URL == "" {
		if err := survey.AskOne(&survey.Input{Message: "What is the url of the deployment?"},
			&opts.URL,
			survey.WithValidator(survey.ComposeValidators(
				survey.Required,
				surveyext.ValidateURL,
			)),
		); err != nil {
			return err
		}
		if res == "Cloud" {
			opts.URL = axiom.CloudURL
		}
	}

	// Make a useful suggestion for the alias to use (subdomain) but omit the
	// sugesstion if a deployment with that alias is already configured.
	hostRef := firstSubDomain(opts.URL)
	if _, ok := opts.Config.Deployments[hostRef]; ok {
		hostRef = ""
	}

	questions := make([]*survey.Question, 0, 3)

	if opts.Alias == "" {
		questions = append(questions, &survey.Question{
			Name: "alias",
			Prompt: &survey.Input{
				Message: "Under which name should the deployment be referenced in the future?",
				Default: hostRef,
			},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(5),
			),
		})
	}

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

func runLogin(ctx context.Context, opts *loginOptions) error {
	// Read token from stdin, if the appropriate option is set.
	if opts.TokenStdIn {
		contents, err := ioutil.ReadAll(opts.IO.In())
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

	if opts.TokenType == config.Personal {
		client, err := axiomClient.New(opts.URL, opts.Token)
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
				cs.SuccessIcon(), cs.Bold(opts.Alias), opts.URL, cs.Bold(user.Name))
		}
	}

	opts.Config.Deployments[opts.Alias] = config.Deployment{
		URL:       opts.URL,
		Token:     opts.Token,
		TokenType: opts.TokenType,
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

func tokenTypeCompletion(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return validTokenTypes, cobra.ShellCompDirectiveNoFileComp
}
