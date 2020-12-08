package cmdutil

import (
	"github.com/spf13/cobra"
)

// ChainPositionalArgs chains one or more cobra.PositionalArgs functions.
func ChainPositionalArgs(fns ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range fns {
			if err := fn(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// PopulateFromArgs populates the given values with the argumetns given on the
// command-line. It returns an error if the application is not running
// interactively and has not enough arguments to populate all the given values.
// It applies cobra.MaximumNArgs() with n being the amount of values to
// populate.
func PopulateFromArgs(f *Factory, ss ...*string) cobra.PositionalArgs {
	populate := func(cmd *cobra.Command, args []string) error {
		if !f.IO.IsStdinTTY() && len(args) < len(ss) {
			return ErrNoPromptArgRequired
		}

		for k := range args {
			if k >= len(ss) {
				return nil
			}
			*ss[k] = args[k]
		}

		return nil
	}

	return ChainPositionalArgs(
		cobra.MaximumNArgs(len(ss)),
		populate,
	)
}
