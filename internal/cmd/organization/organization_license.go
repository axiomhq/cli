package organization

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type licenseOptions struct {
	*cmdutil.Factory

	// ID of the organization to fetch license of. If not supplied as an
	// argument, which is optional, the user will be asked for it.
	ID string
	// Format to output data in. Defaults to tabular output.
	Format string
}

func newLicenseCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &licenseOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "license [<organization-id>] [(-f|--format=)json|table]",
		Short: "Get license of an organization",

		Args:              cmdutil.PopulateFromArgs(f, &opts.ID),
		ValidArgsFunction: cmdutil.OrganizationCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively get the license of an organization:
			$ axiom organization license
			
			# Get the license of an organization and provide the organization id
			# as an argument:
			$ axiom organization license my-org-123
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeLicense(cmd.Context(), opts); err != nil {
				return err
			}
			return runLicense(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func completeLicense(ctx context.Context, opts *licenseOptions) error {
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
	}

	if len(organizationIDs) == 1 {
		opts.ID = organizationIDs[0]
		return nil
	}

	return survey.AskOne(&survey.Select{
		Message: "Which organization to get the license for?",
		Options: organizationIDs,
	}, &opts.ID, opts.IO.SurveyIO())
}

func runLicense(ctx context.Context, opts *licenseOptions) error {
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

	license := organization.License

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing license of organization %s:\n\n", cs.Bold(organization.Name))

			trb.AddField("Valid from", cs.Bold)
			trb.AddField("Expires at", cs.Bold)
			trb.AddField("Max users", cs.Bold)
			trb.AddField("Max queries/sec", cs.Bold)
			trb.AddField("Max query window", cs.Bold)
			trb.AddField("Max audit window", cs.Bold)
			trb.AddField("Auth modes", cs.Bold)
			trb.AddField("RBAC", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, _ int) {
		trb.AddField(license.ValidFrom.Format(time.RFC1123), cs.Gray)
		trb.AddField(license.ExpiresAt.Format(time.RFC1123), cs.Gray)
		trb.AddField(strconv.Itoa(license.MaxQueriesPerSecond), nil)
		trb.AddField(strconv.Itoa(license.MaxUsers), nil)
		trb.AddField(license.MaxQueryWindow.String(), nil)
		trb.AddField(license.MaxAuditWindow.String(), nil)
		trb.AddField(strings.Join(license.WithAuths, ", "), nil)
		trb.AddField(boolToStr(cs, license.WithRBAC), nil)
	}

	return iofmt.FormatToTable(opts.IO, 1, header, nil, contentRow)
}
