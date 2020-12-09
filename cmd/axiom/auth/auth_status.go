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

	axiomClient "github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
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

		var info string
		if deployment.TokenType == config.Personal {
			client, err := axiomClient.New(deployment.URL, deployment.Token)
			if err != nil {
				return err
			}

			user, err := client.Users.Current(ctx)
			if errors.Is(err, axiom.ErrUnauthenticated) {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), "Invalid credentials")
				failed = true
			} else if err != nil {
				info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
				failed = true
			} else {
				info = fmt.Sprintf("%s Logged in as %s (%s)", cs.SuccessIcon(),
					cs.Bold(user.Name), user.Emails[0])
			}
		} else {
			// We cannot validate ingest tokens without actually ingesting.
			info = fmt.Sprintf("%s Using ingest token", cs.WarningIcon())
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
