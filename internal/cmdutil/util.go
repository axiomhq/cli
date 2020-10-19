package cmdutil

import "github.com/spf13/cobra"

// DefaultCompletion sets default values for Args and ValidArgsFunction on all
// child commands. If Args is nil it is set to cobra.NoArgs, if
// ValidArgsFunction is nil and no ValidArgs are given, it is set to
// NoCompletion.
func DefaultCompletion(cmd *cobra.Command) {
	// Make having no arguments the default if nothing else is specified but
	// skip settings this on the root command because it breaks returning errors
	// when giving bad arguments.
	if cmd.Args == nil && cmd.HasParent() {
		cmd.Args = cobra.NoArgs
	}

	// If no ValidArgs are specified by the appropriate struct field or
	// function, set the ValidArgsFunction to NoCompletion.
	if len(cmd.ValidArgs) == 0 && cmd.ValidArgsFunction == nil {
		cmd.ValidArgsFunction = NoCompletion
	}

	for _, c := range cmd.Commands() {
		DefaultCompletion(c)
	}
}

// InheritRootPersistenPreRun inherits the root commands PersistenPreRunE
// function down to child commands which specify their own one by chaining them.
func InheritRootPersistenPreRun(cmd *cobra.Command) {
	// Skip chaining on the root command.
	if cmd.HasParent() && cmd.PersistentPreRunE != nil {
		f := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = ChainRunFuncs(NeedsRootPersistentPreRunE(), f)
	}

	for _, c := range cmd.Commands() {
		InheritRootPersistenPreRun(c)
	}
}
