package organization

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

// NewOrganizationCmd creates and returns the organization command.
func NewOrganizationCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organization <command>",
		Short: "Manage organizations",
		Long:  "Manage organizations.",

		Example: heredoc.Doc(`
			$ axiom organization list
			$ axiom organization license my-org-123
			$ axiom organization info my-org-123
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
		),
	}

	cmd.AddCommand(newInfoCmd(f))
	cmd.AddCommand(newLicenseCmd(f))
	cmd.AddCommand(newListCmd(f))

	return cmd
}

func getOrganizationIDs(ctx context.Context, f *cmdutil.Factory) ([]string, error) {
	client, err := f.Client(ctx)
	if err != nil {
		return nil, err
	}

	stop := f.IO.StartActivityIndicator()
	defer stop()

	organizations, err := client.Organizations.Cloud.List(ctx)
	if err != nil {
		return nil, err
	}

	stop()

	organizationIDs := make([]string, len(organizations))
	for i, organization := range organizations {
		organizationIDs[i] = organization.ID
	}

	return organizationIDs, nil
}

func boolToStr(cs *terminal.ColorScheme, b bool) string {
	if b {
		return cs.SuccessIcon()
	}
	return cs.ErrorIcon()
}

func boolToStrReverseColors(cs *terminal.ColorScheme, b bool) string {
	if b {
		return cs.Red("✓")
	}
	return cs.Green("✖")
}
