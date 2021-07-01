package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/nwidger/jsoncolor"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/iofmt"
)

type options struct {
	*cmdutil.Factory

	// Query to run. If not supplied as an argument, which is optional, the user
	// will be asked for it.
	Query string
	// Format to output data in. Defaults to tabular output.
	Format string
}

// NewQueryCmd creates and returns the query command.
func NewQueryCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "query [<apl-query>] [(-f|--format=)json|table]",
		Short: "Query data using APL",
		Long: heredoc.Doc(`
			Query data from an Axiom dataset using APL, the Axiom Processing
			Language.
		`),

		DisableFlagsInUseLine: true,

		Args: cmdutil.PopulateFromArgs(f, &opts.Query),

		Example: heredoc.Doc(`
			# Query the "nginx-logs" dataset for logs with a 304 status code:
			$ axiom query "nginx-logs | where response == 304"
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
			cmdutil.NeedsDatasets(f),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := complete(opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", iofmt.Table.String(), "Format to output data in")

	_ = cmd.RegisterFlagCompletionFunc("format", formatCompletion)

	return cmd
}

func complete(opts *options) error {
	if opts.Query != "" {
		return nil
	}

	return survey.AskOne(&survey.Input{
		Message: "Which query to run?",
	}, &opts.Query, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	cs := opts.IO.ColorScheme()

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.Out(), "Result of query %s:\n\n", cs.Bold(opts.Query))
	}

	var enc interface {
		Encode(interface{}) error
	}
	if opts.IO.ColorEnabled() {
		enc = jsoncolor.NewEncoder(opts.IO.Out())
	} else {
		enc = json.NewEncoder(opts.IO.Out())
	}

	res, err := client.Datasets.APLQuery(ctx, opts.Query, query.Options{})
	if err != nil {
		return err
	} else if res == nil || len(res.Matches) == 0 {
		return errors.New("query returned no results")
	}

	for _, entry := range res.Matches {
		switch opts.Format {
		case iofmt.JSON.String():
			_ = enc.Encode(entry)
		default:
			fmt.Fprintf(opts.IO.Out(), "%s\t",
				cs.Gray(entry.Time.Format(time.RFC1123)))
			_ = enc.Encode(entry.Data)
		}
		fmt.Fprintln(opts.IO.Out())
	}

	return nil
}

func formatCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(iofmt.Formats()))
	for _, validFormat := range iofmt.Formats() {
		if strings.HasPrefix(validFormat.String(), toComplete) {
			res = append(res, validFormat.String())
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}
