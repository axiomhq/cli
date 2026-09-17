package cmdutil

import (
	"slices"

	"github.com/spf13/cobra"
)

// PopulateFromArgs populates the given values with the argumetns given on the
// command-line. It returns an error if the application is not running
// interactively and has not enough arguments to populate all the given values,
// or if one of them is empty, because commands prompt for empty values.
// It applies cobra.MaximumNArgs() with n being the amount of values to
// populate.
func PopulateFromArgs(f *Factory, ss ...*string) cobra.PositionalArgs {
	populate := func(_ *cobra.Command, args []string) error {
		if !f.IO.IsInteractive() && (len(args) < len(ss) || slices.Contains(args, "")) {
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

	return cobra.MatchAll(
		cobra.MaximumNArgs(len(ss)),
		populate,
	)
}
