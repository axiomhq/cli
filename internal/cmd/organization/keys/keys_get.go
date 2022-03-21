package keys

import (
	"context"
	"fmt"
	"io"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type getOptions struct {
	*cmdutil.Factory

	// ID of the organization to fetch keys of. If not supplied as an argument,
	// which is optional, the user will be asked for it.
	ID string
	// Format to output data in. Defaults to tabular output.
	Format string
}

func newGetCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &getOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "get [<organization-id>] [(-f|--format)=json|table]",
		Short: "Get shared access keys of an organization",

		Args:              cmdutil.PopulateFromArgs(f, &opts.ID),
		ValidArgsFunction: cmdutil.OrganizationCompletionFunc(f),

		Example: heredoc.Doc(`
			# Interactively get the shared access keys of an organization:
			$ axiom organization keys get
			
			# Get the shared access keys of an organization and provide the
			# organization id as an argument:
			$ axiom organization keys get my-org-123
		`),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := completeGet(cmd.Context(), opts); err != nil {
				return err
			}
			return runGet(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func completeGet(ctx context.Context, opts *getOptions) error {
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

func runGet(ctx context.Context, opts *getOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	organization, err := client.Organizations.Cloud.Get(ctx, opts.ID)
	if err != nil {
		return err
	}

	keys, err := client.Organizations.Cloud.ViewSharedAccessKeys(ctx, opts.ID)
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
		return iofmt.FormatToJSON(opts.IO.Out(), keys, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing shared access keys of organization %s:\n\n", cs.Bold(organization.Name))

			trb.AddField("ID", cs.Bold)
			trb.AddField("Key", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		if k == 0 {
			trb.AddField("Primary", cs.Gray)
			trb.AddField(keys.Primary, nil)
		} else if k == 1 {
			trb.AddField("Secondary", cs.Gray)
			trb.AddField(keys.Secondary, nil)
		}
	}

	return iofmt.FormatToTable(opts.IO, 2, header, nil, contentRow)
}
