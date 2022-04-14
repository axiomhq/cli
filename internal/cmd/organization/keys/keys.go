package keys

import (
	"context"
	"sort"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

// NewCmd creates and returns the keys command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys <command>",
		Short: "Manage organization shared access keys",
		Long:  "Manage organization shared access keys.",

		Example: heredoc.Doc(`
			$ axiom organization keys get
			$ axiom organization keys rotate my-org-123
			$ axiom organization keys get my-org-123
		`),

		Annotations: map[string]string{
			"IsManagement": "true",
		},

		PersistentPreRunE: cmdutil.NeedsCloudDeployment(f),
	}

	cmd.AddCommand(newGetCmd(f))
	cmd.AddCommand(newRotateCmd(f))

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
	sort.Strings(organizationIDs)

	return organizationIDs, nil
}
