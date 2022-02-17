package organization

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
	"github.com/axiomhq/cli/pkg/utils"
)

type listOptions struct {
	*cmdutil.Factory

	// Format to output data in. Defaults to tabular output.
	Format string
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list [(-f|--format=)json|table]",
		Short: "List all organizations",

		Aliases: []string{"ls"},

		Example: heredoc.Doc(`
			# List all organizations:
			$ axiom organization list
		`),

		PreRunE: cmdutil.NeedsCloudDeployment(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", cmdutil.FormatCompletion)

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	progStop := opts.IO.StartActivityIndicator()
	defer progStop()

	organizations, err := client.Organizations.Cloud.List(ctx)
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
		return iofmt.FormatToJSON(opts.IO.Out(), organizations, opts.IO.ColorEnabled())
	}

	cs := opts.IO.ColorScheme()

	var header iofmt.HeaderBuilderFunc
	if opts.IO.IsStdoutTTY() {
		header = func(w io.Writer, trb iofmt.TableRowBuilder) {
			fmt.Fprintf(opts.IO.Out(), "Showing %s:\n\n", utils.Pluralize(cs, "organization", len(organizations)))
			trb.AddField("ID", cs.Bold)
			trb.AddField("Name", cs.Bold)
			trb.AddField("Plan", cs.Bold)
			trb.AddField("Plan created", cs.Bold)
			trb.AddField("Plan expires", cs.Bold)
			trb.AddField("Trialed", cs.Bold)
		}
	}

	contentRow := func(trb iofmt.TableRowBuilder, k int) {
		organization := organizations[k]

		trb.AddField(organization.ID, nil)
		trb.AddField(organization.Name, nil)
		trb.AddField(strings.Title(organization.Plan.String()), nil)
		trb.AddField(organization.PlanCreated.Format(time.RFC1123), cs.Gray)
		trb.AddField(organization.PlanExpires.Format(time.RFC1123), cs.Gray)
		trb.AddField(boolToStrReverseColors(cs, organization.Trialed), nil)
	}

	return iofmt.FormatToTable(opts.IO, len(organizations), header, nil, contentRow)
}
