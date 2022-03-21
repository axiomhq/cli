package token

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/surveyext"
)

type createOptions struct {
	*cmdutil.Factory

	// Name of the token to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Name string `survey:"name"`
	// Description of the token to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Description string `survey:"description"`
	// Scopes of the token to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Scopes []string `survey:"scopes"`
	// Permissions of the token to create. If not supplied as a flag, which is
	// optional, the user will be asked for it.
	Permissions []string `survey:"permissions"`

	tokenType string
}

func newCreateCmd(f *cmdutil.Factory, tokenType string) *cobra.Command {
	opts := &createOptions{
		Factory:   f,
		tokenType: tokenType,
	}

	cmd := &cobra.Command{
		Use:   "create [(-n|--name) <token-name>] [(-d|--description) <token-description>]",
		Short: "Create a token",

		Aliases: []string{"new"},

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively create a token:
			$ axiom token ` + tokenType + ` create
			
			# Create a token and provide the parameters on the command-line:
			$ axiom token ` + tokenType + ` create --name=ingest-all --description="Ingest token for all datasets" --scope="*"
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeCreate(cmd.Context(), opts); err != nil {
				return err
			}
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of the token")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description of the token")
	cmd.Flags().StringSliceVarP(&opts.Scopes, "scope", "s", nil, "Scope(s) of the token (for api tokens). Dataset name or '*' for all datasets.")
	cmd.Flags().StringSliceVarP(&opts.Permissions, "permission", "p", nil, "Permission(s) of the token (for api tokens)")

	_ = cmd.RegisterFlagCompletionFunc("name", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("description", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("scope", scopeCompletionFunc(f))
	_ = cmd.RegisterFlagCompletionFunc("permission", permissionCompletion)

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("name")
		_ = cmd.MarkFlagRequired("description")
	}

	return cmd
}

func completeCreate(ctx context.Context, opts *createOptions) error {
	questions := make([]*survey.Question, 0, 4)

	tokenNames, err := getTokenNames(ctx, opts.Factory, opts.tokenType)
	if err != nil {
		return err
	}

	if opts.Name == "" {
		questions = append(questions, &survey.Question{
			Name:   "name",
			Prompt: &survey.Input{Message: "What is the name of the token?"},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(3),
				surveyext.NotIn(tokenNames),
			),
		})
	}

	if opts.Description == "" {
		questions = append(questions, &survey.Question{
			Name:   "description",
			Prompt: &survey.Input{Message: "What is the description of the token?"},
		})
	}

	if opts.tokenType == typeAPI && len(opts.Scopes) == 0 {
		datasetNames, err := getDatasetNames(ctx, opts.Factory)
		if err != nil {
			return err
		}

		if len(datasetNames) > 0 {
			questions = append(questions, &survey.Question{
				Name: "scopes",
				Prompt: &survey.MultiSelect{
					Message: "What scopes should the token have?",
					Options: append([]string{"*"}, datasetNames...),
				},
			})
		} else {
			opts.Scopes = []string{"*"}
		}
	}

	if opts.tokenType == typeAPI && len(opts.Permissions) == 0 {
		questions = append(questions, &survey.Question{
			Name: "permissions",
			Prompt: &survey.MultiSelect{
				Message: "What permissions should the token have?",
				Options: validPermissions,
			},
		})
	}

	return survey.Ask(questions, opts, opts.IO.SurveyIO())
}

func runCreate(ctx context.Context, opts *createOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var (
		createFunc func(context.Context, axiom.TokenCreateUpdateRequest) (*axiom.Token, error)
		viewFunc   func(context.Context, string) (*axiom.RawToken, error)
	)
	switch opts.tokenType {
	case typeAPI:
		createFunc = client.Tokens.API.Create
		viewFunc = client.Tokens.API.View
	case typePersonal:
		createFunc = client.Tokens.Personal.Create
		viewFunc = client.Tokens.Personal.View
	default:
		return fmt.Errorf("unknown token type: %s", opts.tokenType)
	}

	// Translate the permission string to the `axiom.Permission` type.
	permissions := make([]axiom.Permission, len(opts.Permissions))
	for k, v := range opts.Permissions {
		if permissions[k], err = permissionFromString(v); err != nil {
			return err
		}
	}

	token, err := createFunc(ctx, axiom.TokenCreateUpdateRequest{
		Name:        opts.Name,
		Description: opts.Description,
		Scopes:      opts.Scopes,
		Permissions: permissions,
	})
	if err != nil {
		return err
	}

	rawToken, err := viewFunc(ctx, token.ID)
	if err != nil {
		return err
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Created token %s:\n\n%s\n",
			cs.SuccessIcon(), cs.Bold(opts.Name), rawToken.Token)
	}

	return nil
}

func permissionFromString(s string) (permission axiom.Permission, err error) {
	switch s {
	// case emptyPermission.String():
	// 	permission = emptyPermission
	case axiom.CanIngest.String():
		permission = axiom.CanIngest
	case axiom.CanQuery.String():
		permission = axiom.CanQuery
	default:
		err = fmt.Errorf("unknown permission %q", s)
	}

	return permission, err
}
