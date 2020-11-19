package cmdutil

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// A CompletionFunc is dynamically invoked without running any of the Run()
// methods on the cobra.Command. Keep this in mind when implementing the, since
// not all factories are initialized.
type CompletionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// NoCompletion disables completion.
func NoCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// DatasetCompletionFunc returns a completion function which completes the
// datasets from the configured active deployment.
func DatasetCompletionFunc(f *Factory) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Just complete the first argument.
		if len(args) > 0 {
			return NoCompletion(cmd, args, toComplete)
		}

		// FIXME(lukasmalkmus): Get rid of this fix which makes sure we never
		// pass a nil context.
		ctx := cmd.Context()
		if ctx == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
		}

		client, err := f.Client()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		datasets, err := client.Datasets.List(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		res := make([]string, 0, len(datasets))
		for _, dataset := range datasets {
			if strings.HasPrefix(dataset.Name, toComplete) {
				res = append(res, dataset.Name)
			}
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	}
}
