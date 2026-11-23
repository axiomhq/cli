package org

import (
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the org command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "org <command>",
		Short:       "Manage organizations",
		Annotations: map[string]string{"IsManagement": "true"},
	}
	cmd.AddCommand(newCreateCmd(f))
	return cmd
}
