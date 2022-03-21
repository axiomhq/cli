package organization

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type infoOptions struct {
	*cmdutil.Factory

	// ID of the organization to fetch info of. If not supplied as an argument,
	// which is optional, the user will be asked for it.
	ID string
	// Format to output data in. Defaults to tabular output.
	Format string
}

func newInfoCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &infoOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "info [<organization-id>] [(-f|--format=)json|table]",
		Short: "Get info about an organization",

		Args:              cmdutil.PopulateFromArgs(f, &opts.ID),
		ValidArgsFunction: cmdutil.OrganizationCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively get info of an organization:
			$ axiom organization info
			
			# Get info of an organization and provide the organization id as an
			# argument:
			$ axiom organization info my-org-123
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeInfo(cmd.Context(), opts); err != nil {
				return err
			}
			return runInfo(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func completeInfo(ctx context.Context, opts *infoOptions) error {
	// A requirement for this command to execute is the presence of an active
	// deployment, so no need to check for existence.
	activeDeployment, _ := opts.Config.GetActiveDeployment()
	if !client.IsCloudURL(activeDeployment.URL) && !opts.Config.ForceCloud {
		opts.ID = defaultSelfhostOrganizationID
	}

	if opts.ID != "" {
		return nil
	}

	organizationIDs, err := getOrganizationIDs(ctx, opts.Factory)
	if err != nil {
		return err
	} else if len(organizationIDs) == 1 {
		opts.ID = organizationIDs[0]
		return nil
	}

	return survey.AskOne(&survey.Select{
		Message: "Which organization to get info for?",
		Options: organizationIDs,
	}, &opts.ID, opts.IO.SurveyIO())
}

func runInfo(ctx context.Context, opts *infoOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	organization, err := client.Organizations.Selfhost.Get(ctx, opts.ID)
	if err != nil {
		return err
	}

	progStop()

	pagerStop, err := opts.IO.StartPager(ctx)
	if err != nil {
		return err
	}
	defer pagerStop()

	if opts.Format == iofmt.JSON.String() {
		return iofmt.FormatToJSON(opts.IO.Out(), organization, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing info of organization %s:\n\n", cs.Bold(organization.Name))
			trb.AddField("ID", cs.Bold)
			trb.AddField("Plan", cs.Bold)
			trb.AddField("Plan created", cs.Bold)
			trb.AddField("Plan expires", cs.Bold)
			trb.AddField("Trialed", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, _ int) {
		trb.AddField(organization.ID, nil)
		trb.AddField(organization.Plan.String(), nil)
		trb.AddField(organization.PlanCreated.Format(time.RFC1123), cs.Gray)
		trb.AddField(organization.PlanExpires.Format(time.RFC1123), cs.Gray)
		trb.AddField(boolToStrReverseColors(cs, organization.Trialed), nil)
	}

	return iofmt.FormatToTable(opts.IO, 1, header, nil, contentRow)
}
