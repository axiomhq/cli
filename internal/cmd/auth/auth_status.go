package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/muesli/reflow/dedent"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
)

type statusOptions struct {
	*cmdutil.Factory

	// Alias of the deployment to check the authentication status for.
	Alias string
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &statusOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "status [<deployment-alias>]",
		Short: "View authentication status",

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Alias),
		ValidArgsFunction: deploymentCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Check authentication status of all configured deployments:
			$ axiom auth status
			
			# Check authentication status of a specified deployment:
			$ axiom auth status axiom-eu-west-1
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsDeployments(f),
			cmdutil.NeedsValidDeployment(f, &opts.Alias),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), opts)
		},
	}

	return cmd
}

func runStatus(ctx context.Context, opts *statusOptions) error {
	deploymentAliases := opts.Config.DeploymentAliases()
	if opts.Alias != "" {
		deploymentAliases = []string{opts.Alias}
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var (
		cs         = opts.IO.ColorScheme()
		failed     bool
		statusInfo = map[string][]string{}
	)
	for _, v := range deploymentAliases {
		deployment, ok := opts.Config.Deployments[v]
		if !ok {
			continue
		}

		client, err := client.New(ctx, deployment.URL, deployment.Token, deployment.OrganizationID, opts.Config.Insecure)
		if err != nil {
			return err
		}

		var info string
		if axiom.IsPersonalToken(deployment.Token) {
			var user *axiom.AuthenticatedUser
			if user, err = client.Users.Current(ctx); errors.Is(err, axiom.ErrUnauthenticated) {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), "Invalid credentials")
				failed = true
			} else if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				failed = true
			} else {
				if deployment.OrganizationID == "" {
					info = fmt.Sprintf("%s Logged in as %s", cs.SuccessIcon(),
						cs.Bold(user.Name))
				} else {
					var organization *axiom.Organization
					if organization, err = client.Organizations.Get(ctx, deployment.OrganizationID); err != nil {
						info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
						failed = true
					} else {
						info = fmt.Sprintf("%s Logged in to %s as %s", cs.SuccessIcon(),
							cs.Bold(organization.Name), cs.Bold(user.Name))
					}
				}
			}
		} else {
			if err = client.Tokens.Ingest.Validate(ctx); errors.Is(err, axiom.ErrUnauthenticated) {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), "Invalid credentials")
				failed = true
			} else if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				failed = true
			} else {
				if deployment.OrganizationID == "" {
					info = fmt.Sprintf("%s Using ingest token", cs.WarningIcon())
				} else {
					info = fmt.Sprintf("%s Logged in to %s using ingest token", cs.WarningIcon(),
						cs.Bold(deployment.OrganizationID))
				}
			}
		}

		statusInfo[v] = append(statusInfo[v], info)
	}

	stop()

	if opts.IO.IsStderrTTY() {
		var buf strings.Builder
		for _, alias := range deploymentAliases {
			if alias == opts.Config.ActiveDeployment {
				fmt.Fprintf(&buf, "%s %s\n", cs.Yellow("➜"), cs.Bold(alias))
			} else {
				fmt.Fprintf(&buf, "  %s\n", cs.Bold(alias))
			}
			for _, line := range statusInfo[alias] {
				fmt.Fprintf(&buf, "    %s\n", line)
			}
		}
		fmt.Fprint(opts.IO.ErrOut(), dedent.String(buf.String()))
	}

	if failed {
		return cmdutil.ErrSilent
	}

	return nil
}
