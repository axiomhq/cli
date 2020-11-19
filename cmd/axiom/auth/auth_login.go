package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type loginOptions struct {
	*cmdutil.Factory

	// Url of the backend to authenticate with. If not supplied as an argument,
	// which is optional, the user will be asked for it.
	URL string `survey:"url"`
	// Alias of the backend for future reference. If not supplied as an
	// argument, which is optional, the user will be asked for it.
	Alias string `survey:"alias"`
	// Username of the user who wants to authenticate against the backend. If
	// not supplied a an argument, which is optional, the user will be asked for
	// it.
	Username string `survey:"username"`
	// Password of the user who wants to authenticate against the backend. The
	// user will be asked for it unless "password-stdin" is set.
	Password string `survey:"password"`
	// Read password from stdin instead of prompting the user for it.
	PasswordStdIn bool
	// Force the creation and skip the confirmation prompt.
	Force bool
}

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "login [--url <backend-url>] [(-a|--alias) <backend-alias>] [(-u|--username) <username|email>] [--password-stdin] [-f|--force]",
		Short: "Login to an Axiom instance",

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively authenticate against an Axiom backend:
			$ axiom auth login
			
			# Provide parameters on the command-line:
			$ echo $MY_AXIOM_PASSWORD | axiom auth login --url="axiom.example.com" --alias="my-axiom" --username="lukas@axiom.co" --password-stdin
		`),

		PreRunE: func(*cobra.Command, []string) error {
			if !opts.IO.IsStdinTTY() && !opts.PasswordStdIn {
				return cmdutil.NewFlagErrorf("--password-stdin required when not running interactively")
			} else if opts.PasswordStdIn {
				if opts.URL == "" {
					return cmdutil.NewFlagErrorf("--url required when --password-stdin is set")
				} else if opts.Alias == "" {
					return cmdutil.NewFlagErrorf("--alias required when --password-stdin is set")
				} else if opts.Username == "" {
					return cmdutil.NewFlagErrorf("--username required when --password-stdin is set")
				}
				return nil
			}
			return completeLogin(opts)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.URL, "url", "", "Url of the backend")
	cmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias of this backend")
	cmd.Flags().StringVarP(&opts.Username, "username", "u", "", "Username to authenticate with")
	cmd.Flags().BoolVar(&opts.PasswordStdIn, "password-stdin", false, "Read password from stdin")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip the confirmation prompt")

	_ = cmd.RegisterFlagCompletionFunc("url", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("alias", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("username", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("password-stdin", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("force", cmdutil.NoCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("url")
		_ = cmd.MarkFlagRequired("alias")
		_ = cmd.MarkFlagRequired("username")
		_ = cmd.MarkFlagRequired("password-stdin")
	}

	return cmd
}

func completeLogin(opts *loginOptions) error {
	questions := make([]*survey.Question, 0, 4)

	if opts.URL == "" {
		questions = append(questions, &survey.Question{
			Name:   "url",
			Prompt: &survey.Input{Message: "What is the url of the backend?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				surveyext.ValidateURL,
			),
		})
	}

	if opts.Alias == "" {
		questions = append(questions, &survey.Question{
			Name:   "alias",
			Prompt: &survey.Input{Message: "Under which name should the backend be referenced in the future?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(5),
			),
		})
	}

	if opts.Username == "" {
		questions = append(questions, &survey.Question{
			Name:   "username",
			Prompt: &survey.Input{Message: "What is your username?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(5),
			),
		})
	}

	if opts.Password == "" {
		questions = append(questions, &survey.Question{
			Name:   "password",
			Prompt: &survey.Password{Message: "What is your password?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(5),
			),
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runLogin(ctx context.Context, opts *loginOptions) error {
	// Read password from stdin, if the appropriate option is set.
	if opts.PasswordStdIn {
		contents, err := ioutil.ReadAll(opts.IO.In())
		if err != nil {
			return err
		}

		opts.Password = strings.TrimSuffix(string(contents), "\n")
		opts.Password = strings.TrimSuffix(opts.Password, "\r")
	}

	// If a backend with the alias exists in the config, we ask the user if he
	// wants to overwrite it, if "--force" is not set. When no TTY is attached,
	// we abort and return, not overwritting anything.
	if _, ok := opts.Config.Backends[opts.Alias]; ok && !opts.Force {
		if !opts.IO.IsStdinTTY() {
			return cmdutil.ErrSilent
		}

		msg := fmt.Sprintf("Backend with alias %q already configured! Overwrite?", opts.Alias)
		if overwrite, err := surveyext.AskConfirm(msg, opts.IO.SurveyIO()); err != nil {
			return err
		} else if !overwrite {
			return cmdutil.ErrSilent
		}
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	client, err := axiom.NewClient(opts.URL, opts.Password)
	if err != nil {
		return err
	}

	valid, err := client.Authentication.Valid(ctx)
	if err != nil {
		return err
	} else if !valid {
		if opts.IO.IsStderrTTY() {
			cs := opts.IO.ColorScheme()
			fmt.Fprintf(opts.IO.ErrOut(), "%s Invalid authentication credentials\n",
				cs.ErrorIcon())
		}
		return cmdutil.ErrSilent
	}

	opts.Config.Backends[opts.Alias] = config.Backend{
		URL:      opts.URL,
		Username: opts.Username,
		Token:    opts.Password,
	}

	if err := opts.Config.Write(); err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Logged in to backend %s (%s) as %s\n",
			cs.SuccessIcon(), cs.Bold(opts.Alias), opts.URL, cs.Bold(opts.Username))
	}

	return nil
}
