package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go"
	"github.com/muesli/reflow/dedent"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

type statusOptions struct {
	*cmdutil.Factory

	// Alias of the backend to check the authentication status for.
	Alias string
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &statusOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "status [<backend-alias>]",
		Short: "View authentication status",

		DisableFlagsInUseLine: true,

		Args: cmdutil.ChainPositionalArgs(
			cobra.MaximumNArgs(1),
			cmdutil.PopulateFromArgs(f, &opts.Alias),
		),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Check authentication status of all configured backends:
			$ axiom auth status
			
			# Check authentication status of a specified backend:
			$ axiom auth status my-axiom
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsBackends(f),
			cmdutil.NeedsValidBackend(f, &opts.Alias),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), opts)
		},
	}

	return cmd
}

func runStatus(ctx context.Context, opts *statusOptions) error {
	backendAliases := opts.Config.BackendAliases()
	if opts.Alias != "" {
		backendAliases = []string{opts.Alias}
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var (
		cs         = opts.IO.ColorScheme()
		failed     bool
		statusInfo = map[string][]string{}
	)
	for _, v := range backendAliases {
		backend, ok := opts.Config.Backends[v]
		if !ok {
			continue
		}

		client, err := axiom.NewClient(backend.URL, backend.Token)
		if err != nil {
			return err
		}

		var info string
		if valid, err := client.Authentication.Valid(ctx); err != nil {
			info = fmt.Sprintf("%s %s", cs.ErrorIcon(), err)
			failed = true
		} else if !valid {
			info = fmt.Sprintf("%s Invalid authentication credentials",
				cs.ErrorIcon())
			failed = true
		} else {
			info = fmt.Sprintf("%s Logged in to backend %s as %s",
				cs.SuccessIcon(), backend.URL, cs.Bold(backend.Username))
		}

		statusInfo[v] = append(statusInfo[v], info)
	}

	stop()

	if opts.IO.IsStderrTTY() {
		var buf strings.Builder
		for _, alias := range backendAliases {
			if alias == opts.Config.ActiveBackend {
				fmt.Fprintf(&buf, "%s %s\n", cs.Yellow("‣"), cs.Bold(alias))
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
