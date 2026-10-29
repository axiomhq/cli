package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
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

	var completionFunc cmdutil.CompletionFunc
	if f != nil && f.Config != nil {
		completionFunc = deploymentCompletionFunc(f.Config)
	}

	cmd := &cobra.Command{
		Use:   "status [<alias>]",
		Short: "View authentication status",

		DisableFlagsInUseLine: true,

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completionFunc,

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
			if len(args) > 0 {
				opts.Alias = args[0]
			}
			return runStatus(cmd.Context(), opts)
		},
	}

	return cmd
}

func runStatus(ctx context.Context, opts *statusOptions) error {
	if opts == nil || opts.Config == nil {
		return errors.New("no configuration provided")
	}

	if opts.IO == nil {
		opts.IO = terminal.NewIO()
	}

	cs := opts.IO.ColorScheme()

	deployments := opts.Config.Deployments
	if deployments == nil {
		deployments = make(map[string]config.Deployment)
	}

	deploymentAliases := opts.Config.DeploymentAliases()
	if len(deploymentAliases) == 0 {
		if activeDep, hasActive := opts.Config.GetActiveDeployment(); hasActive && activeDep.Token != "" {
			alias := opts.Config.ActiveDeployment
			if alias == "" {
				alias = "default"
			}
			deployments = map[string]config.Deployment{alias: activeDep}
			deploymentAliases = []string{alias}
		} else {
			fmt.Fprintf(opts.IO.ErrOut(), "%s No deployments configured!\n\n  Setup a deployment by logging into it:\n  $ %s\n",
				cs.ErrorIcon(), cs.Bold("axiom auth login"))
			return cmdutil.ErrSilent
		}
	}

	if opts.Alias != "" {
		if _, ok := deployments[opts.Alias]; !ok {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Chosen deployment %s is not configured!\n",
				cs.ErrorIcon(), cs.Bold(opts.Alias))
			return cmdutil.ErrSilent
		}
		deploymentAliases = []string{opts.Alias}
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var (
		statusInfo   = map[string][]string{}
		statusInfoMu = new(sync.Mutex)
		// We don't care about the context here. If an error occurs, still try
		// to get the status of the other deployments.
		eg, _ = errgroup.WithContext(ctx)
	)
	for _, deploymentAlias := range deploymentAliases {
		deploymentAlias := deploymentAlias
		deployment, ok := deployments[deploymentAlias]
		if !ok {
			continue
		}

		eg.Go(func() error {
			var info string
			defer func() {
				statusInfoMu.Lock()
				defer statusInfoMu.Unlock()

				statusInfo[deploymentAlias] = append(statusInfo[deploymentAlias], info)
			}()

			if deployment.Token == "" {
				info = fmt.Sprintf("%s Not authenticated: missing token", cs.ErrorIcon())
				return errors.New("missing token")
			}

			client, err := client.New(ctx, deployment.URL, deployment.Token, deployment.OrganizationID, "", "", opts.Config.Insecure)
			if err != nil {
				if errors.Is(err, axiom.ErrUnauthenticated) {
					info = fmt.Sprintf("%s Invalid credentials", cs.ErrorIcon())
				} else {
					info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				}
				return err
			}

			if client == nil {
				info = fmt.Sprintf("%s Failed to initialize client", cs.ErrorIcon())
				return errors.New("nil client")
			}

			if client.Users == nil {
				info = fmt.Sprintf("%s User service unavailable", cs.ErrorIcon())
				return errors.New("nil user service")
			}

			user, err := client.Users.Current(ctx)
			if err != nil {
				if errors.Is(err, axiom.ErrUnauthenticated) {
					info = fmt.Sprintf("%s Invalid credentials", cs.ErrorIcon())
				} else {
					info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				}
				return err
			}

			if user == nil {
				info = fmt.Sprintf("%s Unable to fetch user details", cs.ErrorIcon())
				return errors.New("nil user")
			}

			userName := user.Name
			if userName == "" {
				userName = user.Email
			}
			if userName == "" {
				userName = user.ID
			}

			if deployment.OrganizationID == "" {
				info = fmt.Sprintf("%s Logged in as %s", cs.SuccessIcon(),
					cs.Bold(userName))
				return nil
			}

			if client.Organizations == nil {
				info = fmt.Sprintf("%s Organization service unavailable", cs.ErrorIcon())
				return errors.New("nil organization service")
			}

			organization, err := client.Organizations.Get(ctx, deployment.OrganizationID)
			if err != nil {
				if errors.Is(err, axiom.ErrUnauthenticated) {
					info = fmt.Sprintf("%s Invalid credentials", cs.ErrorIcon())
				} else {
					info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				}
				return err
			}

			if organization == nil {
				info = fmt.Sprintf("%s Unable to fetch organization details", cs.ErrorIcon())
				return errors.New("nil organization")
			}

			orgName := organization.Name
			if orgName == "" {
				orgName = organization.ID
			}

			info = fmt.Sprintf("%s Logged in to %s as %s", cs.SuccessIcon(),
				cs.Bold(orgName), cs.Bold(userName))

			return nil
		})
	}

	failed := false
	if err := eg.Wait(); err != nil {
		failed = true
	}

	stop()

	var buf strings.Builder
	for _, alias := range deploymentAliases {
		if opts.IO.IsStdoutTTY() {
			if alias == opts.Config.ActiveDeployment {
				fmt.Fprintf(&buf, "%s %s\n", cs.Yellow("➜"), cs.Bold(alias))
			} else {
				fmt.Fprintf(&buf, "  %s\n", cs.Bold(alias))
			}
			for _, line := range statusInfo[alias] {
				fmt.Fprintf(&buf, "    %s\n", line)
			}
		} else {
			fmt.Fprintf(&buf, "%s\n", alias)
			for _, line := range statusInfo[alias] {
				fmt.Fprintf(&buf, "  %s\n", line)
			}
		}
	}
	fmt.Fprint(opts.IO.Out(), utils.Dedent(buf.String()))

	if failed {
		return cmdutil.ErrSilent
	}

	return nil
}
