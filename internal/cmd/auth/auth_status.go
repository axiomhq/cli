package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/utils"
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
		Use:   "status [<alias>]",
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
			cmdutil.AsksForSetup(f, NewLoginCmd(f)),
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
		statusInfo = map[string][]string{}
		// We don't care about the context here. If an errors occurs, still try
		// to get the status of the other deployments.
		eg, _ = errgroup.WithContext(ctx)
	)
	for _, deploymentAlias := range deploymentAliases {
		deploymentAlias := deploymentAlias
		deployment, ok := opts.Config.Deployments[deploymentAlias]
		if !ok {
			continue
		}

		eg.Go(func() error {
			var info string
			defer func() {
				statusInfo[deploymentAlias] = append(statusInfo[deploymentAlias], info)
			}()

			client, err := client.New(ctx, deployment.URL, deployment.Token, deployment.OrganizationID, opts.Config.Insecure)
			if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				return err
			}

			user, err := client.Users.Current(ctx)
			if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				return err
			}

			if deployment.OrganizationID == "" {
				info = fmt.Sprintf("%s Logged in as %s", cs.SuccessIcon(),
					cs.Bold(user.Name))
				return nil
			}

			organization, err := client.Organizations.Get(ctx, deployment.OrganizationID)
			if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				return err
			}

			info = fmt.Sprintf("%s Logged in to %s as %s", cs.SuccessIcon(),
				cs.Bold(organization.Name), cs.Bold(user.Name))

			return nil
		})
	}

	failed := false
	if err := eg.Wait(); err != nil {
		failed = true
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
		fmt.Fprint(opts.IO.ErrOut(), utils.Dedent(buf.String()))
	}

	if failed {
		return cmdutil.ErrSilent
	}

	return nil
}
