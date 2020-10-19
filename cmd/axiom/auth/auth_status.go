package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
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

		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: backendCompletionFunc(f.Config),

		Example: heredoc.Doc(`
			# Check authentication status of all configured backends:
			$ axiom auth status
			
			# Check authentication status of a specified backend:
			$ axiom auth status my-axiom
		`),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.Alias = args[0]
			}

			return cmdutil.ChainRunFuncs(
				cmdutil.NeedsBackends(f),
				cmdutil.NeedsValidBackend(f, opts.Alias),
			)(cmd, args)
		},

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

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	// TODO: Get authentication status, I guess we need ctx in here soon ;)
	_ = ctx

	time.Sleep(time.Second * 2)

	stop()

	cs := opts.IO.ColorScheme()

	var (
		failed     bool // TODO: Set to true if fetching a backend fails.
		statusInfo = map[string][]string{}
	)
	for _, v := range backendAliases {
		backend, ok := opts.Config.Backends[v]
		if !ok {
			continue
		}

		loginInfo := fmt.Sprintf("%s Logged in to %s as %s",
			cs.SuccessIcon(), backend.URL, cs.Bold(backend.Username))
		expireInfo := fmt.Sprintf("%s Your token is about to expire! Run %s %s to refresh",
			cs.WarningIcon(), cs.Bold("axiom auth refresh"), cs.Bold(v))

		statusInfo[v] = append(statusInfo[v], loginInfo)
		statusInfo[v] = append(statusInfo[v], expireInfo)
	}

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
