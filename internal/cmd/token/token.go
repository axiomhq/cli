package token

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

//  All valid token types.
const (
	typeAPI      = "api"
	typePersonal = "personal"
)

var validPermissions = []string{
	axiom.CanIngest.String(),
	axiom.CanQuery.String(),
}

// NewTokenCmd creates and returns the token command.
func NewTokenCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token <type> <command>",
		Short: "Manage tokens",
		Long:  "Manage tokens.",

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
		),
	}

	cmd.AddCommand(newTokenCmd(f, typeAPI))
	cmd.AddCommand(newTokenCmd(f, typePersonal))

	return cmd
}

func newTokenCmd(f *cmdutil.Factory, tokenType string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   tokenType + " <command>",
		Short: "Manage " + tokenType + " tokens",
		Long:  "Manage " + tokenType + " tokens.",

		Example: heredoc.Doc(`
			$ axiom token ` + tokenType + ` create --name=my-token
			$ axiom token ` + tokenType + ` list
			$ axiom token ` + tokenType + ` delete my-token
		`),
	}

	cmd.AddCommand(newCreateCmd(f, tokenType))
	cmd.AddCommand(newDeleteCmd(f, tokenType))

	return cmd
}

func getDatasetNames(ctx context.Context, f *cmdutil.Factory) ([]string, error) {
	client, err := f.Client(ctx)
	if err != nil {
		return nil, err
	}

	stop := f.IO.StartActivityIndicator()
	defer stop()

	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		return nil, err
	}

	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	return datasetNames, nil
}

func scopeCompletionFunc(f *cmdutil.Factory) cmdutil.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Just complete the first argument.
		if len(args) > 0 {
			return cmdutil.NoCompletion(cmd, args, toComplete)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		client, err := f.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		datasets, err := client.Datasets.List(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		res := make([]string, 0, len(datasets)+1)
		for _, dataset := range datasets {
			if strings.HasPrefix(dataset.Name, toComplete) {
				res = append(res, dataset.Name)
			}
		}
		res = append(res, "*")

		return res, cobra.ShellCompDirectiveNoFileComp
	}
}

func permissionCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(validPermissions))
	for _, permission := range validPermissions {
		if strings.HasPrefix(permission, toComplete) {
			res = append(res, permission)
		}
	}

	return res, cobra.ShellCompDirectiveNoFileComp
}

func tokenCompletionFunc(f *cmdutil.Factory, tokenType string) cmdutil.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Just complete the first argument.
		if len(args) > 0 {
			return cmdutil.NoCompletion(cmd, args, toComplete)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		client, err := f.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var listFunc func(ctx context.Context) ([]*axiom.Token, error)
		switch tokenType {
		case typeAPI:
			listFunc = client.Tokens.API.List
		case typePersonal:
			listFunc = client.Tokens.Personal.List
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		tokens, err := listFunc(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		res := make([]string, 0, len(tokens))
		for _, token := range tokens {
			if strings.HasPrefix(token.Name, toComplete) {
				res = append(res, token.Name)
			}
		}

		return res, cobra.ShellCompDirectiveNoFileComp
	}
}

func getTokenNames(ctx context.Context, f *cmdutil.Factory, tokenType string) ([]string, error) {
	client, err := f.Client(ctx)
	if err != nil {
		return nil, err
	}

	var listFunc func(context.Context) ([]*axiom.Token, error)
	switch tokenType {
	case typeAPI:
		listFunc = client.Tokens.API.List
	case typePersonal:
		listFunc = client.Tokens.Personal.List
	default:
		return nil, fmt.Errorf("unknown token type: %s", tokenType)
	}

	stop := f.IO.StartActivityIndicator()
	defer stop()

	tokens, err := listFunc(ctx)
	if err != nil {
		return nil, err
	}

	stop()

	tokenNames := make([]string, len(tokens))
	for i, token := range tokens {
		tokenNames[i] = token.Name
	}

	return tokenNames, nil
}

func getTokenIDFromName(ctx context.Context, client *axiom.Client, tokenType, name string) (string, error) {
	var listFunc func(context.Context) ([]*axiom.Token, error)
	switch tokenType {
	case typeAPI:
		listFunc = client.Tokens.API.List
	case typePersonal:
		listFunc = client.Tokens.Personal.List
	default:
		return "", fmt.Errorf("unknown token type: %s", tokenType)
	}

	tokens, err := listFunc(ctx)
	if err != nil {
		return "", err
	}

	for _, token := range tokens {
		if token.Name == name {
			return token.ID, nil
		}
	}

	return "", fmt.Errorf("could not find id for token %s", name)
}
