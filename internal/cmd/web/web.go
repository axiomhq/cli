package web

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the web command.
func NewCmd(_ *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "web",
		Short: "Open Axiom in the browser",
		Long:  `Open Axiom in the systems default web browser.`,

		RunE: func(_ *cobra.Command, _ []string) error {
			return browser.OpenURL(client.AppURL)
		},
	}
}
