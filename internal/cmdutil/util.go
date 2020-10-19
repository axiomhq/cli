package cmdutil

import "github.com/spf13/cobra"

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
