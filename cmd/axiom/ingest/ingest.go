package ingest

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode"

	axiomdb "axicode.axiom.co/watchmakers/axiomdb/client"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/utils"
)

type options struct {
	*cmdutil.Factory

	// Dataset to query.
	Dataset string
	// TimestampField to take the ingestion time from.
	TimestampField string
	// TimestampFormat the timestamp is formatted with.
	TimestampFormat string
}

// NewIngestCmd creates and returns the ingest command.
func NewIngestCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "ingest <dataset-name> [--timestamp-field <timestamp-field>] [--timestamp-format <timestamp-format>]",
		Short: "Ingest data",
		Long:  `Ingest data into an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Pipe the contents of a JSON logfile into a dataset named
			# "nginx-logs":
			$ echo nginx-logs.json | axiom ingest nginx-logs
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.NeedsActiveBackend(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Dataset = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.TimestampField, "timestamp-field", "_time", "Field to take the ingestion time from")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format the timestamp is formatted with")

	return cmd
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	var (
		r   io.Reader = opts.Factory.IO.In()
		br            = bufio.NewReader(r)
		typ axiomdb.ContentType
	)
	for {
		var c rune
		if c, _, err = br.ReadRune(); err == io.EOF {
			return errors.New("couldn't find beginning of valid JSON")
		} else if err != nil {
			return err
		} else if c == '[' {
			typ = axiomdb.JSON
		} else if c == '{' {
			typ = axiomdb.NDJSON
		} else if unicode.IsSpace(c) {
			continue
		} else {
			return errors.New("cannot determine content type")
		}

		if err = br.UnreadRune(); err != nil {
			return err
		}
		break
	}

	bufSize := br.Buffered()
	buf, err := br.Peek(bufSize)
	if err != nil {
		return err
	}
	alreadyRead := bytes.NewReader(buf)

	r = io.MultiReader(alreadyRead, r)

	res, err := client.Datasets.Ingest(ctx, opts.Dataset, gzipStream(opts.IO.ErrOut(), r), typ, axiomdb.GZIP, axiomdb.IngestOptions{
		TimestampField:  opts.TimestampField,
		TimestampFormat: opts.TimestampFormat,
	})
	if err != nil {
		return fmt.Errorf("could not ingest into dataset %q: %w", opts.Dataset, err)
	}

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut(), "%s Ingested %s\n",
			cs.SuccessIcon(),
			utils.Pluralize(cs, "event", int(res.Ingested)),
		)

		if res.Failed > 0 {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Failed to ingest %s:\n\n",
				cs.ErrorIcon(),
				utils.Pluralize(cs, "event", int(res.Failed)),
			)
			for _, fail := range res.Failures {
				fmt.Fprintf(opts.IO.ErrOut(), "%s: %s\n",
					cs.Gray(fail.Timestamp.Format(time.RFC1123)), err,
				)
			}
		}
	}

	return nil
}

func gzipStream(wErr io.Writer, r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func(r io.Reader) {
		defer pw.Close()

		// Does not fail when using a predefined compression level.
		gzw, _ := gzip.NewWriterLevel(pw, gzip.BestSpeed)
		defer gzw.Close()

		if _, err := io.Copy(gzw, r); err != nil {
			fmt.Fprintf(wErr, "error compressing data to ingest: %s\n", err)
		}
	}(r)

	return pr
}
